package main

import (
	"flag"
	"github.com/annakonkova23/collect-metrics/internal/handler"
	"log"
)

const (
	defaultHost = "localhost:8080"
)

func main() {
	host := flag.String("a", defaultHost, "Хост")
	flag.Parse()
	if host == nil {
		panic("Не указан адрес")
	}
	h := handler.NewServer(*host)
	log.Println("Сервер создан")
	err := h.StartAndListen()
	if err != nil {
		panic(err)
	}
}

///github.com/annakonkova23/collect-metrics
