package model

import (
	"math/rand"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func BenchmarkMemStorage_SetMetric_Gauge(b *testing.B) {
	storage := NewMemStorage()
	metric := &Metrics{
		ID:    "gauge_test",
		MType: Gauge,
		Value: floatPtr(3.1415),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = storage.SetMetric(metric)
		}
	})
}

func BenchmarkMemStorage_SetMetric_Counter(b *testing.B) {
	storage := NewMemStorage()
	metric := &Metrics{
		ID:    "counter_test",
		MType: Counter,
		Delta: intPtr(1),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = storage.SetMetric(metric)
		}
	})
}

func BenchmarkMemStorage_SetMetric_UpdateCounter(b *testing.B) {
	// Сначала добавим метрику
	storage := NewMemStorage()
	first := &Metrics{
		ID:    "counter_update",
		MType: Counter,
		Delta: intPtr(1),
	}
	storage.SetMetric(first)

	update := &Metrics{
		ID:    "counter_update",
		MType: Counter,
		Delta: intPtr(1),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = storage.SetMetric(update)
		}
	})
}

func BenchmarkMemStorage_SetMetric_ConcurrentMixed(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Случайный тип метрики
			metricType := Gauge
			var metric *Metrics
			if rand.Intn(2) == 0 {
				metricType = Counter
				metric = &Metrics{
					ID:    "mixed_" + randString(5),
					MType: metricType,
					Delta: intPtr(1),
				}
			} else {
				metric = &Metrics{
					ID:    "mixed_" + randString(5),
					MType: metricType,
					Value: floatPtr(rand.Float64() * 100),
				}
			}

			_, _ = storage.SetMetric(metric)
		}
	})
}

// Вспомогательные функции
func intPtr(i int64) *int64 {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
