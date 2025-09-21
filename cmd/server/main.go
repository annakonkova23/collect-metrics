package main

import (
	"collect-metrics/internal/handler"
	"log"
)

func main() {

	h := handler.NewServer("localhost:8080")
	log.Println("Сервер создан")
	err := h.StartAndListen()
	if err != nil {
		panic(err)
	}
}

///github.com/annakonkova23/collect-metrics
