package main

import (
	"context"
	"errors"

	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewAgentOptions()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	URL := ""
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	var errs []error
	if cfg.Host == "" {
		msg := "не указан адрес"
		errs = append(errs, errors.New(msg))
	}
	if cfg.PollInterval <= 0 {
		msg := "неверно указана частота опроса"
		errs = append(errs, errors.New(msg))
	}
	if cfg.ReportInterval <= 0 {
		msg := "неверно указана частота отправки"
		errs = append(errs, errors.New(msg))
	}
	if len(errs) > 0 {
		panic(errors.Join(errs...))
	}

	URL = "http://" + cfg.Host + "/updates/"
	logger.Info("Параметры", zap.String("URL", URL), zap.Int("PollInterval", cfg.PollInterval), zap.Int("ReportInterval", cfg.ReportInterval))
	sender := service.NewSender(URL, cfg.PollInterval, cfg.ReportInterval, cfg.Key, logger)
	logger.Info("Отправитель создан")
	sender.Start(ctx)

}
