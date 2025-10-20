package service

import (
	"errors"
	"fmt"
	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/model"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

var ErrorNotFound = errors.New("not exists name metric")

type Collector struct {
	MemStorage      *model.MemStorage
	logger          *zap.Logger
	FileStoragePath string
	StoreInterval   int
	mu              sync.Mutex
}

type Metric struct {
	Name  string
	Type  string
	Value string
}

func NewCollector(cfg *config.ServerOptions, logger *zap.Logger) (*Collector, error) {
	clr := &Collector{
		FileStoragePath: cfg.FileStoragePath,
		StoreInterval:   cfg.StoreInterval,
		logger:          logger,
		MemStorage:      model.NewMemStorage(),
	}
	if cfg.Restore {
		clr.logger.Info("Инициализация метрик", zap.Bool("Restore", cfg.Restore))
		metrics := clr.LoadFromFile()
		clr.MemStorage.InitMetrics(metrics)
	}
	if cfg.StoreInterval > 0 {
		go clr.ProcessUploadFile()
	}
	return clr, nil
}

func (c *Collector) SaveMetricsByParam(name, typeMetric, value string) error {
	metric := &model.Metrics{
		ID:    name,
		MType: typeMetric,
	}
	if typeMetric == model.Counter {
		if delta, err := strconv.ParseInt(value, 10, 64); err == nil {
			metric.Delta = &delta
		} else {
			return fmt.Errorf("неверный формат значения для счетчика: %s", value)
		}
	}
	if typeMetric == model.Gauge {
		if valueFloat, err := strconv.ParseFloat(value, 64); err == nil {
			metric.Value = &valueFloat
		} else {
			return fmt.Errorf("неверный формат значения для float64: %s", value)
		}
	}
	_, err := c.SaveMetric(metric)
	if err != nil {
		return err
	}
	return nil
}

func (c *Collector) SaveMetric(metric *model.Metrics) (*model.Metrics, error) {
	metric, err := c.MemStorage.SetMetric(metric)
	if err != nil {
		return nil, err
	}
	if c.StoreInterval == 0 {
		err := c.SaveToFile()
		if err != nil {
			c.logger.Error("Ошибка сохранения файла", zap.Error(err),
				zap.String("nameMetric", metric.ID),
				zap.String("typeMetric", metric.MType),
				zap.Float64("value", *metric.Value),
			)

		}
	}
	return metric, nil
}

func (c *Collector) GetMetricValueByParam(nameMetric, typeMetric string) (string, bool) {
	value, ok := c.MemStorage.GetMetricValue(nameMetric, typeMetric)
	return value, ok

}

func (c *Collector) GetMetricJSON(nameMetric, typeMetric string) (string, error) {
	metric, ok := c.MemStorage.GetMetric(nameMetric, typeMetric)
	if !ok {
		return "", ErrorNotFound
	}
	js, _ := metric.MarshalJSON()
	metricJSON := string(js)
	c.logger.Info("GetMetricJson", zap.String("metricJson", metricJSON))
	return metricJSON, nil

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

func (c *Collector) ProcessUploadFile() {
	c.logger.Info("Запуск процесса сохранения файла", zap.Int("StoreInterval", c.StoreInterval))
	for {
		err := c.SaveToFile()
		if err != nil {
			c.logger.Error("Ошибка сохранения файла", zap.Error(err))
		}
		time.Sleep(time.Duration(c.StoreInterval) * time.Second)
	}
}
