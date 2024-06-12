package storage

import "sync"

type Gauge float64
type Counter int64

type MetricsStorage struct {
	mu      sync.Mutex
	Metrics map[string]interface{}
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		Metrics: make(map[string]interface{}),
	}
}

func (m *MetricsStorage) Save(metricName string, metricValue interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Metrics[metricName] = metricValue
}

func (m *MetricsStorage) Get(metricName string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.Metrics[metricName]
	if !ok {
		return nil
	}
	return value
}
