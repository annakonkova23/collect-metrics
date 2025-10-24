package main

import (
	"flag"
	"fmt"
	"github.com/annakonkova23/collect-metrics/internal/handler"
	"go.uber.org/zap"
	"os"
)

const (
	defaultHost = "localhost:8080"
)

func main() {
	host := flag.String("a", defaultHost, "Хост")
	flag.Parse()
	fmt.Println(os.Getenv("ADDRESS"))
	if envHost := os.Getenv("ADDRESS"); envHost != "" {
		host = &envHost
	}
	if host == nil {
		panic("Не указан адрес")
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	h := handler.NewServer(*host, logger)
	logger.Info("Сервер создан",
		zap.String("host", *host),
	)
	err = h.StartAndListen()
	if err != nil {
		panic(err)
	}
}

///github.com/annakonkova23/collect-metrics
