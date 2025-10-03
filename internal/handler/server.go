package handler

import (
	"github.com/annakonkova23/collect-metrics/internal/service"
	"net/http"
)

type Server struct {
	mux       *http.ServeMux
	url       string
	Collector *service.Collector
}

func NewServer(url string) *Server {
	return &Server{
		mux:       http.NewServeMux(),
		url:       url,
		Collector: service.NewCollector(),
	}
}

func (s *Server) StartAndListen() error {
	s.mux.HandleFunc("/update/", s.updateHandler)
	if err := http.ListenAndServe(s.url, s.mux); err != nil {
		return err
	}
	return nil
}
