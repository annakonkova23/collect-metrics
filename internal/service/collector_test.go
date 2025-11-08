package service_test

import (
	"context"
	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/config/db"
	"github.com/annakonkova23/collect-metrics/internal/model"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestCollector_GetMetricJson(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	metricExist := &model.Metrics{
		ID:    "StackSys",
		MType: "gauge",
	}
	value := float64(6)
	metricExistResult := &model.Metrics{
		ID:    "StackSys",
		MType: "gauge",
		Value: &value,
	}
	metricExistResultJSON, _ := metricExistResult.MarshalJSON()
	metricNotExist := &model.Metrics{
		ID:    "Stack",
		MType: "gauge",
	}
	tests := []struct {
		name   string
		metric *model.Metrics
		want   string
		err    error
	}{
		{
			name:   "ExistMetric",
			metric: metricExist,
			want:   string(metricExistResultJSON),
			err:    nil,
		},
		{
			name:   "NotExistMetric",
			metric: metricNotExist,
			want:   "",
			err:    service.ErrorNotFound,
		},
	}
	cfg := &config.ServerOptions{}
	dbConn := db.NewDbconnect(cfg.DatabaseDSN)
	c, err := service.NewCollector(context.Background(), cfg, logger, dbConn)
	if err != nil {
		t.Errorf("Ошибка при создании сервиса: %v", err)
	}
	_, _ = c.SaveMetric(metricExistResult)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetMetricJSON(tt.metric.ID, tt.metric.MType)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.want, got)

		})
	}
}
