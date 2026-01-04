// Package dbmanager отвечает за управление подключением к базе данных и выполнение миграций.
//
// Основные функции:
//   - Проверка доступности БД (Ping)
//   - Применение SQL-миграций с помощью goose
//   - Классификация PostgreSQL-ошибок для определения повторяемых (retriable) сбоев
//
// Пример использования:
//
//	db, err := sql.Open("pgx", dsn)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	dbm := dbmanager.NewDBManager(db)
//	if err := dbm.Ping(); err != nil {
//	    log.Fatal("БД недоступна", err)
//	}
//
//	if err := dbm.CreateObjectDB(); err != nil {
//	    log.Fatal("Ошибка миграций", err)
//	}
//
//	if dbm.IsTransportError(err) {
//	    // Повторить попытку подключения или операции
//	}
package dbmanager

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
)

const (
	// dirScript — путь к директории с SQL-миграциями.
	dirScript = "./migrations"
)

// DBManager инкапсулирует операции по управлению базой данных.
//
// Содержит:
//   - Подключение к БД (*sql.DB)
//   - Классификатор ошибок PostgreSQL для определения повторяемых сбоев
type DBManager struct {
	DB            *sql.DB                  // Подключение к базе данных
	classificater *PostgresErrorClassifier // Классификатор ошибок PostgreSQL
}

// NewDBManager создаёт новый экземпляр менеджера базы данных.
//
// Параметры:
//   - db: активное соединение с PostgreSQL (уже открытое)
//
// Возвращает указатель на *DBManager.
func NewDBManager(db *sql.DB) *DBManager {
	return &DBManager{
		DB:            db,
		classificater: NewPostgresErrorClassifier(),
	}
}

// Ping проверяет активность соединения с базой данных.

// Возвращает:
//   - nil, если соединение работает
//   - ошибку, если БД недоступна или произошёл сетевой сбой
func (d *DBManager) Ping() error {
	if err := d.DB.Ping(); err != nil {
		return err
	}
	return nil
}

// CreateObjectDB применяет миграции из директории ./migrations.

// Возвращает:
//   - nil, если миграции применены успешно
//   - ошибку с префиксом "Ошибка создания объектов БД", если goose.Up() завершился неудачно
func (d *DBManager) CreateObjectDB() error {
	if err := goose.Up(d.DB, dirScript); err != nil {
		msg := fmt.Sprintf("Ошибка создания объектов БД %s", err.Error())
		return errors.New(msg)
	}
	return nil
}

// IsTransportError определяет, является ли ошибка транспортной
// (например, временный сбой сети, разрыв соединения), которую можно повторить.
//
// Проверяет, является ли ошибка *pgconn.PgError и относится ли к категории Retriable
// (например, ошибка подключения, deadlock, timeout).
//
// Параметры:
//   - err: ошибка, полученная при выполнении SQL-запроса
//
// Возвращает:
//   - true, если ошибка повторяемая (retriable)
//   - false, если ошибка фатальная или не связана с сетью
func (d *DBManager) IsTransportError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if d.classificater.Classify(pgErr) == Retriable {
			return true
		}
	}
	return false
}
