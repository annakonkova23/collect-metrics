package handler

import (
	"bytes"

	//"fmt"
	"net/http"
	"net/http/httptest"

	//"strconv"
	"testing"

	//"github.com/annakonkova23/collect-metrics/internal/model"
	//"github.com/go-resty/resty/v2"
	//"github.com/stretchr/testify/assert"
	"github.com/annakonkova23/collect-metrics/internal/config"
	"go.uber.org/zap"
)

func TestServer_updateJsonHandler(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	cfg := config.NewServerOptions()
	jsonPositiveGauge := `{
		"id": "TestGauge",
		"type": "gauge",
		"value": 1744184459
		} 
	`
	jsonPositiveCounter := `{
		"id": "TestCounter",
		"type": "counter",
		"delta": 17
		} 
	`
	reqPositiveGauge, _ := http.NewRequest("POST", "/update/", bytes.NewBuffer([]byte(jsonPositiveGauge)))
	reqPositiveGauge.Header.Set("Content-Type", "application/json")
	reqPositiveCounter, _ := http.NewRequest("POST", "/update/", bytes.NewBuffer([]byte(jsonPositiveCounter)))
	reqPositiveCounter.Header.Set("Content-Type", "application/json")
	reqNotContentType, _ := http.NewRequest("POST", "/update/", bytes.NewBuffer([]byte(jsonPositiveCounter)))
	tests := []struct {
		name   string // description of this test case
		w      *httptest.ResponseRecorder
		r      *http.Request
		result int
	}{
		{
			name:   "PositiveGauge",
			w:      httptest.NewRecorder(),
			r:      reqPositiveGauge,
			result: http.StatusOK,
		},
		{
			name:   "PositiveCounter",
			w:      httptest.NewRecorder(),
			r:      reqPositiveCounter,
			result: http.StatusOK,
		},
		{
			name:   "NotContentType",
			w:      httptest.NewRecorder(),
			r:      reqNotContentType,
			result: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewServer(cfg, logger)
			if err != nil {
				t.Errorf("Error on creating server: %v", err)
			}
			tt.r.Header.Set("Content-Type", "application/json")
			s.updateJSONHandler(tt.w, tt.r)

		})
	}
}
