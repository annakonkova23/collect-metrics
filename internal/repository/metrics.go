package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/model"
	"go.uber.org/zap"
)

const (
	countAttempt = 3
	delayAttempt = 2
)

func (s *DBStore) RetryLoadMetricsToDB(ctx context.Context, delay int) ([]*model.Metrics, error) {
	var metrics []*model.Metrics
	var err error
	for i := 0; i < countAttempt; i++ {

		metrics, err = s.loadMetricsFromDB(ctx)
		if err != nil {
			if s.database.IsTransportError(err) {
				select {
				case <-ctx.Done():
					s.logger.Info("Отмена контекста")
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

func (s *DBStore) SaveMetricsToDB(ctx context.Context, metrics []*model.Metrics) error {
	tx, err := s.database.DB.Begin()
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

func (s *DBStore) loadMetricsFromDB(ctx context.Context) ([]*model.Metrics, error) {
	var metrics []*model.Metrics
	var data sql.NullString
	rows, err := s.database.DB.QueryContext(ctx, `select json_strip_nulls(json_agg(json_build_object(
													'id',code,
													'type',type_metric,
													'value',gauge_value,
													'delta',counter_value)))
											from storage.metrics_value`)
	if err != nil {
		s.logger.Error("Ошибка запроса метрик из БД", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&data); err != nil {
			s.logger.Error("Ошибка сканирования данных", zap.Error(err))
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		s.logger.Error("Ошибка чтения строк", zap.Error(err))
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
