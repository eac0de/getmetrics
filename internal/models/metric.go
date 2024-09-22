package models

import (
	"fmt"
)

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

func ValidateMetric(metric Metric) error {
	if metric.ID == "" {
		return fmt.Errorf("metric name is required")
	}
	switch metric.MType {
	case Gauge:
		if metric.Value == nil {
			return fmt.Errorf("metric %s with type %s must have filled value", metric.ID, metric.MType)
		}
	case Counter:
		if metric.Delta == nil {
			return fmt.Errorf("metric %s with type %s must have filled delta", metric.ID, metric.MType)
		}
	default:
		return fmt.Errorf("invalid metric type for %s: %s", metric.ID, metric.MType)
	}
	return nil
}

func MergeMetricsList(metricsList []Metric) []Metric {
	metricsMap := MetricsData{
		Gauge:   map[string]float64{},
		Counter: map[string]int64{},
	}

	// Обработка метрик
	for _, metric := range metricsList {
		switch metric.MType {
		case Gauge:
			metricsMap.Gauge[metric.ID] = *metric.Value
		case Counter:
			metricsMap.Counter[metric.ID] += *metric.Delta
		}
	}

	// Формируем результирующий список
	mergeMetricsList := make([]Metric, 0, len(metricsMap.Gauge)+len(metricsMap.Counter))

	// Добавляем все gauge метрики
	for ID, value := range metricsMap.Gauge {
		mergeMetricsList = append(mergeMetricsList, Metric{ID: ID, MType: Gauge, Value: &value})
	}

	// Добавляем все counter метрики
	for ID, value := range metricsMap.Counter {
		mergeMetricsList = append(mergeMetricsList, Metric{ID: ID, MType: Counter, Delta: &value})
	}

	return mergeMetricsList
}
