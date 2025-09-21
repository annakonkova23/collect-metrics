package service

import (
	"fmt"
	"github.com/annakonkova23/collect-metrics/internal/model"
	"log"
	"strings"
)

const (
	ErrorNotFound = "Not exists name metric"
)

type Collector struct {
	MemStorage *model.MemStorage
}

func NewCollector() *Collector {
	return &Collector{MemStorage: model.NewMemStorage()}
}

func (c *Collector) ParseAndSaveMetricsByURL(url string) error {
	var typeMetric, nameMetric, value string
	parts := strings.Split(url, "/")
	if len(parts) < 2 || parts[1] != "update" {
		return fmt.Errorf("%s", "Невалидный Url")
	}
	if len(parts) == 3 || (len(parts) > 3 && parts[3] == "") {
		log.Println("Запрос без именования метрики")
		return fmt.Errorf("%s", ErrorNotFound)
	}
	if len(parts) >= 4 {
		typeMetric = parts[2]
		nameMetric = parts[3]
		if len(parts) >= 5 {
			value = parts[4]
		}
	}
	err := c.MemStorage.SetMetric(nameMetric, typeMetric, value)
	if err != nil {
		return err
	}
	return nil
}
