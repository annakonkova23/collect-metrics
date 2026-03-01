package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/model"
	"github.com/annakonkova23/collect-metrics/internal/repository"
	fh "github.com/annakonkova23/collect-metrics/internal/service/fileHandler"
	mcs "github.com/annakonkova23/collect-metrics/pkg/metrics"
	"go.uber.org/zap"
)

var ErrorNotFound = errors.New("not exists name metric")

// Collector - структура для сбора метрик и их хранения.
type Collector struct {
	MemStorage    *model.MemStorage   // Хранилище метрик в памяти
	logger        *zap.Logger         // Логгер для вывода сообщений
	fileHandler   *fh.FileHandler     // Обработчик файлов для сохранения метрик
	StoreInterval int                 // Интервал сохранения метрик в файл или БД
	mx            sync.RWMutex        // Мьютекс для синхронизации доступа к хранилищу метрик
	conn          *repository.DBStore // Соединение с БД для сохранения метрик
}

type Metric struct {
	Name  string
	Type  string
	Value string
}

// NewCollector - конструктор для создания экземпляра Collector.

// Параметры:
//   - ctx: контекст приложения
//   - cfg: конфигурация сервера
//   - logger: логгер для вывода сообщений
//   - dbConnect: соединение с БД

// Возвращает:
//   - *Collector: указатель на созданный экземпляр Collector
//   - error: ошибка при создании экземпляра или при инициализации хранилища метрик
func NewCollector(ctx context.Context, cfg *config.ServerOptions, logger *zap.Logger, dbConnect *sql.DB) (*Collector, error) {
	clr := &Collector{
		StoreInterval: cfg.StoreInterval,
		logger:        logger,
		MemStorage:    model.NewMemStorage(),
	}
	clr.conn = repository.NewDBStore(logger, dbConnect)
	fileHandler, err := fh.NewFileHandler(cfg.FileStoragePath, logger)
	if err != nil {
		logger.Error("Ошибка создания обработчика файлов", zap.Error(err))
	} else {
		clr.fileHandler = fileHandler
	}
	if clr.conn.IsConnectDB() {
		err := clr.conn.CreateObjectDB()
		if err != nil {
			return nil, err
		}
	}
	if cfg.Restore {
		clr.logger.Info("Инициализация метрик", zap.Bool("Restore", cfg.Restore))
		if clr.conn.IsConnectDB() {
			if err := clr.initMemStorageFromDB(ctx); err != nil {
				return nil, err
			}
		} else if fileHandler != nil {
			if err := clr.initMemStorageFromFile(); err != nil {
				return nil, err
			}
		} else {
			clr.logger.Warn("Инициализация метрик не выполнена, так как не указан способ восстановления метрик")
		}
	}
	if clr.conn == nil && cfg.StoreInterval > 0 && clr.fileHandler != nil {
		go clr.ProcessUploadFile(ctx)
	}
	return clr, nil
}

// initMemStorageFromFile - инициализация хранилища метрик из файла.
func (c *Collector) initMemStorageFromFile() error {
	data := c.fileHandler.LoadFromFile()
	var metrics []*model.Metrics
	if len(data) > 0 {
		err := json.Unmarshal(data, &metrics)
		if err != nil {
			msg := fmt.Sprintf("Ошибка парсинга метрик в структуру: %s", err.Error())
			return fmt.Errorf("%s", msg)
		}
		c.MemStorage.InitMetrics(metrics)

	}
	return nil
}

// initMemStorageFromDB - инициализация хранилища метрик из БД.
func (c *Collector) initMemStorageFromDB(ctx context.Context) error {
	metrics, err := c.conn.RetryLoadMetricsToDB(ctx, 1)
	if err != nil {
		return err
	}
	if len(metrics) > 0 {
		c.MemStorage.InitMetrics(metrics)
	}
	return nil
}

// SaveMetricByParam - сохранение метрики по имени и типу.
func (c *Collector) SaveMetricsByParam(ctx context.Context, name, typeMetric, value string) error {
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
	_, err := c.SaveMetrics(ctx, []*model.Metrics{metric})
	if err != nil {
		return err
	}
	return nil
}

// GetMetricValueByParam - получение значения метрики по имени и типу.
func (c *Collector) GetMetricValueByParam(nameMetric, typeMetric string) (string, bool) {
	value, ok := c.MemStorage.GetMetricValue(nameMetric, typeMetric)
	return value, ok

}

// GetMetricJSON - получение метрики в формате JSON.
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

// GetMetricAllValues - получение всех метрик.
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

// ProcessUploadFile - процесс сохранения метрик в файл.
func (c *Collector) ProcessUploadFile(ctx context.Context) {
	if c.StoreInterval == 0 {
		c.logger.Info("Процесс записи в файл не будет запущен")
		return
	}
	c.logger.Info("Запуск процесса сохранения файла", zap.Int("StoreInterval", c.StoreInterval))
	timer := time.NewTimer(time.Duration(c.StoreInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Прерывание процесса сохранения в файл")
			return
		case <-timer.C:
			c.GetDataAndSaveToFile()
		}
	}
}

// GetDataAndSaveToFile - получение данных из хранилища и сохранение в файл.
func (c *Collector) GetDataAndSaveToFile() {
	if c.fileHandler == nil {
		return
	}
	metrics := c.MemStorage.GetMetricAllValues()
	js, err := json.Marshal(metrics)
	if err != nil {
		c.logger.Error("Ошибка парсинга структуры в json: %s", zap.Error(err))
	}
	err = c.fileHandler.SaveToFile(js)
	if err != nil {
		c.logger.Error("Ошибка сохранения файла", zap.Error(err))
	}
}

// SaveMetrics - сохранение метрик в хранилище.
func (c *Collector) SaveMetrics(ctx context.Context, metrics []*model.Metrics) ([]*model.Metrics, error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	metricsNew := []*model.Metrics{}
	for _, metric := range metrics {
		metric, err := c.MemStorage.SetMetric(metric)
		if err != nil {
			return nil, err
		}
		metricsNew = append(metricsNew, metric)
	}
	if c.conn.IsConnectDB() {
		err := c.conn.SaveMetricsToDB(ctx, metricsNew)
		if err != nil {
			c.logger.Error("Ошибка сохранения метрик в БД", zap.Error(err))
		}
	}
	if c.StoreInterval == 0 && c.fileHandler != nil {
		c.GetDataAndSaveToFile()
	}
	return metrics, nil

}

func (c *Collector) SaveMetricsByProto(ctx context.Context, metricsProto []*mcs.Metric) error {
	metrics := convertProtoMetricsToMetrics(metricsProto)
	_, err := c.SaveMetrics(ctx, metrics)
	if err != nil {
		return err
	}
	return nil
}

func convertProtoMetricsToMetrics(protoMetrics []*mcs.Metric) []*model.Metrics {
	var metrics []*model.Metrics

	for _, pMetric := range protoMetrics {
		metric := &model.Metrics{
			ID: pMetric.GetId(),
		}

		switch pMetric.GetType() {
		case mcs.Metric_GAUGE:
			metric.MType = model.Gauge
			value := pMetric.GetValue()
			metric.Value = &value
		case mcs.Metric_COUNTER:
			metric.MType = model.Counter
			delta := pMetric.GetDelta()
			metric.Delta = &delta
		}

		metrics = append(metrics, metric)
	}

	return metrics
}

// PingDB - проверка подключения к БД.
func (c *Collector) PingDB() error {
	return c.conn.Ping()
}
