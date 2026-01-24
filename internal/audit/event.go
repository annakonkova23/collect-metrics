package audit

import (
	"context"
	"encoding/json"
	"time"
)

// Event — событие аудита.
type Event struct {
	Time    time.Time `json:"ts"`
	Metrics []string  `json:"metrics"`
	IP      string    `json:"ip_address"`
}

// Observer — наблюдатель за событиями аудита.
type Observer interface {
	Notify(ctx context.Context, e Event)
}

// MarshalJSON реализует интерфейс json.Marshaler.
func (e Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		TS      int64    `json:"ts"`
		Metrics []string `json:"metrics"`
		IP      string   `json:"ip_address"`
	}{
		TS:      e.Time.Unix(),
		Metrics: e.Metrics,
		IP:      e.IP,
	})
}
