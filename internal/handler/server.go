// Package handler отвечает за HTTP-интерфейс сервера сбора и отображения метрик.
//
// Реализует:
//   - Приём метрик через POST-запросы (в формате JSON и path-параметров)
//   - Получение значений метрик по имени и типу
//   - Просмотр всех метрик через веб-интерфейс
//   - Проверку подлинности запросов (HMAC-SHA256)
//   - Сжатие/распаковку тел запросов и ответов (gzip)
//   - Логирование запросов
//   - Аудит операций (запись в файл и/или отправка на удалённый URL)
//
// Использует:
//   - go-chi/chi — для маршрутизации
//   - zap — для логирования
//   - audit — для аудита
//   - service.Collector — для хранения и обработки метрик
//
// Пример инициализации:
//
//	ctx := context.Background()
//	cfg := config.NewServerOptions()
//	logger, _ := zap.NewProduction()
//	collector := service.NewCollector()
//
//	server, err := handler.NewServer(ctx, cfg, logger, collector)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer server.Shutdown(ctx)
//
//	if err := server.StartAndListen(); err != nil && err != http.ErrServerClosed {
//	    log.Fatal(err)
//	}
package handler

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net"
	"net/http"

	"github.com/annakonkova23/collect-metrics/internal/audit"
	"github.com/annakonkova23/collect-metrics/internal/config"
	"github.com/annakonkova23/collect-metrics/internal/crypto"
	mw "github.com/annakonkova23/collect-metrics/internal/handler/middleware"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Server инкапсулирует HTTP-сервер, маршрутизатор, обработчики и зависимости.
//
// Обеспечивает:
//   - Регистрацию middleware и маршрутов
//   - Запуск и корректное завершение работы сервера
//   - Интеграцию с аудитом и валидацией запросов
type Server struct {
	router        *chi.Mux            // HTTP-роутер на основе chi
	url           string              // Адрес, на котором работает сервер (например, ":8080")
	Collector     *service.Collector  // Служба сбора и хранения метрик
	Sugar         *zap.SugaredLogger  // Удобный интерфейс логирования
	logger        *zap.Logger         // Структурированный логгер
	key           string              // Ключ для проверки HMAC-SHA256 в заголовке HashSHA256
	auditor       *audit.AuditManager // Менеджер аудита (файл + HTTP)
	srv           *http.Server        // Стандартный HTTP-сервер Go
	keyCrypt      *rsa.PrivateKey     // Ключ для шифрования
	TrustedSubnet string              // Доверенная подсеть
}

// NewServer создаёт новый экземпляр HTTP-сервера.
//
// Параметры:
//   - ctx: контекст для управления жизненным циклом аудит-менеджера
//   - cfg: конфигурация сервера (хост, ключ, размер буфера, пути аудита)
//   - logger: экземпляр *zap.Logger для логирования
//   - collector: сервис, управляющий метриками
//
// Действия:
//   - Инициализирует роутер chi
//   - Создаёт http.Server с обработчиком
//
// Возвращает:
//   - Указатель на *Server и nil в случае успеха
//   - nil и ошибку, если не удалось инициализировать аудит (например, нет прав на файл)
//
// Примечание: сервер не запускается автоматически — требуется вызвать StartAndListen().
func NewServer(ctx context.Context, cfg *config.ServerOptions, logger *zap.Logger, collector *service.Collector) (*Server, error) {
	r := chi.NewRouter()

	server := &http.Server{
		Addr:    cfg.Host,
		Handler: r,
	}

	auditor, err := audit.NewAuditManager(logger, cfg.BufferSize, cfg.AuditFilePath, cfg.AuditURL)
	if err != nil {
		return nil, err
	}

	key, err := crypto.ReadPrivateKey(cfg.KeyPath)
	if err != nil {
		logger.Error("Ошибка чтения приватного ключа", zap.Error(err))
	}

	return &Server{
		router:        r,
		url:           cfg.Host,
		Collector:     collector,
		Sugar:         logger.Sugar(),
		logger:        logger,
		key:           cfg.Key,
		srv:           server,
		auditor:       auditor,
		keyCrypt:      key,
		TrustedSubnet: cfg.TrustedSubnet,
	}, nil
}

// StartAndListen запускает HTTP-сервер и начинает прослушивать подключения.
//
// Регистрирует:
//   - Middleware: логирование, сжатие, проверка хэша
//   - Обработчики для всех эндпоинтов
//
// Эндпоинты:
//   - POST   /update/{type}/{name}/{value} — обновить метрику по пути
//   - POST   /update/ — обновить метрику в формате JSON
//   - POST   /updates/ — обновить несколько метрик (массив JSON)
//   - GET    /value/{type}/{name} — получить значение метрики
//   - GET    /value/ — получить значение метрики в формате JSON
//   - GET    / — отобразить все метрики (HTML)
//   - GET    /ping — проверка доступности БД
//
// В случае ошибки (кроме ErrServerClosed) возвращает ошибку.
// Для graceful shutdown используйте Shutdown(ctx).
func (s *Server) StartAndListen(ctx context.Context) error {
	s.auditor.Start(ctx)
	s.router.Use(mw.WithLogging(s.logger), mw.WithCompress, mw.WithDecrypt(s.keyCrypt), mw.WithCheckHash(s.key))
	s.router.Post("/update/{type}/{name}/{value}", s.updateHandler)
	s.router.Post("/update/", s.updateJSONHandler)
	s.router.Get("/value/{type}/{name}", s.valueHandler)
	s.router.Get("/", s.allValuesHandler)
	s.router.Post("/value/", s.valueJSONHandler)
	s.router.Get("/ping", s.pingDBHandler)
	s.router.Post("/updates/", s.updateSeveralJSONHandler)
	if err := s.srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

// Shutdown корректно останавливает сервер.
//
// Параметры:
//   - ctx: контекст с таймаутом для graceful shutdown
//
// Поведение:
//   - Вызывает http.Server.Shutdown(ctx)
//   - В случае ошибки — логирует и пытается закрыть сервер принудительно
//   - Использует s.logger для записи ошибок
//
// Должен вызываться при завершении приложения, например через defer.
func (s *Server) Shutdown(ctx context.Context) {
	if err := s.srv.Shutdown(ctx); err != nil {
		errClose := s.srv.Close()
		s.logger.Error("Ошибка при закрытии сервера", zap.Error(err), zap.Error(errClose))
	}
}

// IsIPInCIDR проверяет, входит ли IP в CIDR-подсеть.
func (s *Server) IsIPInCIDR(ipStr string) (bool, error) {

	_, cidr, err := net.ParseCIDR(s.TrustedSubnet)
	if err != nil {
		return false, fmt.Errorf("невалидный CIDR '%s': %w", s.TrustedSubnet, err)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("невалидный IP '%s'", ipStr)
	}

	return cidr.Contains(ip), nil
}
