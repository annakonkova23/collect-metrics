package model_test

import (
	"testing"

	"github.com/annakonkova23/collect-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_GetMetric(t *testing.T) {
	value := float64(6)
	metricExistResult := &model.Metrics{
		ID:    "StackSys",
		MType: "gauge",
		Value: &value,
	}
	tests := []struct {
		name       string // description of this test case
		nameMetric string
		typeMetric string
		want       *model.Metrics
		ok         bool
	}{
		{
			name:       "ExistMetric",
			nameMetric: "StackSys",
			typeMetric: "gauge",
			want:       metricExistResult,
			ok:         true,
		},
		{
			name:       "NotExistMetric",
			nameMetric: "Stack",
			typeMetric: "gauge",
			want:       nil,
			ok:         false,
		},
	}
	ms := model.NewMemStorage()
	ms.SetMetric(metricExistResult)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, ok := ms.GetMetric(tt.nameMetric, tt.typeMetric)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.ok, ok)
		})
	}
}
