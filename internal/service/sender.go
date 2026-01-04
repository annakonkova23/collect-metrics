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

// Sender - структура для отправки метрик.
type Sender struct {
	client         *agent.Client   // Клиент для отправки метрик
	updMetric      *MetricsUpdater // Обновитель метрик
	url            string          // URL для отправки метрик
	reportInterval int             // Интервал отправки метрик
	logger         *zap.Logger     // Логгер
	key            string          // Ключ для подписи метрик
}

// NewSender - конструктор для Sender.
// Параметры:
// - url: URL для отправки метрик
// - pollInterval: Интервал опроса метрик
// - reportInterval: Интервал отправки метрик
// - key: Ключ для подписи метрик
// - logger: Логгер
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

// SendRequest - метод для отправки метрик.
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

// GetHash - метод для получения хеша метрик.
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

// GetURLForMetric - метод для получения URL для метрики.
func (s *Sender) GetURLForMetric(name, typeM, value string) string {
	return s.url + typeM + "/" + name + "/" + value
}

// Start - метод для запуска отправки метрик.
func (s *Sender) Start(ctx context.Context, numWorkers int) {
	s.logger.Info("Старт отправления метрик")
	chCalc := make(chan []byte, numWorkers)
	go s.updMetric.UpateUtilMetric(ctx, chCalc)
	go s.updMetric.UpdateRuntimeMetric(ctx, chCalc)
	for range numWorkers {
		go s.SendRequest(ctx, chCalc)
	}

}
