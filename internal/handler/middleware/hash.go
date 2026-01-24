package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
)

func equalHash(r *http.Request, key string) error {
	if key == "" {
		return nil
	}
	receivedHash := r.Header.Get("HashSHA256")
	if receivedHash == "" {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	expectedHash := h.Sum(nil)
	expectedHashHex := hex.EncodeToString(expectedHash)
	if !hmac.Equal([]byte(expectedHashHex), []byte(receivedHash)) {
		return errors.New("хеши не совпадают")
	}
	return nil
}

// WithCheckHash возвращает middleware, которое проверяет подпись тела запроса
// с помощью HMAC-SHA256 в заголовке "HashSHA256".
//
// Использование:
//
//	r.Use(middleware.WithCheckHash("my-super-secret-key"))
//
// Поведение:
//   - Для каждого запроса вычисляется HMAC от тела с использованием переданного ключа.
//   - Сравнивается с hex-кодированным значением из заголовка HashSHA256.
//   - Если проверка не пройдена — возвращается 400 Bad Request.
//   - Если ключ пустой — middleware не выполняет проверку.
func WithCheckHash(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if err := equalHash(r, key); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
