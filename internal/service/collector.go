package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/annakonkova23/collect-metrics/internal/model"
	"go.uber.org/zap"
)

var ErrorNotFound = errors.New("not exists name metric")

type Collector struct {
	MemStorage *model.MemStorage
	logger     *zap.Logger
}

type Metric struct {
	Name  string
	Type  string
	Value string
}

func NewCollector(logger *zap.Logger) *Collector {
	return &Collector{MemStorage: model.NewMemStorage(), logger: logger}
}

func (c *Collector) ParseAndSaveMetricsByURL(url string) error {
	var typeMetric, nameMetric, value string
	parts := strings.Split(url, "/")
	if len(parts) < 2 || parts[1] != "update" {
		return fmt.Errorf("%s", "Невалидный Url")
	}
	if len(parts) == 3 || (len(parts) > 3 && parts[3] == "") {
		return fmt.Errorf("%s", ErrorNotFound)
	}
	if len(parts) >= 4 {
		typeMetric = parts[2]
		nameMetric = parts[3]
		if len(parts) >= 5 {
			value = parts[4]
		}
	}
	err := c.MemStorage.SetMetric(nameMetric, typeMetric, value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Collector) ParseAndSaveMetricsByParam(name, typeMetric, value string) error {
	err := c.MemStorage.SetMetric(name, typeMetric, value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Collector) SaveMetric(metric *model.Metrics) (*model.Metrics, error) {
	metric, err := c.MemStorage.SetMetricByMetric(metric)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func (c *Collector) GetMetricValue(url string) (string, bool, error) {
	var typeMetric, nameMetric string
	parts := strings.Split(url, "/")
	if len(parts) < 2 || parts[1] != "value" {
		return "", false, fmt.Errorf("%s", "Невалидный Url")
	}
	if len(parts) == 3 || (len(parts) > 3 && parts[3] == "") {
		return "", false, fmt.Errorf("%s", ErrorNotFound)
	}
	if len(parts) >= 4 {
		typeMetric = parts[2]
		nameMetric = parts[3]
	}
	value, ok := c.GetMetricValueByParam(nameMetric, typeMetric)
	return value, ok, nil

}

func (c *Collector) GetMetricValueByParam(nameMetric, typeMetric string) (string, bool) {
	value, ok := c.MemStorage.GetMetricValue(nameMetric, typeMetric)
	return value, ok

}

func (c *Collector) GetMetricJson(nameMetric, typeMetric string) string {
	metric := c.MemStorage.GetMetric(nameMetric, typeMetric)
	js, _ := metric.MarshalJSON()
	metricJson := string(js)
	c.logger.Info("GetMetricJson", zap.String("metricJson", metricJson))
	return metricJson

}

func (c *Collector) GetMetricAllValues() []*Metric {
	metric := c.MemStorage.GetMetricAllValues()
	metricResult := make([]*Metric, len(metric))
	for i, m := range metric {
		metricResult[i] = &Metric{Name: m.ID, Type: m.MType}
		value := ""
		if m.MType == model.Counter {
			if m.Delta != nil {
				value = strconv.FormatInt(*m.Delta, 10)
			}
		}
		if m.MType == model.Gauge {
			if m.Value != nil {
				value = strconv.FormatFloat(*m.Value, 'f', -1, 64)
			}
		}
		metricResult[i].Value = value
	}
	return metricResult

}
