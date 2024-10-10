package models

const (
	// Gauge обозначает тип метрики для значения типа "гейдж".
	Gauge = "gauge"
	// Counter обозначает тип метрики для значения типа "счетчик".
	Counter = "counter"
)

// Metric представляет метрику с ее параметрами.
type Metric struct {
	// ID - имя метрики.
	ID string `json:"id" db:"id"`
	// MType - тип метрики, который может быть "gauge" или "counter".
	MType string `json:"type" db:"type"`
	// Delta - значение метрики в случае передачи счетчика (counter).
	Delta *int64 `json:"delta,omitempty" db:"delta"`
	// Value - значение метрики в случае передачи гейджа (gauge).
	Value *float64 `json:"value,omitempty" db:"value"`
}

// MetricsData хранит данные о метриках.
type MetricsData struct {
	// Counter - карта для хранения счетчиков.
	Counter map[string]int64 `json:"counter"`
	// Gauge - карта для хранения гейджа.
	Gauge map[string]float64 `json:"gauge"`
}
