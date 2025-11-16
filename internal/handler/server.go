package handler

import (
	"context"
	"net/http"

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
}

func NewServer(ctx context.Context, cfg *config.ServerOptions, logger *zap.Logger, collector *service.Collector) (*Server, error) {
	return &Server{
		router:    chi.NewRouter(),
		url:       cfg.Host,
		Collector: collector,
		Sugar:     logger.Sugar(),
		logger:    logger,
		key:       cfg.Key,
	}, nil
}

func (s *Server) StartAndListen() error {
	mw.Key = s.key
	s.router.Use(mw.WithLogging, mw.WithCompress, mw.WithCheckHash)
	s.router.Post("/update/{type}/{name}/{value}", s.updateHandler)
	s.router.Post("/update/", s.updateJSONHandler)
	s.router.Get("/value/{type}/{name}", s.valueHandler)
	s.router.Get("/", s.allValuesHandler)
	s.router.Post("/value/", s.valueJSONHandler)
	s.router.Get("/ping", s.pingDBHandler)
	s.router.Post("/updates/", s.updateSeveralJSONHandler)
	if err := http.ListenAndServe(s.url, s.router); err != nil {
		return err
	}
	return nil
}
