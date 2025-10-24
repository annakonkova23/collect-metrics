package handler

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"
)

type CustomResponseWriter struct {
	w           http.ResponseWriter
	header      http.Header
	buf         *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

func NewCustomResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		w:          w,
		header:     make(http.Header),
		buf:        &bytes.Buffer{},
		statusCode: http.StatusOK,
	}
}

// Возвращает внутренний заголовок (временное хранилище)
func (crw *CustomResponseWriter) Header() http.Header {
	return crw.header
}

// Сохраняет статус-код, но не пишет в оригинальный writer
func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	if crw.wroteHeader {
		return
	}
	crw.statusCode = statusCode
	crw.wroteHeader = true
}

// Сохраняет тело в буфер
func (crw *CustomResponseWriter) Write(b []byte) (int, error) {
	return crw.buf.Write(b)
}

func (s *Server) WithLogging(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		newReq := r.Clone(r.Context())

		if r.Header.Get("Content-Encoding") == "gzip" &&
			(r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Content-Type") == "text/html") {

			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Ошибка декодирования gzip", http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			decompressedBody, err := io.ReadAll(gz)
			if err != nil {
				http.Error(w, "Ошибка чтения разжатого тела", http.StatusInternalServerError)
				return
			}

			newReq.Body = io.NopCloser(bytes.NewBuffer(decompressedBody))
			newReq.ContentLength = int64(len(decompressedBody))
			newReq.Header.Del("Content-Encoding")
		}

		crw := NewCustomResponseWriter(w)

		h.ServeHTTP(crw, newReq)

		if crw.header.Get("Content-Type") == "" {
			crw.header.Set("Content-Type", "text/plain")
		}

		contentType := crw.header.Get("Content-Type")

		shouldGzip := false
		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") &&
			(contentType == "application/json" || contentType == "text/html") {
			shouldGzip = true

		}

		for k, vv := range crw.header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		if shouldGzip {
			w.Header().Set("Content-Encoding", "gzip")
		}

		w.WriteHeader(crw.statusCode)

		if shouldGzip {
			var gzipBuf bytes.Buffer
			gz := gzip.NewWriter(&gzipBuf)
			_, err := gz.Write(crw.buf.Bytes())
			if err != nil {
				s.Sugar.Error("Ошибка сжатия:", err)
				return
			}
			if err := gz.Close(); err != nil {
				s.Sugar.Error("Ошибка закрытия gzip:", err)
				return
			}
			if _, err := w.Write(gzipBuf.Bytes()); err != nil {
				s.Sugar.Error("Ошибка записи сжатого тела:", err)
				return
			}
		} else {
			if _, err := w.Write(crw.buf.Bytes()); err != nil {
				s.Sugar.Error("Ошибка записи тела:", err)
				return
			}
		}

		duration := time.Since(start)
		s.Sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"shouldGzip", shouldGzip,
			"contentType", contentType,
		)
	}
}
