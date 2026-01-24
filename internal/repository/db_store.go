// Package repository отвечает за работу с базой данных.
package repository

import (
	"database/sql"

	dbmanager "github.com/annakonkova23/collect-metrics/internal/repository/dbmanager"
	"go.uber.org/zap"
)

type DBStore struct {
	database *dbmanager.DBManager
	logger   *zap.Logger
}

func NewDBStore(logger *zap.Logger, dbConnect *sql.DB) *DBStore {
	return &DBStore{
		database: dbmanager.NewDBManager(dbConnect),
		logger:   logger,
	}
}

func (s *DBStore) IsConnectDB() bool {
	return s.database.DB != nil
}

func (s *DBStore) CreateObjectDB() error {
	return s.database.CreateObjectDB()
}

func (s *DBStore) Ping() error {
	return s.database.Ping()
}
