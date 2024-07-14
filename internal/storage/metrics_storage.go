package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/eac0de/getmetrics/internal/models"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type MetricsStorage struct {
	mu            sync.Mutex
	SystemMetrics models.SystemMetrics
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		SystemMetrics: models.SystemMetrics{
			Counter: map[string]int64{},
			Gauge:   map[string]float64{},
		},
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
			valueStr, ok := metricValue.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value type for guage metric(1)")
			}
			metricValueFloat, err = strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value type for guage metric(2)")
			}
		}
		m.SystemMetrics.Gauge[metricName] = metricValueFloat
		metric.Value = &metricValueFloat
		metric.MType = Gauge
	case Counter:
		metricValueInt, ok := metricValue.(int64)
		if !ok {
			valueStr, ok := metricValue.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value type for counter metric(1)")
			}
			metricValueInt, err = strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value type for counter metric(2)")
			}
		}
		if oldValue, ok := m.SystemMetrics.Counter[metricName]; ok {
			metricValueInt = metricValueInt + oldValue
		}
		m.SystemMetrics.Counter[metricName] = metricValueInt
		metric.Delta = &metricValueInt
		metric.MType = Counter
	default:
		// Обработка некорректного типа метрики
		return nil, fmt.Errorf("invalid metric type")
	}
	metric.ID = metricName
	return &metric, nil
}

func (m *MetricsStorage) Get(metricType string, metricName string) *models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	var metric models.Metrics
	switch metricType {
	case Gauge:
		value, ok := m.SystemMetrics.Gauge[metricName]
		if !ok {
			return nil
		}
		metric.Value = &value
		metric.MType = Gauge
	case Counter:
		value, ok := m.SystemMetrics.Counter[metricName]
		if !ok {
			return nil
		}
		metric.Delta = &value
		metric.MType = Counter
	default:
		// Обработка некорректного типа метрики
		return nil
	}
	metric.ID = metricName
	return &metric
}

func (m *MetricsStorage) GetAll() []*models.Metrics {
	var metrics []*models.Metrics
	for name, value := range m.SystemMetrics.Gauge {
		valueCopy := value
		metric := models.Metrics{
			ID:    name,
			MType: Gauge,
			Value: &valueCopy,
		}
		metrics = append(metrics, &metric)
	}
	for name, value := range m.SystemMetrics.Counter {
		valueCopy := value
		metric := models.Metrics{
			ID:    name,
			MType: Counter,
			Delta: &valueCopy,
		}
		metrics = append(metrics, &metric)
	}
	return metrics
}

func (m *MetricsStorage) LoadMetricsFromFile(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// Проверяем, что файл не пустой
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		// Если файл пустой, возвращаем nil, так как это не ошибка
		return nil
	}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&m.SystemMetrics); err != nil {
		return err
	}
	return nil
}

func (m *MetricsStorage) SaveMetricsToFile(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.MarshalIndent(m.SystemMetrics, "", "    ")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	log.Println("metrics have been preserved")
	return nil
}

func (m *MetricsStorage) StartSavingMetricsToFile(ctx context.Context, filename string, interval time.Duration) {
	if filename == "" {
		return
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			log.Println("SaveMetricsToFile goroutine is shutting down...")
			return
		case <-ticker.C:
			err := m.SaveMetricsToFile(filename)
			if err != nil {
				fmt.Printf("metrics saving error: %s", err.Error())
			}
		}

	}

}
