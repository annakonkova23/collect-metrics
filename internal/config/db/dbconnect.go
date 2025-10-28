package db

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	dirScript = "./migrations"
)

func NewDBConnect(connString string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

/*
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

func (db *DBconnect) isTransportError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if db.classificater.Classify(pgErr) == Retriable {
			return true
		}
	}
	return false
}
*/
