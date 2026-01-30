package model

import (
	"testing"
)

func TestContainer_GetPut(t *testing.T) {
	container := NewContainer[*Metrics]()

	m1 := &Metrics{
		ID:    "gauge1",
		MType: "gauge",
		Value: floatPtr(3.14),
	}

	container.Put(m1)

	m2 := container.Get()

	if m2 != m1 {
		t.Errorf("Метрики разные")
	}

	if item := container.Get(); item != nil {
		t.Error("Контейнер не пустой")
	}
}

func TestContainer_ResetCalledOnPut(t *testing.T) {
	container := NewContainer[*Metrics]()

	m := &Metrics{
		ID:    "counter1",
		MType: "counter",
		Delta: intPtr(42),
	}

	container.Put(m)

	if m.Delta != nil {
		t.Error("Значение не сброшено")
	}
}

func TestContainer_Concurrent(t *testing.T) {
	container := NewContainer[*Metrics]()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			m := &Metrics{ID: "test", MType: "gauge"}
			container.Put(m)
			_ = container.Get()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

}
