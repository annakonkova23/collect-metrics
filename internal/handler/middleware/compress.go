package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	gzipWriterPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
	gzipBufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

type CustomResponseWriter struct {
	w           http.ResponseWriter
	header      http.Header
	buf         *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

// NewCustomResponseWriter Создает новый экземпляр CustomResponseWriter.
func NewCustomResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		w:          w,
		header:     make(http.Header),
		buf:        &bytes.Buffer{},
		statusCode: http.StatusOK,
	}
}

// Header Возвращает внутренний заголовок (временное хранилище).
func (crw *CustomResponseWriter) Header() http.Header {
	return crw.header
}

// WriteHeader Сохраняет статус-код.
func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	if crw.wroteHeader {
		return
	}
	crw.statusCode = statusCode
	crw.wroteHeader = true
}

// Write Пишет в буфер.
func (crw *CustomResponseWriter) Write(b []byte) (int, error) {
	return crw.buf.Write(b)
}

// WithCompress — middleware для автоматического сжатия ответов и распаковки запросов с использованием gzip.
//
// Обрабатывает входящие запросы:
//   - Если заголовок "Content-Encoding: gzip" присутствует, тело запроса автоматически распаковывается.
//   - Поддерживает типы: application/json, text/html.
//
// Формирует исходящие ответы:
//   - Если клиент отправил "Accept-Encoding: gzip" и Content-Type поддерживается,
//     ответ сжимается с помощью gzip и добавляется заголовок "Content-Encoding: gzip".
//
// Middleware прозрачно интегрируется в цепочку обработки HTTP-запросов.
// Использует кастомный ResponseWriter (CustomResponseWriter) для перехвата тела ответа до отправки.
//
// Пример применения:
//
//	r := chi.NewRouter()
//	r.Use(middleware.WithCompress)
//	r.Post("/update", updateHandler)
//
// Важно:
//   - Не модифицирует поведение для методов, не поддерживаемых по Content-Type.
//   - Ошибки декодирования или сжатия возвращаются с кодом 400.
//
// Возвращает http.Handler, оборачивающий исходный обработчик h.
func WithCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		newReq, err := decodeRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		crw := NewCustomResponseWriter(w)

		h.ServeHTTP(crw, newReq)

		acceptEncoding := r.Header.Get("Accept-Encoding")

		header, statusCode, shouldGzip, body, err := codeResponse(acceptEncoding, crw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	})
}

func decodeRequest(r *http.Request) (*http.Request, error) {

	if r.Header.Get("Content-Encoding") == "gzip" &&
		(r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Content-Type") == "text/html") {
		newReq := r.Clone(r.Context())
		decompressBody, err := decompessRequest(r)
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

func codeResponse(acceptEncoding string, crw *CustomResponseWriter) (http.Header, int, bool, []byte, error) {
	contentType := crw.header.Get("Content-Type")
	shouldGzip := false
	if strings.Contains(acceptEncoding, "gzip") &&
		(contentType == "application/json" || contentType == "text/html") {
		shouldGzip = true
	}

	var body []byte
	if shouldGzip {
		compressBody, err := compessResponse(crw)
		if err != nil {
			return nil, 0, false, nil, err
		}
		body = compressBody
	} else {
		body = crw.buf.Bytes()
	}
	return crw.header, crw.statusCode, shouldGzip, body, nil
}

func compessResponse(crw *CustomResponseWriter) ([]byte, error) {
	buf := gzipBufferPool.Get().(*bytes.Buffer)
	gz := gzipWriterPool.Get().(*gzip.Writer)

	buf.Reset()
	gz.Reset(buf)

	_, err := gz.Write(crw.buf.Bytes())
	if err != nil {
		gz.Close()
		gzipWriterPool.Put(gz)
		gzipBufferPool.Put(buf)
		return nil, err
	}

	if err := gz.Close(); err != nil {
		gzipWriterPool.Put(gz)
		gzipBufferPool.Put(buf)
		return nil, err
	}

	compressed := make([]byte, buf.Len())
	copy(compressed, buf.Bytes())

	gzipWriterPool.Put(gz)
	gzipBufferPool.Put(buf)

	return compressed, nil
}

func decompessRequest(r *http.Request) ([]byte, error) {
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
