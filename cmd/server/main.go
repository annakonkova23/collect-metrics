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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

	go func() {
		if err := h.StartAndListen(ctx); err != nil {
			logger.Error("Сервер завершился с ошибкой", zap.Error(err))
		}
	}()

	if options.GrpcHost != "" {
		go func() {
			if err := StartGrpcServer(logger, options.GrpcHost, options.TrustedSubnet, h); err != nil {
				logger.Error("Ошибка запуска gRPC сервера", zap.Error(err))
			}
		}()
	}

	go func() {
		log.Println("pprof: http://localhost:6061/debug/pprof/")
		if err := http.ListenAndServe("localhost:6061", nil); err != nil {
			log.Fatal(err)
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		h.Shutdown(shutdownCtx)
		fmt.Println("Получен сигнал завершения. Сохраняю данные...")
		logger.Info("Контекст завершён, сохраняю данные", zap.Error(ctx.Err()))

		collector.GetDataAndSaveToFile()
	case <-quitCh:
		fmt.Fprintln(os.Stderr, "Получен сигнал SIGQUIT: goroutine dump (pprof)")
		if p := pprof.Lookup("goroutine"); p != nil {
			_ = p.WriteTo(os.Stderr, 2)
		} else {
			fmt.Fprintln(os.Stderr, "pprof.Lookup(\"goroutine\") вернул nil")
		}
	}

}

func StartGrpcServer(lgr *zap.Logger, host, cidr string, srv *handlerServer.Server) error {
	listen, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(IsIPInCIDR(cidr)))

	mcs.RegisterMetricsServer(s, srv)

	lgr.Info("сервер gRPC начал работу", zap.String("host", host))

	if err := s.Serve(listen); err != nil {
		return err
	}
	return nil
}

func IsIPInCIDR(cidr string) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var ip string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("x-real-ip")
			if len(values) > 0 {
				ip = values[0]
				if is, err := handlerServer.IsIPInCIDR(ip, cidr); err != nil {
					return nil, status.Error(codes.PermissionDenied, err.Error())
				} else {
					if !is {
						return nil, status.Errorf(codes.PermissionDenied, "IP %s не принадлежит доверенной сети %s", ip, cidr)
					}
				}
			}
		}
		return handler(ctx, req)
	}
}
