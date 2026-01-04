package audit

import (
	"context"
	"encoding/json"
	"time"
)

type Event struct {
	Time    time.Time `json:"ts"`
	Metrics []string  `json:"metrics"`
	IP      string    `json:"ip_address"`
}

type Observer interface {
	Notify(ctx context.Context, e Event)
}

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
