package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// WithLogging возвращает middleware, которое логирует информацию о каждом HTTP-запросе.
//
// Логируемые данные:
//   - HTTP-метод (GET, POST и т.д.)
//   - URI запроса
//   - Время обработки запроса (duration)
//
// Параметры:
//   - logger: экземпляр *zap.Logger, используемый для записи логов.
//     Должен быть сконфигурирован заранее. Рекомендуется использовать один экземпляр на всё приложение.
//
// Пример использования:
//
//	logger, _ := zap.NewProduction()
//	defer logger.Sync()
//
//	r := chi.NewRouter()
//	r.Use(middleware.WithLogging(logger))
//	r.Get("/", handler)
func WithLogging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			h.ServeHTTP(w, r)

			logger.Sugar().Infoln("uri", uri, "method", method, "duration", time.Since(start))
		})
	}
}
