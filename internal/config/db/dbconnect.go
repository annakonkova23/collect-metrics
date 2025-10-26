package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	dirScript = "./migrations"
)

type DBconnect struct {
	connString string
}

func NewDbconnect(connString string) *DBconnect {
	return &DBconnect{
		connString: connString,
	}
}

func (db *DBconnect) Connect(createDB bool) (*sql.DB, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	sqlDB, err := sql.Open("pgx", db.connString)
	if createDB {
		if err := db.CreateObjectDB(sqlDB); err != nil {
			return nil, err
		}
	}
	return sqlDB, err
}

func (db *DBconnect) CreateObjectDB(sqlDB *sql.DB) error {
	if err := goose.Up(sqlDB, dirScript); err != nil {
		msg := fmt.Sprintf("Ошибка создания объектов БД %s", err.Error())
		return errors.New(msg)
	}
	return nil
}

func (db *DBconnect) Ping(sqlDB *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err := sqlDB.PingContext(ctx)
	return err
}
