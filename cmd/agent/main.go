package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewAgentOptions()
	ctx, cancel := context.WithCancel(context.Background())
	URL := ""
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	err = checkCfg(cfg)
	if err != nil {
		panic(err)
	}

	URL = "http://" + cfg.Host + "/updates/"

	logger.Info("Параметры", zap.String("URL", URL),
		zap.Int("PollInterval", cfg.PollInterval),
		zap.Int("ReportInterval", cfg.ReportInterval),
		zap.Int("RateLimit", cfg.RateLimiter))

	sender := service.NewSender(URL, cfg.PollInterval, cfg.ReportInterval, cfg.Key, logger)
	logger.Info("Отправитель создан")

	sender.Start(ctx, cfg.RateLimiter)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Info("Получен сигнал. Отмена...")
	cancel()

}

func checkCfg(cfg *config.AgentOptions) error {
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
	if cfg.RateLimiter <= 0 {
		msg := "неверно указано количество исходящих запросов"
		errs = append(errs, errors.New(msg))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
