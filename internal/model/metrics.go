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
//
//easyjson:json
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) Copy() *Metrics {
	cpy := *m
	if m.Delta != nil {
		delta := *m.Delta
		cpy.Delta = &delta
	}
	if m.Value != nil {
		value := *m.Value
		cpy.Value = &value
	}
	return &cpy
}

func GetListIDMetrics(ms []*Metrics) []string {
	metrics := make([]string, 0)
	for _, m := range ms {
		metrics = append(metrics, m.ID)
	}
	return metrics
}
