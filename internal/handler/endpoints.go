package handler

import (
	"collect-metrics/internal/service"
	"log"
	"net/http"
)

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем путь запроса
	path := r.URL.Path // Например: "/update/v1/v2/v3"
	log.Printf("URL:%s", path)
	err := s.Collector.ParseAndSaveMetricsByUrl(path)
	if err != nil {
		if err.Error() == service.ERROR_NOT_FOUND {
			log.Println("Передаём ошибку 404")
			http.Error(w, service.ERROR_NOT_FOUND, http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	w.Header().Set("Content-Type", "Content-Type: text/plain")
	w.WriteHeader(http.StatusOK)
}
