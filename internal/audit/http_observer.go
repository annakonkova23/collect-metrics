package audit

import (
	"context"
	"encoding/json"
	"fmt"
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

func NewHTTPObserver(logger *zap.Logger, url string, bufferSize int) *HTTPObserver {
	return &HTTPObserver{
		client: resty.New(),
		url:    url,
		ch:     make(chan Event, bufferSize),
		logger: logger,
	}
}

// Start запускай при старте приложения, чтобы был воркер отправки.
// ctx — общий контекст приложения (на shutdown отменится).
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

// Notify должен быть быстрым: просто пытается enqueue.
// Если очередь заполнена — дропаем, чтобы не повесить API.
func (o *HTTPObserver) Notify(_ context.Context, e Event) {
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("Notify")
	select {
	case o.ch <- e:
	default:
		// очередь забита — дропаем событие (можно добавить счетчик dropped_http_audit)
	}
}

func (o *HTTPObserver) send(e Event) {
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	fmt.Println("Отправка запроса!")
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
