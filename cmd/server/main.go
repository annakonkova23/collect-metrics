package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/config/db"
	handlerServer "github.com/annakonkova23/collect-metrics/internal/handler"
	"github.com/annakonkova23/collect-metrics/internal/service"
	mcs "github.com/annakonkova23/collect-metrics/pkg/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	config.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	options := config.NewServerOptions()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGQUIT)
	defer signal.Stop(quitCh)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	dbConnect, err := db.NewDBConnect(options.DatabaseDSN)
	if err != nil {
		logger.Error("Не удалось подключиться к БД", zap.Error(err))
	} else {
		defer dbConnect.Close()
	}

	collector, err := service.NewCollector(ctx, options, logger, dbConnect)
	if err != nil {
		log.Fatal(err)
	}

	h, err := handlerServer.NewServer(ctx, options, logger, collector)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Опции", zap.String("options", fmt.Sprintf("%+v", options)))
	logger.Info("Сервер создан",
		zap.String("host", options.Host),
	)

	errCh := make(chan error, 1)

	go func() {
		errCh <- h.StartAndListen(ctx)
	}()

	if options.GrpcHost != "" {
		go func() {
			errCh <- StartGrpcServer(logger, options.GrpcHost, options.TrustedSubnet, h)
		}()
	}

	go func() {
		log.Println("pprof: http://localhost:6061/debug/pprof/")
		if err := http.ListenAndServe("localhost:6061", nil); err != nil {
			errCh <- fmt.Errorf("pprof server failed: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		h.Shutdown(shutdownCtx)
		logger.Info("Получен сигнал завершения. Сохраняю данные...")
		collector.GetDataAndSaveToFile()
		return

	case <-quitCh:
		fmt.Fprintln(os.Stderr, "Получен сигнал SIGQUIT: goroutine dump (pprof)")
		if p := pprof.Lookup("goroutine"); p != nil {
			_ = p.WriteTo(os.Stderr, 2)
		} else {
			fmt.Fprintln(os.Stderr, "pprof.Lookup(\"goroutine\") вернул nil")
		}
		return

	case err := <-errCh:
		if err != nil {
			log.Printf("Критическая ошибка в горутине: %v", err)
			log.Fatal(err)
		}
	}
}

func StartGrpcServer(lgr *zap.Logger, host, cidr string, srv *handlerServer.Server) error {
	listen, err := net.Listen("tcp", host)
	if err != nil {
		return fmt.Errorf("ошибка создания слушателя: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(handlerServer.InterceptorCheckIsIPInCIDR(cidr)))

	mcs.RegisterMetricsServer(s, srv)

	lgr.Info("сервер gRPC начал работу", zap.String("host", host))

	if err := s.Serve(listen); err != nil {
		return fmt.Errorf("ошибка запуска grpc сервера: %v", err)
	}
	return nil
}
