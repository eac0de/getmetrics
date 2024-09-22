package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricHandler(t *testing.T) {
	type wantResp struct {
		status   int
		respBody string
	}
	tests := []struct {
		name    string
		context *chi.Context
		want    wantResp
	}{
		{
			name: "gauge status 200",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", models.Gauge)
				rctx.URLParams.Add("metricName", "test_name")
				rctx.URLParams.Add("metricValue", "1")
				return rctx
			}(),
			want: wantResp{
				status:   200,
				respBody: "1",
			},
		},
		{
			name: "counter status 200",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", models.Counter)
				rctx.URLParams.Add("metricName", "test_name")
				rctx.URLParams.Add("metricValue", "1")
				return rctx
			}(),
			want: wantResp{
				status:   200,
				respBody: "6",
			},
		},
		{
			name: "status 404",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", models.Gauge)
				rctx.URLParams.Add("metricName", "")
				rctx.URLParams.Add("metricValue", "1")
				return rctx
			}(),
			want: wantResp{
				status:   404,
				respBody: "metric name is required\n",
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
				status:   400,
				respBody: "invalid metric type for test_name: invalid_type\n",
			},
		},
		{
			name: "status 400 with invalid_value",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", models.Gauge)
				rctx.URLParams.Add("metricName", "test_name")
				rctx.URLParams.Add("metricValue", "invalid_value")
				return rctx
			}(),
			want: wantResp{
				status:   400,
				respBody: "metric test_name with type gauge must have filled value\n",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	metricsStore.EXPECT().SaveMetric(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	delta := int64(5)
	metricsStore.EXPECT().GetMetric(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.Metric{Delta: &delta}, nil).AnyTimes()
	mh := NewMetricsHandlers(metricsStore, "")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "/update/{metricType}/{metricName}/{metricValue}"
			r := httptest.NewRequest(http.MethodPost, url, nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, test.context))
			w := httptest.NewRecorder()
			mh.UpdateMetricHandler()(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			assert.Equal(t, test.want.status, resp.StatusCode)
			assert.Equal(t, test.want.respBody, string(body))
		})
	}
}

func TestUpdateMetricJSONHandler(t *testing.T) {
	type wantResp struct {
		status int
		metric models.Metric
		msg    string
	}
	tests := []struct {
		name   string
		metric models.Metric
		want   wantResp
	}{
		{
			name: "gauge status 200",
			metric: models.Metric{
				ID:    "test_name",
				MType: models.Gauge,
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: wantResp{
				status: 200,
				metric: models.Metric{
					ID:    "test_name",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(1),
				},
			},
		},
		{
			name: "counter status 200",
			metric: models.Metric{
				ID:    "test_name",
				MType: models.Counter,
				Delta: func(v int64) *int64 { return &v }(1),
			},
			want: wantResp{
				status: 200,
				metric: models.Metric{
					ID:    "test_name",
					MType: models.Counter,
					Delta: func(v int64) *int64 { return &v }(6),
				},
			},
		},

		{
			name: "status 400 without ID",
			metric: models.Metric{
				MType: models.Gauge,
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: wantResp{
				status: 400,
				msg:    "metric name is required\n",
			},
		},
		{
			name: "counter status 400 with invalid_type",
			metric: models.Metric{
				ID:    "test_name",
				MType: "counter",
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: wantResp{
				status: 400,
				msg:    "metric test_name with type counter must have filled delta\n",
			},
		},
		{
			name: "gauge status 400 with invalid_value",
			metric: models.Metric{
				ID:    "test_name",
				MType: models.Gauge,
			},
			want: wantResp{
				status: 400,
				msg:    "metric test_name with type gauge must have filled value\n",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	metricsStore.EXPECT().SaveMetric(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	delta := int64(5)
	metricsStore.EXPECT().GetMetric(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.Metric{Delta: &delta}, nil).AnyTimes()
	mh := NewMetricsHandlers(metricsStore, "")
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
			mh.UpdateMetricJSONHandler()(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, test.want.status, resp.StatusCode)
			if resp.StatusCode == http.StatusOK {
				var metric models.Metric
				if err := json.NewDecoder(resp.Body).Decode(&metric); err != nil {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.want.metric.ID, metric.ID)
				assert.Equal(t, test.want.metric.MType, metric.MType)
				assert.Equal(t, test.want.metric.Value, metric.Value)
				assert.Equal(t, test.want.metric.Delta, metric.Delta)
			} else {
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, test.want.msg, string(body))
			}
		})
	}
}
