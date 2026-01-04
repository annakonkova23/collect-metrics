package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/config/db"
	"github.com/annakonkova23/collect-metrics/internal/handler"
	"github.com/annakonkova23/collect-metrics/internal/model"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"go.uber.org/zap"
)

// ExampleServer_StartAndListen показывает, как запустить сервер
// и использовать его эндпоинты для передачи и получения метрик.
func ExampleServer_StartAndListen() {
	// Настройка логгера
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Конфигурация сервера
	cfg := &config.ServerOptions{
		Host: "localhost:8080",
		Key:  "", // HMAC-проверка отключена
	}

	// Создаём сервер
	ctx := context.Background()

	dbConnect, err := db.NewDBConnect(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("Не удалось подключиться к БД", zap.Error(err))
	} else {
		defer dbConnect.Close()
	}

	// Создаём Collector
	collector, err := service.NewCollector(ctx, cfg, logger, dbConnect)
	if err != nil {
		logger.Error("Не удалось инициализировать Collector", zap.Error(err))
	}

	server, err := handler.NewServer(ctx, cfg, logger, collector)
	if err != nil {
		fmt.Printf("Ошибка инициализации сервера: %v\n", err)
		return
	}

	// Запускаем в отдельной горутине
	go func() {
		if err := server.StartAndListen(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Ошибка запуска сервера: %v\n", err)
		}
	}()

	// Даем серверу время запуститься
	time.Sleep(100 * time.Millisecond)

	// === 1. Обновление метрики через путь (gauge) ===
	resp, err := http.Post("http://localhost:8080/update/gauge/cpu_usage/3.14", "", nil)
	if err != nil {
		fmt.Printf("Ошибка запроса /update: %v\n", err)
	} else {
		fmt.Printf("Update (gauge): %s\n", resp.Status)
		_ = resp.Body.Close()
	}

	// === 2. Получение значения через путь ===
	resp, err = http.Get("http://localhost:8080/value/gauge/cpu_usage")
	if err != nil {
		fmt.Printf("Ошибка запроса /value: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Value (gauge): %s\n", string(body))
		_ = resp.Body.Close()
	}

	// === 3. Обновление метрики в формате JSON ===
	metric := &model.Metrics{
		ID:    "poll_count",
		MType: "counter",
		Delta: ptrInt64(42),
	}
	jsonBody, _ := json.Marshal(metric)
	resp, err = http.Post("http://localhost:8080/update/", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Ошибка JSON update: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Update JSON: %s, Response body: %s\n", resp.Status, string(body))
		_ = resp.Body.Close()
	}

	// === 4. Получение метрики в формате JSON ===
	jsonBody, _ = json.Marshal(metric)
	resp, err = http.Post("http://localhost:8080/value/", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Ошибка JSON value: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Value JSON: %s\n", string(body))
		_ = resp.Body.Close()
	}

	// === 5. Обновление нескольких метрик за раз ===
	metrics := []*model.Metrics{
		{ID: "temp", MType: "gauge", Value: ptrFloat64(25.5)},
		{ID: "requests", MType: "counter", Delta: ptrInt64(100)},
	}
	jsonBody, _ = json.Marshal(metrics)
	resp, err = http.Post("http://localhost:8080/updates/", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Ошибка updates: %v\n", err)
	} else {
		fmt.Printf("Updates: %s\n", resp.Status)
		_ = resp.Body.Close()
	}

	// === 6. Просмотр всех метрик (HTML) ===
	resp, err = http.Get("http://localhost:8080/")
	if err != nil {
		fmt.Printf("Ошибка /: %v\n", err)
	} else {
		fmt.Printf("All values (HTML): %s, Length: %d\n", resp.Status, resp.ContentLength)
		_ = resp.Body.Close()
	}

	// === 7. Проверка доступности БД (если настроена) ===
	resp, err = http.Get("http://localhost:8080/ping")
	if err != nil {
		fmt.Printf("Ошибка /ping: %v\n", err)
	} else {
		fmt.Printf("Ping DB: %s\n", resp.Status)
		_ = resp.Body.Close()
	}

}

// Вспомогательные функции
func ptrInt64(v int64) *int64 {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}
