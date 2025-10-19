package main

import (
	"fmt"

	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/handler"
	"go.uber.org/zap"
)

func main() {
	options := config.NewServerOptions()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	h, err := handler.NewServer(options, logger)
	if err != nil {
		panic(err)
	}
	logger.Info("Опции", zap.String("options", fmt.Sprintf("%+v", options)))
	logger.Info("Сервер создан",
		zap.String("host", "localhost:8080"),
	)
	err = h.StartAndListen()
	if err != nil {
		panic(err)
	}
}

///github.com/annakonkova23/collect-metrics
