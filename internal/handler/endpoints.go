package handler

import (
	"fmt"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
)

type KeyValuePair struct {
	Key   string
	Value string
}

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {

	paramName := chi.URLParam(r, "name")
	paramType := chi.URLParam(r, "type")
	paramValue := chi.URLParam(r, "value")
	log.Printf("param:%s %s %s", paramName, paramType, paramValue)
	err := s.Collector.ParseAndSaveMetricsByParam(paramName, paramType, paramValue)
	if err != nil {
		if err.Error() == service.ErrorNotFound {
			log.Println("Передаём ошибку 404")
			http.Error(w, service.ErrorNotFound, http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "Content-Type: text/plain")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) valueHandler(w http.ResponseWriter, r *http.Request) {
	paramName := chi.URLParam(r, "name")
	paramType := chi.URLParam(r, "type")
	log.Printf("param:%s %s", paramName, paramType)
	value, ok, err := s.Collector.GetMetricValueByParam(paramName, paramType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "Метрика не найдена", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "Content-Type: text/plain")
	w.Write([]byte(value))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) allValuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		msg := fmt.Sprintf("Метод %s не поддерживается", r.Method)
		http.Error(w, msg, http.StatusBadRequest)
	}
	values := s.Collector.GetMetricAllValues()
	if len(values) == 0 {
		http.Error(w, "Нет значений", http.StatusBadRequest)
		return
	}
	GenerateHTMLTable(values, w)
}

// Функция для генерации HTML-таблицы
func GenerateHTMLTable(data map[string]string, w http.ResponseWriter) {
	// Преобразуем map в slice KeyValuePair для удобства
	var pairs []KeyValuePair
	for k, v := range data {
		pairs = append(pairs, KeyValuePair{Key: k, Value: v})
	}

	// Определяем шаблон HTML-таблицы
	const tableTemplate = `
        <table border="1" style="border-collapse: collapse; width: 50%; margin: 20px auto;">
            <thead>
                <tr>
                    <th style="padding: 8px; background-color: #f2f2f2;">Метрика</th>
                    <th style="padding: 8px; background-color: #f2f2f2;">Значение</th>
                </tr>
            </thead>
            <tbody>
                {{range .}}
                <tr>
                    <td style="padding: 6px; border: 1px solid #ddd;">{{.Key}}</td>
                    <td style="padding: 6px; border: 1px solid #ddd;">{{.Value}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    `

	tmpl, err := template.New("table").Parse(tableTemplate)
	if err != nil {
		http.Error(w, "Ошибка при создании шаблона", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, pairs); err != nil {
		http.Error(w, "Ошибка при рендеринге шаблона", http.StatusInternalServerError)
	}
}
