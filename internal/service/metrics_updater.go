package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/model"
	"go.uber.org/zap"
)

const (
	typeUtil    = "util"
	typeRuntime = "runtime"
)

type MetricsUpdater struct {
	logger       *zap.Logger
	runMetric    *model.RuntimeMetric
	utilMetric   *model.UtilMetric
	pollInterval int
}

func NewMetricsUpdater(logger *zap.Logger, pollInterval int) *MetricsUpdater {
	return &MetricsUpdater{
		runMetric:    model.NewRuntimeMetric(),
		utilMetric:   model.NewUtilMetric(),
		logger:       logger,
		pollInterval: pollInterval,
	}
}

func (mu *MetricsUpdater) UpdateRuntimeMetric(ctx context.Context, chanel chan []byte) {
	mu.logger.Info("Запускаем обновление runtime метрик",
		zap.Int("pollInterval", mu.pollInterval),
	)
	ticker := time.NewTicker(time.Duration(mu.pollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			mu.logger.Info("Останавливаем обновление runtime метрик")
			return
		case <-ticker.C:
			mu.runMetric.UpdateMetric()
			jsMetrics, err := mu.codeMapRuntimeToMetricsByte(mu.runMetric.Metrics, typeRuntime)
			if err != nil {
				mu.logger.Error("Ошибка кодирования метрик",
					zap.Error(err),
				)
			} else {
				chanel <- jsMetrics
			}
		}
	}
}

func (mu *MetricsUpdater) UpateUtilMetric(ctx context.Context, chanel chan []byte) {
	mu.logger.Info("Запускаем обновление util метрик",
		zap.Int("pollInterval", mu.pollInterval),
	)
	ticker := time.NewTicker(time.Duration(mu.pollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			mu.logger.Info("Останавливаем обновление runtime метрик")
			return
		case <-ticker.C:
			metricValue, err := mu.utilMetric.CalcMetric()
			if err != nil {
				mu.logger.Error("Ошибка расчёта метрик",
					zap.Error(err),
				)
			}
			jsMetrics, err := mu.codeMapRuntimeToMetricsByte(metricValue, typeUtil)
			if err != nil {
				mu.logger.Error("Ошибка кодирования метрик",
					zap.Error(err),
				)
			} else {
				chanel <- jsMetrics
			}
		}
	}
}

func (mu *MetricsUpdater) codeMapRuntimeToMetricsByte(val map[string]float64, typeCalc string) ([]byte, error) {
	mcs := make([]*model.Metrics, len(val))
	typeM := ""
	i := 0
	for k, v := range val {
		if typeCalc == typeUtil {
			typeM = mu.utilMetric.GetTypeMetric()
		} else {
			typeM = mu.runMetric.GetTypeMetric(k)
		}
		metric := &model.Metrics{ID: k, MType: typeM}
		if typeM == model.Counter {
			delta := int64(v)
			metric.Delta = &delta
		} else {
			metric.Value = &v
		}
		mcs[i] = metric
		i++
	}
	result, err := json.Marshal(mcs)
	if err != nil {
		return nil, err
	}
	return result, nil

}
