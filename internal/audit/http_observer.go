package audit

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type HTTPObserver struct {
	client *resty.Client
	url    string
	ch     chan Event
	logger *zap.Logger
}

// NewHTTPObserver создаёт новый HTTP-наблюдатель, отправляющий аудит-события на удалённый сервер.
//
// Наблюдатель работает асинхронно: события помещаются в канал с заданным размером буфера
// и отправляются в фоновом режиме через HTTP POST-запросы с Content-Type: application/json.
//
// Параметры:
//   - logger: логгер для записи ошибок и отладочной информации.
//   - url: адрес HTTP-эндпоинта, куда будут отправляться события (например, "http://localhost:8080/audit").
//   - bufferSize: размер внутреннего канала событий; определяет, сколько событий можно накопить до блокировки.
//     Если 0, используется значение по умолчанию (100).
//
// Возвращает указатель на *HTTPObserver.
func NewHTTPObserver(logger *zap.Logger, url string, bufferSize int) *HTTPObserver {
	return &HTTPObserver{
		client: resty.New(),
		url:    url,
		ch:     make(chan Event, bufferSize),
		logger: logger,
	}
}

// Start при старте приложения, чтобы был воркер отправки.
// ctx — общий контекст приложения.
func (o *HTTPObserver) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-o.ch:
				o.send(e)
			}
		}
	}()
}

// Notify добавление события в очередь.
func (o *HTTPObserver) Notify(_ context.Context, e Event) {
	select {
	case o.ch <- e:
	default:
	}
}

// send отправка события.
func (o *HTTPObserver) send(e Event) {
	body, err := json.Marshal(e)
	if err != nil {
		return
	}

	response, err := o.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(o.url)

	if err != nil {
		o.logger.Error("Ошибка отправки запроса", zap.Error(err))
		return
	}

	if response.StatusCode() != http.StatusOK {
		o.logger.Error("Ошибка отправки запроса", zap.ByteString("body", response.Body()))
	}

}
