package service

import (
	"github.com/annakonkova23/collect-metrics/internal/model"
	"github.com/annakonkova23/collect-metrics/internal/service/client"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	reportInterval = 10
)

type Sender struct {
	client    *client.Client
	runMetric *model.RuntimeMetric
	url       string
}

func NewSender(url string) *Sender {
	return &Sender{
		client:    client.NewClient(),
		runMetric: model.NewRuntimeMetric(),
		url:       url,
	}
}
func (s *Sender) SendRequest(b chan bool) {
	log.Println("Ждём расчета метрик")
	<-b
	for {
		metrics := s.runMetric.GetMetrics()
		var wg sync.WaitGroup
		for name, value := range metrics {
			wg.Add(1)
			go func() {
				valueStr := strconv.FormatFloat(value, 'f', -1, 64)
				log.Printf("Запрос %s %s", name, valueStr)
				if err := s.client.Post(s.GetUrlForMetric(name, s.runMetric.GetTypeMetric(name), valueStr)); err != nil {
					log.Printf("Error sending metric: %v", err)
				} else {
					log.Printf("Успешный ответ %s %s", name, valueStr)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		time.Sleep(time.Duration(reportInterval) * time.Second)

	}
}

func (s *Sender) GetUrlForMetric(name, typeM, value string) string {
	return s.url + typeM + "/" + name + "/" + value
}

func (s *Sender) Start() {
	b := make(chan bool)
	log.Println("Старт отправления метрик")
	go s.runMetric.UpdateMetric(b)
	s.SendRequest(b)
}
