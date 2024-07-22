package storage

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/eac0de/getmetrics/internal/models"
)

type memoryStorage struct {
	MetricsMap models.MetricsMap
}

func NewMemoryStorage() *memoryStorage {
	mems := memoryStorage{
		MetricsMap: models.MetricsMap{
			Counter: map[string]int64{},
			Gauge:   map[string]float64{},
		},
	}
	return &mems
}

func (mems *memoryStorage) Save(ctx context.Context, metric *models.Metrics) error {
	if metric.ID == "" {
		NewErrorWithHTTPStatus(fmt.Errorf("Metric name is required"), 404)
	}
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s must have filled value", metric.ID, models.Gauge), 400)
		}
		mems.MetricsMap.Gauge[metric.ID] = *metric.Value
	case models.Counter:
		if metric.Delta == nil {
			return NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s must have filled delta", metric.ID, models.Counter), 400)
		}
		existMetric, err := mems.Get(ctx, metric.ID, metric.MType)
		oldDelta := int64(0)
		if err == nil {
			oldDelta = *existMetric.Delta
		}
		mems.MetricsMap.Counter[metric.ID] = *metric.Delta + oldDelta
	default:
		return NewErrorWithHTTPStatus(fmt.Errorf("Invalid metric type for %s: %s", metric.ID, models.Counter), 400)
	}
	return nil
}

func (mems *memoryStorage) SaveMany(ctx context.Context, metricsList []*models.Metrics) error {
	var errList []error
	for _, metric := range metricsList {
		err := mems.Save(ctx, metric)
		if err != nil {
			errList = append(errList, err)
		}
	}
	err := errors.Join(errList...)
	if err != nil {
		return NewErrorWithHTTPStatus(err, 400)
	}
	return nil
}

func (mems *memoryStorage) Get(ctx context.Context, metricName string, metricType string) (*models.Metrics, error) {
	var metric models.Metrics
	switch metricType {
	case models.Gauge:
		value, ok := mems.MetricsMap.Gauge[metricName]
		if !ok {
			return nil, NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s not found", metricName, metricType), 404)
		}
		metric.Value = &value
	case models.Counter:
		delta, ok := mems.MetricsMap.Counter[metricName]
		if !ok {
			return nil, NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s not found", metricName, metricType), 404)
		}
		metric.Delta = &delta
	default:
		return nil, NewErrorWithHTTPStatus(fmt.Errorf("Invalid type for %s: %s", metricName, metricType), 400)
	}
	metric.MType = metricType
	metric.ID = metricName
	return &metric, nil
}

func (mems *memoryStorage) GetAll(ctx context.Context) ([]*models.Metrics, error) {
	var metrics []*models.Metrics
	for name, value := range mems.MetricsMap.Gauge {
		valueCopy := value
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &valueCopy,
		}
		metrics = append(metrics, &metric)
	}
	for name, value := range mems.MetricsMap.Counter {
		valueCopy := value
		metric := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &valueCopy,
		}
		metrics = append(metrics, &metric)
	}
	return metrics, nil
}

func (mems *memoryStorage) Close() error {
	log.Println("MemoryStorage closed correctly")
	return nil
}

func (mems *memoryStorage) Ping(ctx context.Context) error {
	return NewErrorWithHTTPStatus(fmt.Errorf("Apparently the database failed to initialize"), 500)
}
