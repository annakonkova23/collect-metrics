package dbworker

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
)

const (
	dirScript = "./migrations"
)

type DBWorker struct {
	DB            *sql.DB
	classificater *PostgresErrorClassifier
}

func NewDBWorker(db *sql.DB) *DBWorker {
	return &DBWorker{
		DB:            db,
		classificater: NewPostgresErrorClassifier(),
	}
}

func (d *DBWorker) Ping() error {
	if err := d.DB.Ping(); err != nil {
		return err
	}
	return nil
}

func (d *DBWorker) CreateObjectDB() error {
	if err := goose.Up(d.DB, dirScript); err != nil {
		msg := fmt.Sprintf("Ошибка создания объектов БД %s", err.Error())
		return errors.New(msg)
	}
	return nil
}

func (d *DBWorker) IsTransportError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if d.classificater.Classify(pgErr) == Retriable {
			return true
		}
	}
	return false
}
