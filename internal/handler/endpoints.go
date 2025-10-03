package handler

import (
	"github.com/annakonkova23/collect-metrics/internal/service"
	"log"
	"net/http"
)

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем путь запроса
	path := r.URL.Path
	log.Printf("URL:%s", path)
	err := s.Collector.ParseAndSaveMetricsByURL(path)
	if err != nil {
		if err.Error() == service.ErrorNotFound {
			log.Println("Передаём ошибку 404")
			http.Error(w, service.ErrorNotFound, http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	w.Header().Set("Content-Type", "Content-Type: text/plain")
	w.WriteHeader(http.StatusOK)
}
