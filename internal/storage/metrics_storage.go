package storage

import "sync"

type Gauge float64
type Counter int64

type MetricsStorage struct {
	mu            sync.Mutex
	SystemMetrics map[string]interface{}
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		SystemMetrics: make(map[string]interface{}),
	}
}

func (m *MetricsStorage) Save(metricName string, metricValue interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SystemMetrics[metricName] = metricValue
}

func (m *MetricsStorage) Get(metricName string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.SystemMetrics[metricName]
	if !ok {
		return nil
	}
	return value
}

func (m *MetricsStorage) GetAll() map[string]interface{} {
	return m.SystemMetrics
}
