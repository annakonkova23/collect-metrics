package model

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) Copy() *Metrics {
	// Копирование Delta
	var delta *int64
	if m.Delta != nil {
		d := *m.Delta
		delta = &d
	}

	// Копирование Value
	var value *float64
	if m.Value != nil {
		v := *m.Value
		value = &v
	}

	// Создание копии структуры
	return &Metrics{
		ID:    m.ID,
		MType: m.MType,
		Delta: delta,
		Value: value,
		Hash:  m.Hash,
	}
}
