package service

import (
	"context"
	"encoding/json"
	"fmt"
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
func (s *Sender) SendRequest(ctx context.Context, b chan bool) {
	s.logger.Info("Ждём расчета метрик")
	<-b
	ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("11111111timer")
			metrics := s.runMetric.GetMetrics()
			if len(metrics) == 0 {
				continue
			}
			body, _ := json.Marshal(metrics)
			s.logger.Info(fmt.Sprintf("Запрос %s", string(body)))
			if err := s.client.PostWithBody(ctx, s.url, body); err != nil {
				s.logger.Info(fmt.Sprintf("Ошибка отправки метрики %s: %v", string(body), err))
			} else {
				s.logger.Info("Успешный ответ")
			}
		}

	}
}

func (s *Sender) GetURLForMetric(name, typeM, value string) string {
	return s.url + typeM + "/" + name + "/" + value
}

func (s *Sender) Start(ctx context.Context) {
	b := make(chan bool)
	s.logger.Info("Старт отправления метрик")
	go s.runMetric.UpdateMetric(b)
	s.SendRequest(ctx, b)
}
