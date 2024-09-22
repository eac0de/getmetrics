package memstore

import (
	"context"
	stderr "errors"
	"net/http"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/pkg/errors"
)

func (store *MemoryStore) SaveMetric(ctx context.Context, metric models.Metric) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	switch metric.MType {
	case models.Gauge:
		store.MetricsData.Gauge[metric.ID] = *metric.Value
	case models.Counter:
		store.MetricsData.Counter[metric.ID] += *metric.Delta
	}
	return nil
}

func (store *MemoryStore) SaveMetrics(ctx context.Context, metricsList []models.Metric) error {
	var errsList []error
	for _, metric := range metricsList {
		err := store.SaveMetric(ctx, metric)
		if err != nil {
			errsList = append(errsList, err)
		}
	}
	err := stderr.Join(errsList...)
	if err != nil {
		return err
	}
	return nil
}

func (store *MemoryStore) GetMetric(ctx context.Context, metricName string, metricType string) (*models.Metric, error) {
	metric := models.Metric{
		ID:    metricName,
		MType: metricType,
	}
	switch metricType {
	case models.Gauge:
		value, ok := store.MetricsData.Gauge[metric.ID]
		if !ok {
			return nil, errors.NewErrorWithHTTPStatus(
				nil,
				"Metric not found",
				http.StatusNotFound,
			)
		}
		metric.Value = &value
	case models.Counter:
		delta, ok := store.MetricsData.Counter[metric.ID]
		if !ok {
			return nil, errors.NewErrorWithHTTPStatus(
				nil,
				"Metric not found",
				http.StatusNotFound,
			)
		}
		metric.Delta = &delta
	}
	return &metric, nil
}

func (store *MemoryStore) ListAllMetrics(ctx context.Context) ([]*models.Metric, error) {
	var metrics []*models.Metric
	for name, value := range store.MetricsData.Gauge {
		metric := models.Metric{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		}
		metrics = append(metrics, &metric)
	}
	for name, delta := range store.MetricsData.Counter {
		metric := models.Metric{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		}
		metrics = append(metrics, &metric)
	}
	return metrics, nil
}
