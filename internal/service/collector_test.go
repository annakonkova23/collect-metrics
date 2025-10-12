package service_test

import (
	//"fmt"
	"testing"

	"github.com/annakonkova23/collect-metrics/internal/model"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCollector_ParseAndSaveMetricsByURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		err     error
	}{
		{
			name:    "Positive",
			url:     "/update/gauge/StackSys/1343488",
			wantErr: false,
		},
		{
			name:    "Negative not type",
			url:     "/update/StackSys/1343488",
			wantErr: true,
		},
		{
			name:    "Negative not correct type",
			url:     "/update/gaue/StackSys/1343488",
			wantErr: true,
		},
		{
			name:    "Negative not name",
			url:     "/update/gauge/",
			wantErr: true,
			err:     service.ErrorNotFound,
		},
		{
			name:    "Negative not correct value",
			url:     "/update/gauge/StackSys/atyj",
			wantErr: true,
		},
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := service.NewCollector(logger)
			gotErr := c.ParseAndSaveMetricsByURL(tt.url)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ParseAndSaveMetricsByURL() failed: %v", gotErr)
				} else if tt.err != nil {
					assert.EqualError(t, gotErr, tt.err.Error())
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ParseAndSaveMetricsByURL() succeeded unexpectedly")
			}
		})
	}
}

func TestCollector_GetMetricJson(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
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
	metricExistResultJson, _ := metricExistResult.MarshalJSON()
	metricNotExist := &model.Metrics{
		ID:    "Stack",
		MType: "gauge",
	}
	metricNotExistResult := &model.Metrics{}
	metricNotExistResultJson, _ := metricNotExistResult.MarshalJSON()
	tests := []struct {
		name string
		// Named input parameters for target function.
		metric *model.Metrics
		want   string
	}{
		{
			name:   "ExistMetric",
			metric: metricExist,
			want:   string(metricExistResultJson),
		},
		{
			name:   "NotExistMetric",
			metric: metricNotExist,
			want:   string(metricNotExistResultJson),
		},
	}
	c := service.NewCollector(logger)
	_ = c.SaveMetric(metricExistResult)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.GetMetricJson(tt.metric.ID, tt.metric.MType)
			assert.Equal(t, tt.want, got)

		})
	}
}
