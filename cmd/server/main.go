package main

import (
	"context"
	"fmt"
	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/config/db"
	"github.com/annakonkova23/collect-metrics/internal/handler"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	options := config.NewServerOptions()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		fmt.Println("Получен сигнал завершения. Сохраняю данные...")
		cancel()
	}()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
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
		panic(err)
	}
	h, err := handler.NewServer(ctx, options, logger, collector)
	if err != nil {
		panic(err)
	}
	logger.Info("Опции", zap.String("options", fmt.Sprintf("%+v", options)))
	logger.Info("Сервер создан",
		zap.String("host", "localhost:8080"),
	)
	go func() {
		err = h.StartAndListen()
		if err != nil {
			logger.Error(err.Error())
		}
	}()
	<-ctx.Done()
	collector.GetDataAndSaveToFile()
}

///github.com/annakonkova23/collect-metrics
