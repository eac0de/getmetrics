package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricHandler(t *testing.T) {
	type wantResp struct {
		status     int
		metricsMap map[string]interface{}
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
				metricsMap: map[string]interface{}{
					"test_name": storage.Gauge(1),
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
				status:     404,
				metricsMap: map[string]interface{}{},
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
				status:     400,
				metricsMap: map[string]interface{}{},
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
				status:     400,
				metricsMap: map[string]interface{}{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "/update/{metricType}/{metricName}/{metricValue}"
			r := httptest.NewRequest(http.MethodPost, url, nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, test.context))
			w := httptest.NewRecorder()
			metricsStorage := storage.NewMetricsStorage()
			UpdateMetricHandler(metricsStorage)(w, r)
			resp := w.Result()
			func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Printf("Failed to close response body: %v\n", err)
				}
			}(resp.Body)
			assert.Equal(t, test.want.status, resp.StatusCode)
			assert.Equal(t, test.want.metricsMap, metricsStorage.Metrics)
		})
	}
}
