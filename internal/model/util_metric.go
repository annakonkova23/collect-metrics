package model

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type UtilMetric struct {
	TotalMemory     uint64
	FreeMemory      uint64
	CPUUtilization1 float64
}

func NewUtilMetric() *UtilMetric {
	return &UtilMetric{}
}

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

func (m *UtilMetric) GetTypeMetric() string {
	return Gauge
}
