package main

import (
	"flag"
	"log"

	"github.com/annakonkova23/collect-metrics/internal/service"
)

const (
	defaultReportInterval = 10
	defaultHost           = "localhost:8080"
	defaultPollInterval   = 2
)

func main() {
	var host string
	var pollInterval int
	var reportInterval int
	flag.StringVar(&host, "a", defaultHost, "Хост")
	flag.IntVar(&pollInterval, "p", defaultPollInterval, "Частота опроса метрик из пакета runtime (в секундах)")
	flag.IntVar(&reportInterval, "r", defaultReportInterval, "Частота отправки метрик на сервер (в секундах)")
	flag.Parse()
	URL := ""
	if host != "" {
		URL = "http://" + host + "/update/"
		log.Println("URL:", URL)
	} else {
		panic("Не указан адрес")
	}
	if pollInterval <= 0 {
		panic("Неверно указана частота опроса")
	}
	log.Println("PollInterval:", pollInterval)
	if reportInterval <= 0 {
		panic("Неверно указана частота отправки")
	}
	log.Println("ReportInterval:", reportInterval)
	sender := service.NewSender(URL, pollInterval, reportInterval)
	log.Println("Отправитель создан")
	sender.Start()

}

///github.com/annakonkova23/collect-metrics
