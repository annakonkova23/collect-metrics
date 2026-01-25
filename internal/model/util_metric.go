package model

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// generate:reset
// UtilMetric - метрики утилизации
type UtilMetric struct {
	TotalMemory     uint64
	FreeMemory      uint64
	CPUUtilization1 float64
}

// NewUtilMetric - конструктор.
func NewUtilMetric() *UtilMetric {
	return &UtilMetric{}
}

// CalcUtilMetrics - расчет метрик утилизации.
func CalcUtilMetrics() (*UtilMetric, error) {

	memStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	cpuPercentages, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, err
	}

	return &UtilMetric{
		TotalMemory:     memStats.Total,
		FreeMemory:      memStats.Free,
		CPUUtilization1: cpuPercentages[0],
	}, nil
}

// CalcMetric - расчет метрики.
func (m *UtilMetric) CalcMetric() (map[string]float64, error) {
	utilMetric, err := CalcUtilMetrics()
	if err != nil {
		return nil, err
	}
	result := make(map[string]float64)
	result["TotalMemory"] = float64(utilMetric.TotalMemory)
	result["FreeMemory"] = float64(utilMetric.FreeMemory)
	result["CPUUtilization1"] = utilMetric.CPUUtilization1
	return result, nil
}

// GetTypeMetric - тип метрики.
func (m *UtilMetric) GetTypeMetric() string {
	return Gauge
}
