package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/config/db"
	"github.com/annakonkova23/collect-metrics/internal/handler"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"go.uber.org/zap"
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

	h, err := handler.NewServer(ctx, options, logger, collector)
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

	go func() {
		log.Println("pprof: http://localhost:6061/debug/pprof/")
		if err := http.ListenAndServe("localhost:6061", nil); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h.Shutdown(shutdownCtx)
	fmt.Println("Получен сигнал завершения. Сохраняю данные...")
	logger.Info("Контекст завершён, сохраняю данные", zap.Error(ctx.Err()))

	collector.GetDataAndSaveToFile()
}

///github.com/annakonkova23/collect-metrics
