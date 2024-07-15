package models

type (
	Metrics struct {
		ID    string   `json:"id"`              // имя метрики
		MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
		Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
		Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	}

	UnknownMetrics struct {
		ID    string
		MType string              
		DeltaValue interface{}    
	}

	SystemMetrics struct {
		Counter map[string]int64   `json:"counter"`
		Gauge   map[string]float64 `json:"gauge"`
	}
)
