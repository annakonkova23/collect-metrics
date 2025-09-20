package main

import (
	"github.com/annakonkova23/collect-metrics/internal/handler"
)

func main() {

	h := handler.NewServer("localhost:8080")
	err := h.StartAndListen()
	if err != nil {
		panic(err)
	}
}

///github.com/annakonkova23/collect-metrics
