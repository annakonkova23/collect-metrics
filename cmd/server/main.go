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
		fmt.Println(err)
		panic(err)
	}
	defer logger.Sync()
	h, err := handler.NewServer(options, logger)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	logger.Info("Сервер создан",
		zap.String("host", options.Host),
	)
	err = h.StartAndListen()
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}

///github.com/annakonkova23/collect-metrics
