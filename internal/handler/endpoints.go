package handler

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем путь запроса
	path := r.URL.Path // Например: "/update/v1/v2/v3"
	var typeMetric, nameMetric, value string
	// Разбиваем путь на части
	parts := strings.Split(path, "/")
	// parts = ["", "update", "v1", "v2", "v3"]
	//fmt.Fprintf(w, "count %d v1 %s v2 %s v3 %s \n", len(parts), parts[0], parts[1], parts[2])
	// Проверяем, что маршрут начинается с "/update"
	if len(parts) < 2 || parts[1] != "update" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	if len(parts) == 3 {
		http.Error(w, "Not exists name metric", http.StatusNotFound)
		return
	}
	if len(parts) >= 4 {
		typeMetric = parts[2]
		nameMetric = parts[3]
		if len(parts) >= 5 {
			value = parts[4]
		}
	}
	err := s.MemStorage.SetMetric(nameMetric, typeMetric, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "Content-Type: text/plain")
	fmt.Fprintf(w, "result: \n %s ", s.MemStorage.String())
}
