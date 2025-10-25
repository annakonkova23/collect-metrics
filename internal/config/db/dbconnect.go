package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBconnect struct {
	connString string
}

func NewDbconnect(connString string) *DBconnect {
	return &DBconnect{
		connString: connString,
	}
}

func (db *DBconnect) Connect() (*sql.DB, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	sqlDb, err := sql.Open("pgx", db.connString)
	return sqlDb, err
}

func (db *DBconnect) Ping(sqlDb *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err := sqlDb.PingContext(ctx)
	return err
}
