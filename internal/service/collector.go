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
	dbmanager "github.com/annakonkova23/collect-metrics/internal/service/dbManager"
	fh "github.com/annakonkova23/collect-metrics/internal/service/fileHandler"
	"go.uber.org/zap"
)

const (
	countAttempt = 3
	delayAttempt = 2
)

var ErrorNotFound = errors.New("not exists name metric")

type Collector struct {
	MemStorage    *model.MemStorage
	logger        *zap.Logger
	fileHandler   *fh.FileHandler
	StoreInterval int
	mx            sync.Mutex
	conn          *dbmanager.DBManager
}

type Metric struct {
	Name  string
	Type  string
	Value string
}

func NewCollector(ctx context.Context, cfg *config.ServerOptions, logger *zap.Logger, dbConnect *sql.DB) (*Collector, error) {
	clr := &Collector{
		StoreInterval: cfg.StoreInterval,
		logger:        logger,
		MemStorage:    model.NewMemStorage(),
	}
	clr.conn = dbmanager.NewDBManager(dbConnect)
	fileHandler, err := fh.NewFileHandler(cfg.FileStoragePath, logger)
	if err != nil {
		logger.Error("Ошибка создания обработчика файлов", zap.Error(err))
	} else {
		clr.fileHandler = fileHandler
	}
	if clr.conn.DB != nil {
		err := clr.conn.CreateObjectDB()
		if err != nil {
			return nil, err
		}
	}
	if cfg.Restore {
		clr.logger.Info("Инициализация метрик", zap.Bool("Restore", cfg.Restore))
		if clr.conn.DB != nil {
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

func (c *Collector) initMemStorageFromDB(ctx context.Context) error {
	metrics, err := c.RetryLoadMetricsToDB(ctx)
	if err != nil {
		return err
	}
	if len(metrics) > 0 {
		c.MemStorage.InitMetrics(metrics)
	}
	return nil
}

func (c *Collector) loadMetricsFromDB(ctx context.Context) ([]*model.Metrics, error) {
	var metrics []*model.Metrics
	var data sql.NullString
	rows, err := c.conn.DB.QueryContext(ctx, `select json_strip_nulls(json_agg(json_build_object(
													'id',code,
													'type',type_metric,
													'value',gauge_value,
													'delta',counter_value)))
											from storage.metrics_value`)
	if err != nil {
		c.logger.Error("Ошибка запроса метрик из БД", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&data); err != nil {
			c.logger.Error("Ошибка сканирования данных", zap.Error(err))
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		c.logger.Error("Ошибка чтения строк", zap.Error(err))
		return nil, err
	}
	if data.Valid {
		if err := json.Unmarshal([]byte(data.String), &metrics); err != nil {
			msg := fmt.Sprintf("Ошибка парсинга метрик в структуру: %s", err.Error())
			return nil, fmt.Errorf("%s", msg)
		}
	}
	return metrics, nil
}

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
	if c.conn != nil {
		err := c.saveMetricsToDB(ctx, metricsNew)
		if err != nil {
			c.logger.Error("Ошибка сохранения метрик в БД", zap.Error(err))
		}
	}
	if c.StoreInterval == 0 && c.fileHandler != nil {
		c.GetDataAndSaveToFile()
	}
	return metrics, nil

}

func (c *Collector) saveMetricsToDB(ctx context.Context, metrics []*model.Metrics) error {
	tx, err := c.conn.DB.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}

	query := `
        INSERT INTO storage.metrics_value(code, type_metric, gauge_value, counter_value)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (code, type_metric)
        DO UPDATE SET
            gauge_value = EXCLUDED.gauge_value,
            counter_value = EXCLUDED.counter_value;
    `

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	for _, metric := range metrics {
		_, err := stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Value, metric.Delta)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка выполнения запроса для метрики %s: %w", metric.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}

func (c *Collector) PingDB() error {
	return c.conn.Ping()
}

func (c *Collector) RetrySaveMetricsToDB(ctx context.Context, metrics []*model.Metrics) error {
	delay := 1

	for i := 0; i < countAttempt; i++ {

		err := c.saveMetricsToDB(ctx, metrics)
		if err != nil {
			if c.conn.IsTransportError(err) {
				select {
				case <-ctx.Done():
					c.logger.Info("Отмена контекста")
					return err
				case <-time.After(time.Duration(delay) * time.Second):
					delay = delay + delayAttempt
					continue
				}
			} else {
				return err
			}
		}
		return nil
	}
	return nil
}

func (c *Collector) RetryLoadMetricsToDB(ctx context.Context) ([]*model.Metrics, error) {
	delay := 1
	var metrics []*model.Metrics
	var err error
	for i := 0; i < countAttempt; i++ {

		metrics, err = c.loadMetricsFromDB(ctx)
		if err != nil {
			if c.conn.IsTransportError(err) {
				select {
				case <-ctx.Done():
					c.logger.Info("Отмена контекста")
					return nil, err
				case <-time.After(time.Duration(delay) * time.Second):
					delay = delay + delayAttempt
					continue
				}
			} else {
				return nil, err
			}
		}
		return metrics, nil
	}
	return metrics, err
}
