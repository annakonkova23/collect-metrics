package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		logger, err := zap.NewDevelopment()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer logger.Sync()

		h.ServeHTTP(w, r)

		duration := time.Since(start)

		logger.Sugar().Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
		)
	})
}
