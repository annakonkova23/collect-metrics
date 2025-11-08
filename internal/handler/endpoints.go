package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/annakonkova23/collect-metrics/internal/model"
	"github.com/annakonkova23/collect-metrics/internal/service"
	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {

	paramName := chi.URLParam(r, "name")
	paramType := chi.URLParam(r, "type")
	paramValue := chi.URLParam(r, "value")
	s.logger.Debug(fmt.Sprintf("updateHandler param:%s %s %s", paramName, paramType, paramValue))
	err := s.Collector.SaveMetricsByParam(paramName, paramType, paramValue)
	if err != nil {
		if err == service.ErrorNotFound {
			s.logger.Debug("Передаём ошибку 404")
			http.Error(w, service.ErrorNotFound.Error(), http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		s.logger.Error("Неверный Content-Type")
		http.Error(w, "Неверный Content-Type", http.StatusBadRequest)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error(err.Error())
	}
	defer r.Body.Close()
	s.logger.Debug("updateJsonHandler BODY:" + string(bodyBytes))
	metric := &model.Metrics{}
	metric.UnmarshalJSON(bodyBytes)
	metric, err = s.Collector.SaveMetric(metric)
	if err != nil {
		s.logger.Error(err.Error(), zap.String("Body", string(bodyBytes)))
	}
	w.Header().Set("Content-Type", "application/json")
	value, _ := metric.MarshalJSON()
	w.WriteHeader(http.StatusOK)
	w.Write(value)

}

func (s *Server) valueHandler(w http.ResponseWriter, r *http.Request) {
	paramName := chi.URLParam(r, "name")
	paramType := chi.URLParam(r, "type")
	s.logger.Debug(fmt.Sprintf("param:%s %s", paramName, paramType))
	value, ok := s.Collector.GetMetricValueByParam(paramName, paramType)
	if !ok {
		http.Error(w, "Метрика не найдена", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))

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
func GenerateHTMLTable(data []*service.Metric, w http.ResponseWriter) {
	// Определяем шаблон HTML-таблицы
	const tableTemplate = `
        <table border="1" style="border-collapse: collapse; width: 50%; margin: 20px auto;">
            <thead>
                <tr>
                    <th style="padding: 8px; background-color: #f2f2f2;">Метрика</th>
					<th style="padding: 8px; background-color: #f2f2f2;">Тип</th>
                    <th style="padding: 8px; background-color: #f2f2f2;">Значение</th>
                </tr>
            </thead>
            <tbody>
                {{range .}}
                <tr>
                    <td style="padding: 6px; border: 1px solid #ddd;">{{.Name}}</td>
					<td style="padding: 6px; border: 1px solid #ddd;">{{.Type}}</td>
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
	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка при рендеринге шаблона", http.StatusInternalServerError)
	}
}

func (s *Server) valueJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		s.logger.Error("Неверный Content-Type")
		http.Error(w, "Неверный Content-Type", http.StatusBadRequest)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	defer r.Body.Close()
	s.logger.Debug("valueJSONHandler BODY:" + string(bodyBytes))
	metric := &model.Metrics{}
	metric.UnmarshalJSON(bodyBytes)
	value, err := s.Collector.GetMetricJSON(metric.ID, metric.MType)
	if err != nil {
		if err == service.ErrorNotFound {
			s.logger.Debug("Передаём ошибку 404")
			http.Error(w, service.ErrorNotFound.Error(), http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))

}

func (s *Server) pingDBHandler(w http.ResponseWriter, r *http.Request) {
	dbconn, err := s.DB.Connect(false)
	if err != nil {
		s.logger.Error("pingDB:" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.DB.Ping(dbconn)
	if err != nil {
		s.logger.Error("pingDB:" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
