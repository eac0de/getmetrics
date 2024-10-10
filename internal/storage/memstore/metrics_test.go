package memstore

import (
	"context"
	"testing"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSaveMetric(t *testing.T) {
	t.Run("save gauge metric", func(t *testing.T) {
		store := New()
		err := store.SaveMetric(
			context.Background(),
			models.Metric{
				ID:    "test_gauge",
				MType: models.Gauge,
				Value: func(v float64) *float64 { return &v }(1),
			},
		)
		assert.NoError(t, err)
	})
	t.Run("save counter metric", func(t *testing.T) {
		store := New()
		err := store.SaveMetric(
			context.Background(),
			models.Metric{
				ID:    "test_counter",
				MType: models.Counter,
				Delta: func(v int64) *int64 { return &v }(2),
			},
		)
		assert.NoError(t, err)
	})
}

func TestSaveMetrics(t *testing.T) {
	t.Run("save metrics", func(t *testing.T) {
		store := New()
		err := store.SaveMetrics(
			context.Background(),
			[]models.Metric{
				{
					ID:    "test_gauge_1",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(1),
				},
				{
					ID:    "test_gauge_2",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(2),
				},
			},
		)
		assert.NoError(t, err)
	})
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricType  string
		metric      *models.Metric
		metricsData models.MetricsData
		errMsg      string
	}{
		{
			name:       "gauge_success",
			metricName: "test_gauge",
			metricType: models.Gauge,
			metric: &models.Metric{
				ID:    "test_gauge",
				MType: models.Gauge,
				Value: func(v float64) *float64 { return &v }(2),
			},
			metricsData: models.MetricsData{
				Counter: map[string]int64{},
				Gauge: map[string]float64{
					"test_gauge": 2,
				},
			},
		},
		{
			name:       "counter_success",
			metricName: "test_counter",
			metricType: models.Counter,
			metric: &models.Metric{
				ID:    "test_counter",
				MType: models.Counter,
				Delta: func(v int64) *int64 { return &v }(3),
			},
			metricsData: models.MetricsData{
				Counter: map[string]int64{
					"test_counter": 3,
				},
				Gauge: map[string]float64{},
			},
		},
		{
			name:        "gauge_error",
			metricName:  "test_gauge",
			metricType:  models.Gauge,
			metricsData: models.MetricsData{},
			errMsg:      "Metric not found",
		},
		{
			name:        "counter_error",
			metricName:  "test_counter",
			metricType:  models.Counter,
			metricsData: models.MetricsData{},
			errMsg:      "Metric not found",
		},
	}
	for _, test := range tests {
		store := New()
		store.MetricsData = test.metricsData
		t.Run(test.name, func(t *testing.T) {
			metric, err := store.GetMetric(
				context.Background(),
				test.metricName,
				test.metricType,
			)
			if metric != nil {
				assert.Equal(t, metric.ID, test.metric.ID)
				assert.Equal(t, metric.MType, test.metric.MType)
				assert.Equal(t, metric.Value, test.metric.Value)
				assert.Equal(t, metric.Delta, test.metric.Delta)
			}
			if err != nil {
				assert.Equal(t, err.Error(), test.errMsg)
			}
		})
	}
}

func TestListAllMetrics(t *testing.T) {
	store := New()
	store.MetricsData = models.MetricsData{
		Counter: map[string]int64{
			"test_counter": 3,
		},
		Gauge: map[string]float64{
			"test_gauge": 5,
		},
	}
	t.Run("success", func(t *testing.T) {
		_, err := store.ListAllMetrics(context.Background())
		assert.NoError(t, err)
	})

}
