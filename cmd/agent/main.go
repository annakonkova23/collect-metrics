package main

import (
	"github.com/annakonkova23/collect-metrics/internal/service"
	"log"
)

func main() {

	sender := service.NewSender("http://localhost:8080/update/")
	log.Println("Отправитель создан")
	sender.Start()

}

///github.com/annakonkova23/collect-metrics
