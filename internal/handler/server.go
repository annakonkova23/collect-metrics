package handler

import (
	"github.com/annakonkova23/collect-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Server struct {
	router    *chi.Mux
	url       string
	Collector *service.Collector
}

func NewServer(url string) *Server {
	return &Server{
		router:    chi.NewRouter(),
		url:       url,
		Collector: service.NewCollector(),
	}
}

func (s *Server) StartAndListen() error {
	s.router.Post("/update/{type}/{name}/{value}", s.updateHandler)
	s.router.Get("/value/{type}/{name}", s.valueHandler)
	s.router.Get("/", s.allValuesHandler)
	if err := http.ListenAndServe(s.url, s.router); err != nil {
		return err
	}
	return nil
}
