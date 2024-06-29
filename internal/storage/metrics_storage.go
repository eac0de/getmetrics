package storage

import (
	"fmt"
	"strconv"
	"sync"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type (
	MetricsStorage struct {
		mu            sync.Mutex
		SystemMetrics map[string]map[string]interface{}
	}
)

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		SystemMetrics: make(map[string]map[string]interface{}),
	}
}

func (m *MetricsStorage) Save(metricType string, metricName string, metricValue interface{}) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch metricType {
	case Gauge:
		if _, ok := metricValue.(float64); !ok {
			if valueStr, ok := metricValue.(string); ok {
				metricValue, err = strconv.ParseFloat(valueStr, 64)
				if err != nil {
					return fmt.Errorf("invalid value type for guage metric")
				}
			} else {
				return fmt.Errorf("invalid value type for guage metric")
			}
		}
		if m.SystemMetrics[Gauge] == nil {
			m.SystemMetrics[Gauge] = make(map[string]interface{})
		}
	case Counter:
		metricValueInt, ok := metricValue.(int64)
		if !ok {
			if valueStr, ok := metricValue.(string); ok {
				metricValueInt, err = strconv.ParseInt(valueStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid value type for counter metric")
				}
			} else {
				return fmt.Errorf("invalid value type for counter metric")
			}
		}
		if m.SystemMetrics[Counter] == nil {
			m.SystemMetrics[Counter] = make(map[string]interface{})
		}
		if value, ok := m.SystemMetrics[Counter][metricName]; ok {
			if oldValue, ok := value.(int64); ok {
				metricValueInt = metricValueInt + oldValue
			} else {
				return fmt.Errorf("stored counter value has invalid type")
			}
		}
		metricValue = metricValueInt
	default:
		// Обработка некорректного типа метрики
		return fmt.Errorf("invalid metric type")
	}
	m.SystemMetrics[metricType][metricName] = metricValue
	return
}

func (m *MetricsStorage) Get(metricType string, metricName string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.SystemMetrics[metricType][metricName]
	if !ok {
		return nil
	}
	return value
}

func (m *MetricsStorage) GetAll() map[string]map[string]interface{} {
	return m.SystemMetrics
}
