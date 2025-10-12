package handler

import (
	"github.com/annakonkova23/collect-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	router    *chi.Mux
	url       string
	Collector *service.Collector
	Sugar     *zap.SugaredLogger
	logger    *zap.Logger
}

func NewServer(url string, logger *zap.Logger) *Server {
	return &Server{
		router:    chi.NewRouter(),
		url:       url,
		Collector: service.NewCollector(logger),
		Sugar:     logger.Sugar(),
		logger:    logger,
	}
}

func (s *Server) StartAndListen() error {
	s.router.Post("/update/{type}/{name}/{value}", s.WithLogging(http.HandlerFunc(s.updateHandler)))
	s.router.Post("/update/", s.WithLogging(http.HandlerFunc(s.updateJsonHandler)))
	s.router.Get("/value/{type}/{name}", s.WithLogging(http.HandlerFunc(s.valueHandler)))
	s.router.Get("/", s.WithLogging(http.HandlerFunc(s.allValuesHandler)))
	s.router.Post("/value/", s.WithLogging(http.HandlerFunc(s.valueJsonHandler)))
	if err := http.ListenAndServe(s.url, s.router); err != nil {
		return err
	}
	return nil
}

// WithLogging добавляет дополнительный код для регистрации сведений о запросе
// и возвращает новый http.Handler.
func (s *Server) WithLogging(h http.Handler) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		// эндпоинт
		uri := r.RequestURI
		// метод запроса
		method := r.Method

		h.ServeHTTP(w, r) // обслуживание оригинального запроса

		duration := time.Since(start)

		// отправляем сведения о запросе в zap
		s.Sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
		)

	}
	// возвращаем функционально расширенный хендлер
	return logFn
}
