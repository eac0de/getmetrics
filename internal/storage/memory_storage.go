package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/eac0de/getmetrics/internal/models"
)

type memoryStorage struct {
	mu          sync.Mutex
	MetricsDict models.MetricsDict
}

func NewMemoryStorage() *memoryStorage {
	mems := memoryStorage{
		MetricsDict: models.MetricsDict{
			Counter: map[string]int64{},
			Gauge:   map[string]float64{},
		},
	}
	return &mems
}

func (mems *memoryStorage) Save(ctx context.Context, metric models.Metrics) (*models.Metrics, error) {
	mems.mu.Lock()
	defer mems.mu.Unlock()
	if metric.ID == "" {
		return nil, NewErrorWithHTTPStatus(fmt.Errorf("metric name is required"), http.StatusNotFound)
	}
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return nil, NewErrorWithHTTPStatus(fmt.Errorf("metric %s with type %s must have filled value", metric.ID, models.Gauge), http.StatusBadRequest)
		}
		mems.MetricsDict.Gauge[metric.ID] = *metric.Value
	case models.Counter:
		if metric.Delta == nil {
			return nil, NewErrorWithHTTPStatus(fmt.Errorf("metric %s with type %s must have filled delta", metric.ID, models.Counter), http.StatusBadRequest)
		}
		existMetric, err := mems.getMetric(metric.ID, metric.MType)
		oldDelta := int64(0)
		if err == nil {
			oldDelta = *existMetric.Delta
		}
		mems.MetricsDict.Counter[metric.ID] = *metric.Delta + oldDelta
	default:
		return nil, NewErrorWithHTTPStatus(fmt.Errorf("invalid metric type for %s: %s", metric.ID, metric.MType), http.StatusBadRequest)
	}
	return &metric, nil
}

func (mems *memoryStorage) SaveMany(ctx context.Context, metricsList []models.Metrics) ([]*models.Metrics, error) {
	metricsList, err := mems.MergeMetricsList(metricsList)
	if err != nil {
		return nil, err
	}
	var errList []error
	var newMetricsList []*models.Metrics
	for _, metric := range metricsList {
		newMetric, err := mems.Save(ctx, metric)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		newMetricsList = append(newMetricsList, newMetric)
	}
	err = errors.Join(errList...)
	if err != nil {
		return nil, NewErrorWithHTTPStatus(err, http.StatusBadRequest)
	}
	return newMetricsList, nil
}

func (mems *memoryStorage) Get(ctx context.Context, metricName string, metricType string) (*models.Metrics, error) {
	mems.mu.Lock()
	defer mems.mu.Unlock()
	return mems.getMetric(metricName, metricType)
}

func (mems *memoryStorage) GetAll(ctx context.Context) ([]*models.Metrics, error) {
	mems.mu.Lock()
	defer mems.mu.Unlock()
	var metrics []*models.Metrics
	for name, value := range mems.MetricsDict.Gauge {
		valueCopy := value
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &valueCopy,
		}
		metrics = append(metrics, &metric)
	}
	for name, value := range mems.MetricsDict.Counter {
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
	return NewErrorWithHTTPStatus(fmt.Errorf("apparently the database failed to initialize"), http.StatusInternalServerError)
}

func (mems *memoryStorage) MergeMetricsList(metricsList []models.Metrics) ([]models.Metrics, error) {
	metricsMap := models.MetricsDict{
		Gauge:   map[string]float64{},
		Counter: map[string]int64{},
	}
	var errList []error
	for _, metric := range metricsList {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				errList = append(errList, NewErrorWithHTTPStatus(fmt.Errorf("metric %s with type %s must have filled value", metric.ID, models.Gauge), http.StatusBadRequest))
				continue
			}
			metricsMap.Gauge[metric.ID] = *metric.Value
		case models.Counter:
			if metric.Delta == nil {
				errList = append(errList, NewErrorWithHTTPStatus(fmt.Errorf("metric %s with type %s must have filled delta", metric.ID, models.Counter), http.StatusBadRequest))
				continue
			}
			oldDelta, ok := metricsMap.Counter[metric.ID]
			if !ok {
				oldDelta = int64(0)
			}
			delta := *metric.Delta + oldDelta
			metricsMap.Counter[metric.ID] = delta
		default:
			errList = append(errList, NewErrorWithHTTPStatus(fmt.Errorf("invalid metric type for %s: %s", metric.ID, models.Counter), http.StatusBadRequest))
		}
	}
	err := errors.Join(errList...)
	if err != nil {
		return nil, NewErrorWithHTTPStatus(err, http.StatusBadRequest)
	}
	var mergeMetricsList []models.Metrics
	for ID, value := range metricsMap.Gauge {
		mergeMetricsList = append(mergeMetricsList, models.Metrics{ID: ID, MType: models.Gauge, Value: &value})
	}
	for ID, value := range metricsMap.Counter {
		mergeMetricsList = append(mergeMetricsList, models.Metrics{ID: ID, MType: models.Counter, Delta: &value})
	}
	return mergeMetricsList, nil
}

func (mems *memoryStorage) getMetric(metricName string, metricType string) (*models.Metrics, error) {
	var metric models.Metrics
	switch metricType {
	case models.Gauge:
		value, ok := mems.MetricsDict.Gauge[metricName]
		if !ok {
			return nil, NewErrorWithHTTPStatus(fmt.Errorf("metric %s with type %s not found", metricName, metricType), http.StatusNotFound)
		}
		metric.Value = &value
	case models.Counter:
		delta, ok := mems.MetricsDict.Counter[metricName]
		if !ok {
			return nil, NewErrorWithHTTPStatus(fmt.Errorf("metric %s with type %s not found", metricName, metricType), http.StatusNotFound)
		}
		metric.Delta = &delta
	default:
		return nil, NewErrorWithHTTPStatus(fmt.Errorf("invalid type for %s: %s", metricName, metricType), http.StatusBadRequest)
	}
	metric.MType = metricType
	metric.ID = metricName
	return &metric, nil
}
