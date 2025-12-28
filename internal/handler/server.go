package handler

import (
	"context"
	"net/http"

	"github.com/annakonkova23/collect-metrics/internal/audit"
	"github.com/annakonkova23/collect-metrics/internal/config"
	mw "github.com/annakonkova23/collect-metrics/internal/handler/middleware"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	router    *chi.Mux
	url       string
	Collector *service.Collector
	Sugar     *zap.SugaredLogger
	logger    *zap.Logger
	key       string
	auditor   *audit.AuditManager
	srv       *http.Server
}

func NewServer(ctx context.Context, cfg *config.ServerOptions, logger *zap.Logger, collector *service.Collector) (*Server, error) {
	srv := &Server{
		router:    chi.NewRouter(),
		url:       cfg.Host,
		Collector: collector,
		Sugar:     logger.Sugar(),
		logger:    logger,
		key:       cfg.Key,
	}
	auditor, err := audit.NewAuditManager(logger, cfg.BufferSize, cfg.AuditFilePath, cfg.AuditURL)
	if err != nil {
		return nil, err
	}
	server := &http.Server{
		Addr:    cfg.Host,
		Handler: srv.router,
	}
	srv.srv = server
	srv.auditor = auditor
	srv.auditor.Start(ctx)
	return srv, nil
}

func (s *Server) StartAndListen() error {
	s.router.Use(mw.WithLogging, mw.WithCompress, mw.WithCheckHash(s.key))
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

func (s *Server) Shutdown(ctx context.Context) {
	if err := s.srv.Shutdown(ctx); err != nil {
		_ = s.srv.Close()
		s.logger.Error("Ошибка при закрытии сервера", zap.Error(err))
	}
}
