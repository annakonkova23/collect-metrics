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
	"github.com/annakonkova23/collect-metrics/internal/config/db"
	"github.com/annakonkova23/collect-metrics/internal/model"
	fh "github.com/annakonkova23/collect-metrics/internal/service/fileHandler"
	"go.uber.org/zap"
)

var ErrorNotFound = errors.New("not exists name metric")

type Collector struct {
	MemStorage    *model.MemStorage
	logger        *zap.Logger
	fileHandler   *fh.FileHandler
	StoreInterval int
	mx            sync.Mutex
	conn          *sql.DB
}

type Metric struct {
	Name  string
	Type  string
	Value string
}

func NewCollector(ctx context.Context, cfg *config.ServerOptions, logger *zap.Logger, dbConnect *db.DBconnect) (*Collector, error) {
	clr := &Collector{
		StoreInterval: cfg.StoreInterval,
		logger:        logger,
		MemStorage:    model.NewMemStorage(),
	}
	dbConn, err := dbConnect.Connect()
	if err != nil {
		logger.Error("Ошибка открытия подключения к БД", zap.Error(err))
	} else {
		if err := dbConnect.Ping(dbConn); err != nil {
			logger.Error("Ошибка подключения к БД", zap.Error(err))
		} else {
			clr.conn = dbConn
		}
	}
	fileHandler, err := fh.NewFileHandler(cfg.FileStoragePath, logger)
	if err != nil {
		logger.Error("Ошибка создания обработчика файлов", zap.Error(err))
	} else {
		clr.fileHandler = fileHandler
	}
	if cfg.Restore {
		clr.logger.Info("Инициализация метрик", zap.Bool("Restore", cfg.Restore))
		if clr.conn != nil {
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
	var metrics []*model.Metrics
	var data sql.NullString
	rows, err := c.conn.QueryContext(ctx, `select json_strip_nulls(json_agg(json_build_object(
													'id',code,
													'type',type_metric,
													'value',gauge_value,
													'delta',counter_value)))
											from storage.metrics_value`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&data); err != nil {
			return err
		}
	}
	if data.Valid {
		if err := json.Unmarshal([]byte(data.String), &metrics); err != nil {
			msg := fmt.Sprintf("Ошибка парсинга метрик в структуру: %s", err.Error())
			return fmt.Errorf("%s", msg)
		}
	}
	if len(metrics) > 0 {
		c.MemStorage.InitMetrics(metrics)
	}
	return nil
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
	if c.conn != nil {
		c.SaveMetricToDB(metric)
	}
	if c.StoreInterval == 0 && c.fileHandler != nil {
		c.GetDataAndSaveToFile()
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

func (c *Collector) SaveMetricToDB(metric *model.Metrics) {
	res, err := c.conn.ExecContext(context.Background(), `
			insert into storage.metrics_value(code,type_metric, gauge_value, counter_value) 
			values ($1, $2, $3, $4) on conflict (code,type_metric) 
			do update set gauge_value = $5, counter_value = $6`, metric.ID, metric.MType, metric.Value, metric.Delta, metric.Value, metric.Delta)
	if err != nil {
		c.logger.Error("Ошибка сохранения метрики в БД", zap.Error(err))
	}
	if res != nil {
		rowsAffected, _ := res.RowsAffected()
		c.logger.Info("Метрика сохранена в БД", zap.Int64("количество обновленных строк", rowsAffected))
	}
}
