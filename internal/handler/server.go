package handler

import (
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
}

func NewServer(cfg *config.ServerOptions, logger *zap.Logger) (*Server, error) {
	collector, err := service.NewCollector(cfg, logger)
	if err != nil {
		return nil, err
	}
	return &Server{
		router:    chi.NewRouter(),
		url:       cfg.Host,
		Collector: collector,
		Sugar:     logger.Sugar(),
		logger:    logger,
	}, nil
}

func (s *Server) StartAndListen() error {
	s.router.Post("/update/{type}/{name}/{value}", s.WithLogging(http.HandlerFunc(s.updateHandler)))
	s.router.Post("/update/", s.WithLogging(http.HandlerFunc(s.updateJSONHandler)))
	s.router.Get("/value/{type}/{name}", s.WithLogging(http.HandlerFunc(s.valueHandler)))
	s.router.Get("/", s.WithLogging(http.HandlerFunc(s.allValuesHandler)))
	s.router.Post("/value/", s.WithLogging(http.HandlerFunc(s.valueJSONHandler)))
	if err := http.ListenAndServe(s.url, s.router); err != nil {
		return err
	}
	return nil
}
