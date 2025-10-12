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
	"go.uber.org/zap"
)

func TestServer_updateJsonHandler(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	url := "localhost:8080"
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
	reqPositiveGauge, err := http.NewRequest("POST", "/update/", bytes.NewBuffer([]byte(jsonPositiveGauge)))
	reqPositiveGauge.Header.Set("Content-Type", "application/json")
	reqPositiveCounter, err := http.NewRequest("POST", "/update/", bytes.NewBuffer([]byte(jsonPositiveCounter)))
	reqPositiveCounter.Header.Set("Content-Type", "application/json")
	reqNotContentType, err := http.NewRequest("POST", "/update/", bytes.NewBuffer([]byte(jsonPositiveCounter)))
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
			s := NewServer(url, logger)
			s.updateJsonHandler(tt.w, tt.r)

		})
	}
}

/*
func TestServer_valueJsonHandler(t *testing.T) {

	httpc := resty.New()

	tests := []struct {
		name   string
		method string
		value  float64
		delta  int64
		update int
		ok     bool
		static bool
	}{
		{method: "counter", name: "PollCount"},
		{method: "gauge", name: "RandomValue"},
		{method: "gauge", name: "Alloc"},
		{method: "gauge", name: "BuckHashSys", static: true},
		{method: "gauge", name: "Frees"},
		{method: "gauge", name: "GCCPUFraction", static: true},
		{method: "gauge", name: "GCSys", static: true},
		{method: "gauge", name: "HeapAlloc"},
		{method: "gauge", name: "HeapIdle"},
		{method: "gauge", name: "HeapInuse"},
		{method: "gauge", name: "HeapObjects"},
		{method: "gauge", name: "HeapReleased", static: true},
		{method: "gauge", name: "HeapSys", static: true},
		{method: "gauge", name: "LastGC", static: true},
		{method: "gauge", name: "Lookups", static: true},
		{method: "gauge", name: "MCacheInuse", static: true},
		{method: "gauge", name: "MCacheSys", static: true},
		{method: "gauge", name: "MSpanInuse", static: true},
		{method: "gauge", name: "MSpanSys", static: true},
		{method: "gauge", name: "Mallocs"},
		{method: "gauge", name: "NextGC", static: true},
		{method: "gauge", name: "NumForcedGC", static: true},
		{method: "gauge", name: "NumGC", static: true},
		{method: "gauge", name: "OtherSys", static: true},
		{method: "gauge", name: "PauseTotalNs", static: true},
		{method: "gauge", name: "StackInuse", static: true},
		{method: "gauge", name: "StackSys", static: true},
		{method: "gauge", name: "Sys", static: true},
		{method: "gauge", name: "TotalAlloc"},
	}

	req := httpc.R().
		SetHeader("Content-Type", "application/json")

	for _, tt := range tests {

		var result model.Metrics
		_, _ = req.
			SetBody(&model.Metrics{
				ID:    tt.name,
				MType: tt.method,
			}).
			SetResult(&result).
			Post("http://localhost:8080/value/")

		//assert.NoError(t, err, "Error on POST /value/")
		fmt.Println(tt.name)
		fmt.Println(result.MType)
		fmt.Println(result.Value)
		fmt.Println(result.Delta)
		assert.True(t, ((result.MType == "gauge" && result.Value != nil) || (result.MType == "counter" && result.Delta != nil)),
			"Получен не однозначный результат (тип метода не соответствует возвращаемому значению) '%q %s %s'", req.Method, req.URL, tt.name)

		//assert.Containsf(t, resp.Header().Get("Content-Type"), "application/json",
		//"Заголовок ответа Content-Type содержит несоответствующее значение")

	}

}*/

/*
func TestServer_valueCounterHandler(t *testing.T) {

	httpc := resty.New()

	id := "GetSet" + strconv.Itoa(2)

	t.Run("update", func(t *testing.T) {
		//value1, value2 := 5, 4
		req := httpc.R().
			SetHeader("Content-Type", "application/json")

		// Запросим предыдущее значение с сервера, на случай если оно там уже есть.
		var result model.Metrics
		_, _ = req.
			SetBody(&model.Metrics{
				ID:    id,
				MType: "counter",
			}).
			SetResult(&result).
			Post("http://localhost:8080/value/")

		assert.NotNil(t, result.Delta,
			"Получено не инициализированное значение Delta '%q %s'", req.Method, req.URL)

	})

}
*/
