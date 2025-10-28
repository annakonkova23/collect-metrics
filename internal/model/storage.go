package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
)

type MemStorage struct {
	mx      sync.RWMutex
	metrics []*Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make([]*Metrics, 0),
	}
}

func (ms *MemStorage) SetMetric(metric *Metrics) (*Metrics, error) {
	var errs []error

	if metric.ID == "" {
		msg := "Имя метрики не может быть пустым"
		errs = append(errs, errors.New(msg))
	}

	if metric.MType != Counter && metric.MType != Gauge {
		msg := fmt.Sprintf("Некорректный тип метрики[%s]", metric.MType)
		errs = append(errs, errors.New(msg))
	}

	if metric.MType == Counter {
		if metric.Delta == nil {
			msg := fmt.Sprintf("Некорректное значение метрики [%s]", metric.ID)
			errs = append(errs, errors.New(msg))
		}
	} else if metric.MType == Gauge {
		if metric.Value == nil {
			msg := fmt.Sprintf("Некорректное значение метрики [%s]", metric.ID)
			errs = append(errs, errors.New(msg))
		}
	}

	// Если есть ошибки, возвращаем их объединение
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	ms.mx.Lock()
	defer ms.mx.Unlock()

	idx := slices.IndexFunc(ms.metrics, func(m *Metrics) bool {
		return m.ID == metric.ID && m.MType == metric.MType
	})

	var metricResult *Metrics

	if metric.MType == Counter {
		if idx >= 0 {
			if ms.metrics[idx].Delta != nil {
				delta := *ms.metrics[idx].Delta + *metric.Delta
				ms.metrics[idx].Delta = &delta
				metricResult = ms.metrics[idx].Copy()
			}
		} else {
			ms.metrics = append(ms.metrics, metric)
			metricResult = metric
		}
	} else if metric.MType == Gauge {
		if idx >= 0 {
			value := *metric.Value
			ms.metrics[idx].Value = &value
			metricResult = ms.metrics[idx].Copy()
		} else {
			ms.metrics = append(ms.metrics, metric)
			metricResult = metric
		}
	}

	return metricResult, nil
}

func (ms *MemStorage) String() string {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	js, _ := json.Marshal(ms.metrics)
	return string(js)
}

func (ms *MemStorage) GetMetricValue(name, typeMetric string) (string, bool) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	idx := slices.IndexFunc(ms.metrics, func(m *Metrics) bool {
		if m.ID == name && m.MType == typeMetric {
			return true
		}
		return false
	})
	if idx >= 0 {
		if typeMetric == Counter {
			if ms.metrics[idx].Delta != nil {
				return strconv.FormatInt(*ms.metrics[idx].Delta, 10), true
			}

		}
		if typeMetric == Gauge {
			if ms.metrics[idx].Value != nil {
				return strconv.FormatFloat(*ms.metrics[idx].Value, 'f', -1, 64), true
			}
		}

	}
	return "", false
}

func (ms *MemStorage) GetMetric(name, typeMetric string) (*Metrics, bool) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	idx := slices.IndexFunc(ms.metrics, func(m *Metrics) bool {
		if m.ID == name && m.MType == typeMetric {
			return true
		}
		return false
	})
	if idx >= 0 {

		return ms.metrics[idx].Copy(), true
	}
	return nil, false
}

func (ms *MemStorage) GetMetricAllValues() []*Metrics {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	metrics := make([]*Metrics, len(ms.metrics))
	for i, m := range ms.metrics {
		metrics[i] = m.Copy()
	}
	return metrics
}

func (ms *MemStorage) InitMetrics(mts []*Metrics) {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	ms.metrics = mts
}
