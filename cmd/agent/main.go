package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/annakonkova23/collect-metrics/internal/service"
	"go.uber.org/zap"
)

const (
	defaultReportInterval = 5
	defaultHost           = "localhost:8080"
	defaultPollInterval   = 2
	env
)

func main() {
	var host, flagHost, envHost, envPollInterval, envReportInterval string
	var pollInterval, flagPollInterval int
	var reportInterval, flagReportInterval int
	flag.StringVar(&flagHost, "a", defaultHost, "Хост")
	flag.IntVar(&flagPollInterval, "p", defaultPollInterval, "Частота опроса метрик из пакета runtime (в секундах)")
	flag.IntVar(&flagReportInterval, "r", defaultReportInterval, "Частота отправки метрик на сервер (в секундах)")
	flag.Parse()

	if envHost = os.Getenv("ADDRESS"); envHost != "" {
		host = envHost
	} else {
		host = flagHost
	}
	var err error
	if envReportInterval = os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		reportInterval, err = strconv.Atoi(envReportInterval)
		if err != nil {
			reportInterval = defaultReportInterval
		}
	} else {
		reportInterval = flagReportInterval
	}
	if envPollInterval = os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		pollInterval, err = strconv.Atoi(envPollInterval)
		if err != nil {
			pollInterval = defaultPollInterval
		}
	} else {
		pollInterval = flagPollInterval
	}
	URL := ""
	logger, err := zap.NewDevelopment()
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	defer logger.Sync()
	if host != "" {
		URL = "http://" + host + "/update/"
		logger.Info("Параметры",
			zap.String("URL", URL),
		)
	} else {
		logger.Error("Не указан адрес")
		panic("Не указан адрес")
	}
	if pollInterval <= 0 {
		panic("Неверно указана частота опроса")
	}
	logger.Info("Параметры",
		zap.Int("PollInterval", pollInterval),
	)
	if reportInterval <= 0 {
		panic("Неверно указана частота отправки")
	}
	logger.Info("Параметры",
		zap.Int("ReportInterval", reportInterval))
	sender := service.NewSender(URL, pollInterval, reportInterval, logger)
	logger.Info("Отправитель создан")
	sender.Start()

}
