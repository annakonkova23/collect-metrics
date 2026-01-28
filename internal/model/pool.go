package model

import "sync"

type Resetable interface {
	Reset()
}

type Container[T Resetable] struct {
	items []T
	mx    sync.RWMutex
}

// New - создает новый контейнер
func NewContainer[T Resetable]() *Container[T] {
	return &Container[T]{
		items: make([]T, 0),
	}
}

// Get возвращает объект из пула.
// Если пул пуст — возвращается нулевое значение (nil для указателей).
func (c *Container[T]) Get() T {
	c.mx.Lock()
	defer c.mx.Unlock()
	if len(c.items) == 0 {
		var zero T
		return zero
	}
	item := c.items[len(c.items)-1]
	c.items = c.items[:len(c.items)-1]
	return item
}

// Put возвращает объект в пул.
// Перед этим вызывает Reset(), чтобы очистить состояние.
func (c *Container[T]) Put(item T) {
	c.mx.Lock()
	defer c.mx.Unlock()
	item.Reset()
	c.items = append(c.items, item)
}
