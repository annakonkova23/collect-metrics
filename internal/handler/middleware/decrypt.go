package middleware

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"

	"github.com/annakonkova23/collect-metrics/internal/model"
)

func WithDecrypt(key *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			newReq := r.Clone(r.Context())
			if key != nil {
				if r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Content-Type") == "text/html" {
					var req model.EncryptedRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					dcr, err := HybridDecrypt(key, req.EncryptedKey, req.EncryptedData)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					newReq.Body = io.NopCloser(bytes.NewBuffer(dcr))
				}

			}
			next.ServeHTTP(w, newReq)
		})
	}
}
