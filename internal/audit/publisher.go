package audit

import (
	"context"
	"sync"
)

// Publisher
type Publisher struct {
	mu        sync.RWMutex
	observers []Observer

	ch chan Event
}

// NewPublisher создаёт новый Publisher с заданным размером буфера.
func NewPublisher(buffer int) *Publisher {
	return &Publisher{
		ch: make(chan Event, buffer),
	}
}

// Subscribe добавляет наблюдателя .
func (p *Publisher) Subscribe(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers = append(p.observers, o)
}

// Start запускает Notify для всех подписчиков.
func (p *Publisher) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-p.ch:
				p.mu.RLock()
				obs := append([]Observer(nil), p.observers...)
				p.mu.RUnlock()

				for _, o := range obs {
					o.Notify(ctx, e)
				}
			}
		}
	}()
}

// Publish добавление событие на публикацию.
func (p *Publisher) Publish(ctx context.Context, e Event) {
	select {
	case p.ch <- e:
	default:
	}
}
