package main

import (
	"log"

	gk "github.com/annakonkova23/collect-metrics/internal/keygen"
)

func main() {
	err := gk.Generate("")
	if err != nil {
		log.Fatal(err)
	}
}
