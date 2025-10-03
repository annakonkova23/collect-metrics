package service

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/agent"
	"github.com/annakonkova23/collect-metrics/internal/model"
)

type Sender struct {
	client         *agent.Client
	runMetric      *model.RuntimeMetric
	url            string
	reportInterval int
}

func NewSender(url string, pollInterval, reportInterval int) *Sender {
	return &Sender{
		client:         agent.NewClient(),
		runMetric:      model.NewRuntimeMetric(pollInterval),
		url:            url,
		reportInterval: reportInterval,
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
				if err := s.client.Post(s.GetURLForMetric(name, s.runMetric.GetTypeMetric(name), valueStr)); err != nil {
					log.Printf("Ошибка отправки метрики %s: %v", name, err)
				} else {
					log.Printf("Успешный ответ %s %s", name, valueStr)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		time.Sleep(time.Duration(s.reportInterval) * time.Second)

	}
}

func (s *Sender) GetURLForMetric(name, typeM, value string) string {
	return s.url + typeM + "/" + name + "/" + value
}

func (s *Sender) Start() {
	b := make(chan bool)
	log.Println("Старт отправления метрик")
	go s.runMetric.UpdateMetric(b)
	s.SendRequest(b)
}
