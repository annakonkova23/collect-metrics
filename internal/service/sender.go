package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/agent"
	"github.com/annakonkova23/collect-metrics/internal/model"
	mcs "github.com/annakonkova23/collect-metrics/pkg/metrics"
	"go.uber.org/zap"
)

const (
	mode_http = "http"
	mode_grpc = "grpc"
)

// Sender - структура для отправки метрик.
type Sender struct {
	client         *agent.Client   // Клиент для отправки метрик
	updMetric      *MetricsUpdater // Обновитель метрик
	httpURL        string          // URL для отправки метрик
	reportInterval int             // Интервал отправки метрик
	logger         *zap.Logger     // Логгер
	key            string          // Ключ для подписи метрик
	modeSend       string          // Режим отправки метрик
	grpcURL        string          // URL для отправки метрик
}

// NewSender - конструктор для Sender.
// Параметры:
// - url: URL для отправки метрик
// - pollInterval: Интервал опроса метрик
// - reportInterval: Интервал отправки метрик
// - key: Ключ для подписи метрик
// - logger: Логгер
func NewSender(httpURL string, pollInterval, reportInterval int, key string, keyPath string, logger *zap.Logger, useGrpc bool, grpcURL string) *Sender {
	mode := mode_http
	if useGrpc {
		mode = mode_grpc
	}

	logger.Info("Установлен режим отправки метрик", zap.String("mode", mode))
	return &Sender{
		client:         agent.NewClient(keyPath, logger.Sugar()),
		updMetric:      NewMetricsUpdater(logger, pollInterval),
		httpURL:        httpURL,
		reportInterval: reportInterval,
		logger:         logger,
		key:            key,
		modeSend:       mode,
		grpcURL:        grpcURL,
	}
}

// SendRequest - метод для отправки метрик.
func (s *Sender) SendRequest(ctx context.Context, chMetrics <-chan []*model.Metrics) {
	s.logger.Info("Ждём расчета метрик")
	ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for metrics := range chMetrics {
				switch s.modeSend {
				case mode_http:
					{
						body, err := json.Marshal(metrics)
						if err != nil {
							s.logger.Error("Ошибка json.Marshal", zap.Error(err))
						}
						s.logger.Info(fmt.Sprintf("Запрос %s", string(body)))
						if err := s.client.PostWithBody(ctx, s.httpURL, body, s.GetHash(body)); err != nil {
							s.logger.Error(fmt.Sprintf("Ошибка отправки метрики %s: %v", string(body), err))
						} else {
							s.logger.Info("Успешный ответ")
						}
					}
				case mode_grpc:
					{
						s.logger.Info("GRPC запрос на отправку метрик")
						if err := s.client.SendRequest(ctx, s.grpcURL, convertMetricsToProtoMetrics(metrics)); err != nil {
							s.logger.Error("Ошибка отправки метрики", zap.Error(err))
						} else {
							s.logger.Info("Успешный ответ")
						}
					}
				}
			}

		}
	}
}

func convertMetricsToProtoMetrics(metrics []*model.Metrics) []*mcs.Metric {
	if len(metrics) == 0 {
		return nil
	}
	metricsProto := make([]*mcs.Metric, len(metrics))
	for i, metric := range metrics {
		metricsProto[i] = &mcs.Metric{}
		metricsProto[i].SetId(metric.ID)
		switch metric.MType {
		case model.Gauge:
			metricsProto[i].SetType(mcs.Metric_GAUGE)
			if metric.Value != nil {
				metricsProto[i].SetValue(*metric.Value)
			}
		case model.Counter:
			metricsProto[i].SetType(mcs.Metric_COUNTER)
			if metric.Delta != nil {
				metricsProto[i].SetDelta(*metric.Delta)
			}
		}
	}
	return metricsProto
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
	return s.httpURL + typeM + "/" + name + "/" + value
}

// Start - метод для запуска отправки метрик.
func (s *Sender) Start(ctx context.Context, numWorkers int) {
	s.logger.Info("Старт отправления метрик")
	chCalc := make(chan []*model.Metrics, numWorkers)
	go s.updMetric.UpateUtilMetric(ctx, chCalc)
	go s.updMetric.UpdateRuntimeMetric(ctx, chCalc)
	for range numWorkers {
		go s.SendRequest(ctx, chCalc)
	}

}
