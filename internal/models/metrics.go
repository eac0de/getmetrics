package models

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type (
	Metrics struct {
		ID    string   `json:"id"`              // имя метрики
		MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
		Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
		Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	}

	MetricsDict struct {
		Counter map[string]int64   `json:"counter"`
		Gauge   map[string]float64 `json:"gauge"`
	}
)
