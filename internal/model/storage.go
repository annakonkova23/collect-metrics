package model

import (
	"encoding/json"
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

func (ms *MemStorage) SetMetric(name, typeMetr, value string) error {
	if name == "" {
		return fmt.Errorf("%s", "Имя метрики не может быть пустым")
	}
	if typeMetr != Counter && typeMetr != Gauge {
		msg := fmt.Sprintf("Некорректный тип метрики[%s]", typeMetr)
		return fmt.Errorf("%s", msg)
	}
	if value == "" {
		return fmt.Errorf("%s", "Не заполнено значение")

	}
	var deltaVal *int64
	var valueVal *float64
	ms.mx.Lock()
	defer ms.mx.Unlock()
	idx := slices.IndexFunc(ms.metrics, func(m *Metrics) bool {
		if m.ID == name && m.MType == typeMetr {
			return true
		}
		return false
	})
	if typeMetr == Counter {
		if delta, err := strconv.ParseInt(value, 10, 64); err != nil {
			msg := fmt.Sprintf("Некорректное значение счетчика [%s]", value)
			return fmt.Errorf("%s", msg)
		} else {
			if idx >= 0 {
				if ms.metrics[idx].Delta != nil {
					delta = *ms.metrics[idx].Delta + delta
				}
			}
			deltaVal = &delta

		}
	}

	if typeMetr == Gauge {
		if valueFloat, err := strconv.ParseFloat(value, 64); err != nil {
			msg := fmt.Sprintf("Некорректное значение float64 [%s]", value)
			return fmt.Errorf("%s", msg)
		} else {
			valueVal = &valueFloat
		}
	}
	if idx >= 0 {
		ms.metrics[idx].Value = valueVal
		ms.metrics[idx].Delta = deltaVal
	} else {
		ms.metrics = append(ms.metrics, &Metrics{
			ID:    name,
			MType: typeMetr,
			Delta: deltaVal,
			Value: valueVal,
		})
	}
	return nil
}

func (ms *MemStorage) String() string {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	js, _ := json.Marshal(ms.metrics)
	return string(js)
}

func (ms *MemStorage) GetMetric(name, typeMetric string) (string, bool) {
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

func (ms *MemStorage) GetMetricAllValues() []*Metrics {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	metrics := make([]*Metrics, len(ms.metrics))
	for i, m := range ms.metrics {
		metrics[i] = m.Copy()
	}
	return metrics
}
