package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

var (
	Key string
)

func EqualHash(r *http.Request) bool {
	if Key == "" {
		return true
	}
	receivedHash := r.Header.Get("HashSHA256")
	if receivedHash == "" {
		return true
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return true
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	h := hmac.New(sha256.New, []byte(Key))
	h.Write(body)
	expectedHash := h.Sum(nil)
	expectedHashHex := hex.EncodeToString(expectedHash)
	return hmac.Equal([]byte(expectedHashHex), []byte(receivedHash))

}

func WithCheckHash(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !EqualHash(r) {
			http.Error(w, "Хеши не совпадают", http.StatusBadRequest)
			return
		}

		h.ServeHTTP(w, r)

	})
}
