package model

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

type MemStorage struct {
	metrics sync.Map
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: sync.Map{},
	}
}

func (s *MemStorage) SetMetric(name, typeMetr, value string) error {
	if name == "" {
		return fmt.Errorf("%s", "name is empty")
	}
	if typeMetr != Counter && typeMetr != Gauge {
		return fmt.Errorf("Некорректный тип метрики[%s]", typeMetr)
	}
	if value == "" {
		return fmt.Errorf("%s", "Не заполнено значение")

	}
	var deltaVal *int64
	var valueVal *float64
	log.Println(typeMetr)
	log.Println(name)
	if typeMetr == Counter {
		if delta, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("Некорректное значение счетчика [%s]", value)
		} else {
			log.Println(delta)
			valueMetric, ok := s.metrics.Load(name)
			if !ok {
				deltaVal = &delta
				vF := float64(delta)
				valueVal = &vF
			} else {
				if valueMetric.(Metrics).Value != nil {
					valueFloat := *valueMetric.(Metrics).Value + float64(delta)
					valueVal = &valueFloat
					deltaVal = &delta
				}

			}
		}

	}
	if typeMetr == Gauge {
		if valueFloat, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("Некорректное значение счетчика [%s]", value)
		} else {
			valueVal = &valueFloat
		}
	}
	s.metrics.Store(name, Metrics{
		ID:    name,
		MType: typeMetr,
		Delta: deltaVal,
		Value: valueVal,
	})
	return nil
}

func (m *MemStorage) String() string {
	str := make([]string, 0)
	m.metrics.Range(func(key, value interface{}) bool {
		// Проверка типа ключа (в данном примере ожидаем строку)
		js, _ := json.Marshal(value.(Metrics))
		str = append(str, string(js))
		return true
	})
	return strings.Join(str, "\n")
}
