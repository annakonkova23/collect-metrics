package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/agent"
	"go.uber.org/zap"
)

type Sender struct {
	client         *agent.Client
	updMetric      *MetricsUpdater
	url            string
	reportInterval int
	logger         *zap.Logger
	key            string
}

func NewSender(url string, pollInterval, reportInterval int, key string, logger *zap.Logger) *Sender {
	return &Sender{
		client:         agent.NewClient(logger.Sugar()),
		updMetric:      NewMetricsUpdater(logger, pollInterval),
		url:            url,
		reportInterval: reportInterval,
		logger:         logger,
		key:            key,
	}
}
func (s *Sender) SendRequest(ctx context.Context, bodys <-chan []byte) {
	s.logger.Info("Ждём расчета метрик")
	ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for body := range bodys {
				s.logger.Info(fmt.Sprintf("Запрос %s", string(body)))
				if err := s.client.PostWithBody(ctx, s.url, body, s.GetHash(body)); err != nil {
					s.logger.Error(fmt.Sprintf("Ошибка отправки метрики %s: %v", string(body), err))
				} else {
					s.logger.Info("Успешный ответ")
				}
			}
		}

	}
}

func (s *Sender) GetHash(body []byte) string {
	if s.key == "" {
		return ""
	}
	secretKey := []byte(s.key)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(body))
	signature := h.Sum(nil)
	signatureHex := hex.EncodeToString(signature)
	return signatureHex
}

func (s *Sender) GetURLForMetric(name, typeM, value string) string {
	return s.url + typeM + "/" + name + "/" + value
}

func (s *Sender) Start(ctx context.Context, numWorkers int) {
	s.logger.Info("Старт отправления метрик")
	chCalc := make(chan []byte, numWorkers)
	go s.updMetric.UpateUtilMetric(ctx, chCalc)
	go s.updMetric.UpdateRuntimeMetric(ctx, chCalc)
	for range numWorkers {
		go s.SendRequest(ctx, chCalc)
	}

}
