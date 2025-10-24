package handler

import (
	"bytes"
	"compress/gzip"
	"fmt"
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

func (crw *CustomResponseWriter) Write(b []byte) (int, error) {
	return crw.buf.Write(b)
}

func (s *Server) WithLoggingAndCompress(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		newReq, err := s.decodeRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		crw := NewCustomResponseWriter(w)

		h.ServeHTTP(crw, newReq)

		acceptEncoding := r.Header.Get("Accept-Encoding")

		header, statusCode, shouldGzip, body, err := s.codeResponse(acceptEncoding, crw)
		if err != nil {
			s.Sugar.Error("Ошибка сжатия ответа:", err)
			return
		}
		for k, vv := range header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		if shouldGzip {
			w.Header().Set("Content-Encoding", "gzip")
		}

		w.WriteHeader(statusCode)

		if _, err := w.Write(body); err != nil {
			s.Sugar.Error("Ошибка записи тела:", err)
			return
		}

		duration := time.Since(start)
		s.Sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"shouldGzip", shouldGzip,
		)
	}
}

func (s *Server) decodeRequest(r *http.Request) (*http.Request, error) {

	if r.Header.Get("Content-Encoding") == "gzip" &&
		(r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Content-Type") == "text/html") {
		newReq := r.Clone(r.Context())
		decompressBody, err := s.decompessRequest(r)
		if err != nil {
			return nil, err
		}
		newReq.Body = io.NopCloser(bytes.NewBuffer(decompressBody))
		newReq.ContentLength = int64(len(decompressBody))
		newReq.Header.Del("Content-Encoding")
		return newReq, nil
	} else {
		return r, nil
	}

}

func (s *Server) codeResponse(acceptEncoding string, crw *CustomResponseWriter) (http.Header, int, bool, []byte, error) {
	contentType := crw.header.Get("Content-Type")
	shouldGzip := false
	if strings.Contains(acceptEncoding, "gzip") &&
		(contentType == "application/json" || contentType == "text/html") {
		shouldGzip = true
	}

	var body []byte
	if shouldGzip {
		compressBody, err := s.compessResponse(crw)
		if err != nil {
			return nil, 0, false, nil, err
		}
		body = compressBody
	} else {
		body = crw.buf.Bytes()
	}
	return crw.header, crw.statusCode, shouldGzip, body, nil
}

func (s *Server) compessResponse(crw *CustomResponseWriter) ([]byte, error) {
	var gzipBuf bytes.Buffer
	gz := gzip.NewWriter(&gzipBuf)
	_, err := gz.Write(crw.buf.Bytes())
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return gzipBuf.Bytes(), nil
}

func (s *Server) decompessRequest(r *http.Request) ([]byte, error) {
	gz, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, fmt.Errorf("%s", "Ошибка декодирования gzip")
	}
	defer gz.Close()

	decompressedBody, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("%s", "Ошибка чтения разжатого тела")
	}

	return decompressedBody, nil
}
