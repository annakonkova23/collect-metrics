package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	fmt.Println("Строка подключения:", db.connString)
	sqlDB, err := sql.Open("pgx", db.connString)
	err = sqlDB.Ping()
	if err != nil {
		fmt.Println("Ошибка ping:", err.Error())
	}
	return sqlDB, err
}

func (db *DBconnect) Ping(sqlDB *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fmt.Println("Соединение с БД установлено", db.connString)
	err := sqlDB.PingContext(ctx)
	return err
}
