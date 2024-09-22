package models

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type (
	Metric struct {
		ID    string   `json:"id" db:"id"`                 // имя метрики
		MType string   `json:"type" db:"type"`             // параметр, принимающий значение gauge или counter
		Delta *int64   `json:"delta,omitempty" db:"delta"` // значение метрики в случае передачи counter
		Value *float64 `json:"value,omitempty" db:"value"` // значение метрики в случае передачи gauge
	}

	MetricsData struct {
		Counter map[string]int64   `json:"counter"`
		Gauge   map[string]float64 `json:"gauge"`
	}
)
