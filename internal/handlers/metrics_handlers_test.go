package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricHandler(t *testing.T) {
	type wantResp struct {
		status  int
		metrics models.MetricsMap
	}
	tests := []struct {
		name    string
		context *chi.Context
		want    wantResp
	}{
		{
			name: "status 200",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", "gauge")
				rctx.URLParams.Add("metricName", "test_name")
				rctx.URLParams.Add("metricValue", "1")
				return rctx
			}(),
			want: wantResp{
				status: 200,
				metrics: models.MetricsMap{
					Gauge:   map[string]float64{"test_name": 1},
					Counter: map[string]int64{},
				},
			},
		},
		{
			name: "status 404",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", "gauge")
				rctx.URLParams.Add("metricName", "")
				rctx.URLParams.Add("metricValue", "1")
				return rctx
			}(),
			want: wantResp{
				status: 404,
				metrics: models.MetricsMap{
					Gauge:   map[string]float64{},
					Counter: map[string]int64{},
				},
			},
		},
		{
			name: "status 400 with invalid_type",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", "invalid_type")
				rctx.URLParams.Add("metricName", "test_name")
				rctx.URLParams.Add("metricValue", "1")
				return rctx
			}(),
			want: wantResp{
				status: 400,
				metrics: models.MetricsMap{
					Gauge:   map[string]float64{},
					Counter: map[string]int64{},
				},
			},
		},
		{
			name: "status 400 with invalid_value",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", "gauge")
				rctx.URLParams.Add("metricName", "test_name")
				rctx.URLParams.Add("metricValue", "invalid_value")
				return rctx
			}(),
			want: wantResp{
				status: 400,
				metrics: models.MetricsMap{
					Gauge:   map[string]float64{},
					Counter: map[string]int64{},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "/update/{metricType}/{metricName}/{metricValue}"
			r := httptest.NewRequest(http.MethodPost, url, nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, test.context))
			w := httptest.NewRecorder()
			metricsStorage := storage.NewMemoryStorage()
			mhs := NewMetricsHandlerService(metricsStorage)
			mhs.UpdateMetricHandler()(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, test.want.status, resp.StatusCode)
			assert.Equal(t, test.want.metrics, metricsStorage.MetricsMap)
		})
	}
}

func TestUpdateMetricJSONHandler(t *testing.T) {
	type wantResp struct {
		status  int
		metrics models.MetricsMap
	}
	tests := []struct {
		name   string
		metric models.Metrics
		want   wantResp
	}{
		{
			name: "status 200",
			metric: models.Metrics{
				ID:    "test_name",
				MType: "gauge",
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: wantResp{
				status: 200,
				metrics: models.MetricsMap{
					Gauge:   map[string]float64{"test_name": 1},
					Counter: map[string]int64{},
				},
			},
		},

		{
			name: "status 404",
			metric: models.Metrics{
				MType: "gauge",
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: wantResp{
				status: 404,
				metrics: models.MetricsMap{
					Gauge:   map[string]float64{},
					Counter: map[string]int64{},
				},
			},
		},
		{
			name: "status 400 with invalid_type",
			metric: models.Metrics{
				ID:    "test_name",
				MType: "counter",
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: wantResp{
				status: 400,
				metrics: models.MetricsMap{
					Gauge:   map[string]float64{},
					Counter: map[string]int64{},
				},
			},
		},
		{
			name: "status 400 with invalid_value",
			metric: models.Metrics{
				ID:    "test_name",
				MType: "gauge",
			},
			want: wantResp{
				status: 400,
				metrics: models.MetricsMap{
					Gauge:   map[string]float64{},
					Counter: map[string]int64{},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "/update/"
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(test.metric)
			if err != nil {
				log.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodPost, url, &buf)
			w := httptest.NewRecorder()
			metricsStorage := storage.NewMemoryStorage()
			mhs := NewMetricsHandlerService(metricsStorage)
			mhs.UpdateMetricJSONHandler()(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, test.want.status, resp.StatusCode)
			assert.Equal(t, test.want.metrics, metricsStorage.MetricsMap)
		})
	}
}
