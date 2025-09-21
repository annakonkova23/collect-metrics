package model

import (
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type RMetric int

const (
	pollInterval = 2
	nameCounter  = "PollCount"
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
	Metrics map[string]float64
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

func NewRuntimeMetric() *RuntimeMetric {
	return &RuntimeMetric{
		Metrics: make(map[string]float64),
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

		}
	}
	valMetric["RandomValue"] = rand.Float64()
	return valMetric
}

func (rm *RuntimeMetric) UpdateMetric(chanel chan bool) {
	log.Println("Запускаем обновление метрик")
	var once sync.Once
	for {
		rm.Metrics = rm.CalcMetric()
		once.Do(func() {
			log.Println("Посчитали метрики")
			close(chanel)
		})
		rm.Metrics[nameCounter]++
		time.Sleep(pollInterval * time.Second)
	}
}

func (rm *RuntimeMetric) GetTypeMetric(name string) string {
	if name == nameCounter {
		return Counter
	} else {
		return Gauge
	}
}

func (rm *RuntimeMetric) GetMetrics() map[string]float64 {
	return rm.Metrics

}
