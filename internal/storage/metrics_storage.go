package storage

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/eac0de/getmetrics/internal/models"
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

func (m *MetricsStorage) Save(metricType string, metricName string, metricValue interface{}) (*models.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var err error
	var metric models.Metrics
	switch metricType {
	case Gauge:
		metricValueFloat, ok := metricValue.(float64)
		if !ok {
			if valueStr, ok := metricValue.(string); ok {
				metricValueFloat, err = strconv.ParseFloat(valueStr, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid value type for guage metric")
				}
			} else {
				return nil, fmt.Errorf("invalid value type for guage metric")
			}
		}
		if m.SystemMetrics[Gauge] == nil {
			m.SystemMetrics[Gauge] = make(map[string]interface{})
		}
		metricValue = metricValueFloat
		metric.Value = &metricValueFloat
	case Counter:
		metricValueInt, ok := metricValue.(int64)
		if !ok {
			if valueStr, ok := metricValue.(string); ok {
				metricValueInt, err = strconv.ParseInt(valueStr, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid value type for counter metric")
				}
			} else {
				return nil, fmt.Errorf("invalid value type for counter metric")
			}
		}
		if m.SystemMetrics[Counter] == nil {
			m.SystemMetrics[Counter] = make(map[string]interface{})
		}
		if value, ok := m.SystemMetrics[Counter][metricName]; ok {
			if oldValue, ok := value.(int64); ok {
				metricValueInt = metricValueInt + oldValue
			} else {
				return nil, fmt.Errorf("stored counter value has invalid type")
			}
		}
		metricValue = metricValueInt
		metric.Delta = &metricValueInt
	default:
		// Обработка некорректного типа метрики
		return nil, fmt.Errorf("invalid metric type")
	}
	metric.ID = metricName
	metric.MType = metricType
	m.SystemMetrics[metricType][metricName] = metricValue
	return &metric, nil
}

func (m *MetricsStorage) Get(metricType string, metricName string) *models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.SystemMetrics[metricType][metricName]
	if !ok {
		return nil
	}
	metric := models.Metrics{
		ID: metricName,
	}
	switch metricType {
	case Gauge:
		valueFloat, _ := value.(float64)
		metric.Value = &valueFloat
		metric.MType = Gauge
	case Counter:
		valueFloat, _ := value.(int64)
		metric.Delta = &valueFloat
		metric.MType = Counter
	}
	return &metric
}

func (m *MetricsStorage) GetAll() map[string]map[string]interface{} {
	return m.SystemMetrics
}
