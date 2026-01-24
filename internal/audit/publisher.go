package audit

import (
	"context"
	"fmt"
	"sync"
)

type Publisher struct {
	mu        sync.RWMutex
	observers []Observer

	ch chan Event
}

func NewPublisher(buffer int) *Publisher {
	return &Publisher{
		ch: make(chan Event, buffer),
	}
}

func (p *Publisher) Subscribe(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers = append(p.observers, o)
}

// Start запускается один раз при старте сервера.
func (p *Publisher) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-p.ch:
				fmt.Println("!!!!!!Данные из канала получины")
				p.mu.RLock()
				obs := append([]Observer(nil), p.observers...)
				p.mu.RUnlock()

				for _, o := range obs {
					fmt.Println("!!!!!!Вызовы notify")
					o.Notify(ctx, e)
				}
			}
		}
	}()
}

// Publish НЕ блокирует хендлер: если буфер забит — событие можно дропнуть или считать метрику дропа.
func (p *Publisher) Publish(ctx context.Context, e Event) {
	select {
	case p.ch <- e:
	default:
		// буфер заполнен — дропаем (или добавь счетчик dropped_audit_events)
	}
}
