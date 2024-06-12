package handlers

import (
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateMetricHandler(t *testing.T) {
	type wantResp struct {
		status     int
		metricsMap map[string]interface{}
	}
	tests := []struct {
		name   string
		target string
		want   wantResp
	}{
		{
			name:   "status 200",
			target: "/update/gauge/test_name/1",
			want: wantResp{
				status: 200,
				metricsMap: map[string]interface{}{
					"test_name": storage.Gauge(1),
				},
			},
		},
		{
			name:   "status 404",
			target: "/update/gauge/1",
			want: wantResp{
				status:     404,
				metricsMap: map[string]interface{}{},
			},
		},
		{
			name:   "status 400 with invalid_type",
			target: "/update/invalid_type/test_name/1",
			want: wantResp{
				status:     400,
				metricsMap: map[string]interface{}{},
			},
		},
		{
			name:   "status 400 with invalid_value",
			target: "/update/gauge/test_name/invalid_value",
			want: wantResp{
				status:     400,
				metricsMap: map[string]interface{}{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, test.target, nil)
			w := httptest.NewRecorder()
			metricsStorage := storage.NewMetricsStorage()
			UpdateMetricHandler(metricsStorage)(w, r)
			res := w.Result()
			assert.Equal(t, test.want.status, res.StatusCode)
			assert.Equal(t, test.want.metricsMap, metricsStorage.Metrics)
		})
	}
}
