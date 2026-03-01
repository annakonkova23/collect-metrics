package service

import (
	"context"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/model"
	"go.uber.org/zap"
)

const (
	typeUtil    = "util"
	typeRuntime = "runtime"
)

// MetricsUpdater - структура для обновления метрик.
type MetricsUpdater struct {
	logger       *zap.Logger          // Логгер
	runMetric    *model.RuntimeMetric // Метрика runtime
	utilMetric   *model.UtilMetric    // Метрика util
	pollInterval int                  // Интервал обновления метрик
}

// NewMetricsUpdater - конструктор для MetricsUpdater.
// Параметры:
// - logger: Логгер
// - pollInterval: Интервал обновления метрик
func NewMetricsUpdater(logger *zap.Logger, pollInterval int) *MetricsUpdater {
	return &MetricsUpdater{
		runMetric:    model.NewRuntimeMetric(),
		utilMetric:   model.NewUtilMetric(),
		logger:       logger,
		pollInterval: pollInterval,
	}
}

// UpdateRuntimeMetric - метод для обновления метрик runtime.
func (mu *MetricsUpdater) UpdateRuntimeMetric(ctx context.Context, chanel chan []*model.Metrics) {
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
			metrics, err := mu.codeMapRuntimeToMetrics(mu.runMetric.Metrics, typeRuntime)
			if err != nil {
				mu.logger.Error("Ошибка кодирования метрик",
					zap.Error(err),
				)
			} else {
				chanel <- metrics
			}
		}
	}
}

// UpateUtilMetric - метод для обновления метрик util.
func (mu *MetricsUpdater) UpateUtilMetric(ctx context.Context, chanel chan []*model.Metrics) {
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
			metrics, err := mu.codeMapRuntimeToMetrics(metricValue, typeUtil)
			if err != nil {
				mu.logger.Error("Ошибка кодирования метрик",
					zap.Error(err),
				)
			} else {
				chanel <- metrics
			}
		}
	}
}

// codeMapRuntimeToMetricsByte - метод для кодирования метрик в JSON.
func (mu *MetricsUpdater) codeMapRuntimeToMetrics(val map[string]float64, typeCalc string) ([]*model.Metrics, error) {
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
	return mcs, nil

}
