package main

import (
	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"go.uber.org/zap"
)

const (
	defaultReportInterval = 5
	defaultHost           = "localhost:8080"
	defaultPollInterval   = 2
)

func main() {
	cfg := config.NewAgentOptions()
	URL := ""
	logger, err := zap.NewDevelopment()
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	defer logger.Sync()
	if cfg.Host != "" {
		URL = "http://" + cfg.Host + "/updates/"
		logger.Info("Параметры",
			zap.String("URL", URL),
		)
	} else {
		logger.Error("Не указан адрес")
		panic("Не указан адрес")
	}
	if cfg.PollInterval <= 0 {
		panic("Неверно указана частота опроса")
	}
	logger.Info("Параметры",
		zap.Int("PollInterval", cfg.PollInterval),
	)
	if cfg.ReportInterval <= 0 {
		panic("Неверно указана частота отправки")
	}
	logger.Info("Параметры",
		zap.Int("ReportInterval", cfg.ReportInterval))
	sender := service.NewSender(URL, cfg.PollInterval, cfg.ReportInterval, logger)
	logger.Info("Отправитель создан")
	sender.Start()

}
