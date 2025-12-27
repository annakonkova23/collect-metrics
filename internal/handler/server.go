package handler

import (
	"context"
	"net/http"

	"github.com/annakonkova23/collect-metrics/internal/config"
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
	s.router.Post("/update/{type}/{name}/{value}", s.WithLoggingAndCompress(http.HandlerFunc(s.updateHandler)))
	s.router.Post("/update/", s.WithLoggingAndCompress(http.HandlerFunc(s.updateJSONHandler)))
	s.router.Get("/value/{type}/{name}", s.WithLoggingAndCompress(http.HandlerFunc(s.valueHandler)))
	s.router.Get("/", s.WithLoggingAndCompress(http.HandlerFunc(s.allValuesHandler)))
	s.router.Post("/value/", s.WithLoggingAndCompress(http.HandlerFunc(s.valueJSONHandler)))
	s.router.Get("/ping", s.WithLoggingAndCompress(http.HandlerFunc(s.pingDBHandler)))
	s.router.Post("/updates/", s.WithLoggingAndCompress(http.HandlerFunc(s.updateSeveralJSONHandler)))
	if err := http.ListenAndServe(s.url, s.router); err != nil {
		return err
	}
	return nil
}
