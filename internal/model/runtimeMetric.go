package model

import (
	"go.uber.org/zap"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type RMetric int

const (
	nameCounter = "PollCount"
)
const (
	Alloc RMetric = iota
	BuckHashSys
	Frees
	GCCPUFraction
	GCSys
	HeapAlloc
	HeapIdle
	HeapInuse
	HeapObjects
	HeapReleased
	HeapSys
	LastGC
	Lookups
	MCacheInuse
	MCacheSys
	MSpanInuse
	MSpanSys
	Mallocs
	NextGC
	NumForcedGC
	NumGC
	OtherSys
	PauseTotalNs
	StackInuse
	StackSys
	Sys
	TotalAlloc
	PollCount
	RandomValue
)

var allMetrics = []RMetric{Alloc,
	BuckHashSys,
	Frees,
	GCCPUFraction,
	GCSys,
	HeapAlloc,
	HeapIdle,
	HeapInuse,
	HeapObjects,
	HeapReleased,
	HeapSys,
	LastGC,
	Lookups,
	MCacheInuse,
	MCacheSys,
	MSpanInuse,
	MSpanSys,
	Mallocs,
	NextGC,
	NumForcedGC,
	NumGC,
	OtherSys,
	PauseTotalNs,
	StackInuse,
	StackSys,
	Sys,
	TotalAlloc,
}

type RuntimeMetric struct {
	Metrics      map[string]float64
	mx           sync.RWMutex
	pollInterval int
	logger       *zap.Logger
}

func (r RMetric) String() string {
	return [...]string{"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}[r]
}

func NewRuntimeMetric(pollInterval int, logger *zap.Logger) *RuntimeMetric {
	return &RuntimeMetric{
		Metrics:      make(map[string]float64),
		pollInterval: pollInterval,
		logger:       logger,
	}
}
func (rm *RuntimeMetric) CalcMetric() map[string]float64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	valMetric := make(map[string]float64)
	for _, mc := range allMetrics {
		metric := mc.String()
		switch mc {
		case Alloc:
			valMetric[metric] = float64(stats.Alloc)
		case BuckHashSys:
			valMetric[metric] = float64(stats.BuckHashSys)
		case GCCPUFraction:
			valMetric[metric] = stats.GCCPUFraction
		case GCSys:
			valMetric[metric] = float64(stats.GCSys)
		case HeapAlloc:
			valMetric[metric] = float64(stats.HeapAlloc)
		case HeapIdle:
			valMetric[metric] = float64(stats.HeapIdle)
		case HeapInuse:
			valMetric[metric] = float64(stats.HeapInuse)
		case HeapObjects:
			valMetric[metric] = float64(stats.HeapObjects)
		case HeapReleased:
			valMetric[metric] = float64(stats.HeapReleased)
		case HeapSys:
			valMetric[metric] = float64(stats.HeapSys)
		case LastGC:
			valMetric[metric] = float64(stats.LastGC)
		case Lookups:
			valMetric[metric] = float64(stats.Lookups)
		case MCacheInuse:
			valMetric[metric] = float64(stats.MCacheInuse)
		case MCacheSys:
			valMetric[metric] = float64(stats.MCacheSys)
		case MSpanInuse:
			valMetric[metric] = float64(stats.MSpanInuse)
		case MSpanSys:
			valMetric[metric] = float64(stats.MSpanSys)
		case Mallocs:
			valMetric[metric] = float64(stats.Mallocs)
		case NextGC:
			valMetric[metric] = float64(stats.NextGC)
		case NumForcedGC:
			valMetric[metric] = float64(stats.NumForcedGC)
		case NumGC:
			valMetric[metric] = float64(stats.NumGC)
		case OtherSys:
			valMetric[metric] = float64(stats.OtherSys)
		case PauseTotalNs:
			valMetric[metric] = float64(stats.PauseTotalNs)
		case StackInuse:
			valMetric[metric] = float64(stats.StackInuse)
		case StackSys:
			valMetric[metric] = float64(stats.StackSys)
		case Sys:
			valMetric[metric] = float64(stats.Sys)
		case TotalAlloc:
			valMetric[metric] = float64(stats.TotalAlloc)
		case Frees:
			valMetric[metric] = float64(stats.Frees)
		}
	}
	valMetric["RandomValue"] = rand.Float64()
	return valMetric
}

func (rm *RuntimeMetric) UpdateMetric(chanel chan bool) {
	rm.logger.Info("Запускаем обновление метрик",
		zap.Int("pollInterval", rm.pollInterval),
	)
	var once sync.Once
	cnt := 1
	for {
		valMetric := rm.CalcMetric()
		once.Do(func() {
			rm.logger.Info("Метрики первый раз рассчитались")
			close(chanel)
		})
		rm.mx.Lock()
		for k, v := range valMetric {
			rm.Metrics[k] = v
		}
		rm.logger.Info("Метрики рассчитались", zap.Int("cnt", cnt))
		cnt++
		rm.Metrics[nameCounter] = rm.Metrics[nameCounter] + 1
		rm.mx.Unlock()
		time.Sleep(time.Duration(rm.pollInterval) * time.Second)
	}
}

func (rm *RuntimeMetric) GetTypeMetric(name string) string {
	if name == nameCounter {
		return Counter
	} else {
		return Gauge
	}
}

func (rm *RuntimeMetric) GetMetricsValue() map[string]float64 {
	rm.mx.RLock()
	defer rm.mx.RUnlock()
	return rm.Metrics

}

func (rm *RuntimeMetric) GetMetrics() []*Metrics {
	rm.mx.RLock()
	defer rm.mx.RUnlock()
	mcs := make([]*Metrics, len(rm.Metrics))
	i := 0
	for k, v := range rm.Metrics {
		typeM := rm.GetTypeMetric(k)
		metric := &Metrics{ID: k, MType: typeM}
		if typeM == Counter {
			delta := int64(v)
			metric.Delta = &delta
		} else {
			metric.Value = &v
		}
		mcs[i] = metric
		i++
	}
	return mcs

}
