package service

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/agent"
	"github.com/annakonkova23/collect-metrics/internal/model"
	"go.uber.org/zap"
)

type Sender struct {
	client         *agent.Client
	runMetric      *model.RuntimeMetric
	url            string
	reportInterval int
	logger         *zap.Logger
}

func NewSender(url string, pollInterval, reportInterval int, logger *zap.Logger) *Sender {
	return &Sender{
		client:         agent.NewClient(logger.Sugar()),
		runMetric:      model.NewRuntimeMetric(pollInterval, logger),
		url:            url,
		reportInterval: reportInterval,
		logger:         logger,
	}
}
func (s *Sender) SendRequest(b chan bool) {
	s.logger.Info("Ждём расчета метрик")
	<-b
	for {
		metrics := s.runMetric.GetMetrics()
		var wg sync.WaitGroup
		for name, value := range metrics {
			wg.Add(1)
			go func() {
				valueStr := strconv.FormatFloat(value, 'f', -1, 64)
				s.logger.Info(fmt.Sprintf("Запрос %s %s", name, valueStr))
				if err := s.client.Post(s.GetURLForMetric(name, s.runMetric.GetTypeMetric(name), valueStr)); err != nil {
					s.logger.Info(fmt.Sprintf("Ошибка отправки метрики %s: %v", name, err))
				} else {
					s.logger.Info(fmt.Sprintf("Успешный ответ %s %s", name, valueStr))
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
	s.logger.Info("Старт отправления метрик")
	go s.runMetric.UpdateMetric(b)
	s.SendRequest(b)
}
