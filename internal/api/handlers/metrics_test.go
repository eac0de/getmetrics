package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/mocks"
	"github.com/eac0de/getmetrics/pkg/errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type (
	want struct {
		statusCode int
		metric     models.Metric
		text       string
	}
	saveMetricReturn struct {
		metric *models.Metric
		err    error
	}
)

func TestUpdateMetricHandler(t *testing.T) {
	tests := []struct {
		name    string
		context *chi.Context
		want    want
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
			want: want{
				statusCode: http.StatusOK,
				text:       "1",
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
			want: want{
				statusCode: http.StatusOK,
				text:       "6",
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
			want: want{
				statusCode: http.StatusNotFound,
				text:       "metric name is required\n",
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
			want: want{
				statusCode: http.StatusBadRequest,
				text:       "invalid metric type for test_name: invalid_type\n",
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
			want: want{
				statusCode: http.StatusBadRequest,
				text:       "metric test_name with type gauge must have filled value\n",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	metricsStore.EXPECT().SaveMetric(
		gomock.Any(),
		gomock.Any(),
	).Return(nil).AnyTimes()
	delta := int64(5)
	metricsStore.EXPECT().GetMetric(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(&models.Metric{
		Delta: &delta},
		nil,
	).AnyTimes()
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
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.text, string(body))
		})
	}
}

func TestUpdateMetricJSONHandler(t *testing.T) {
	tests := []struct {
		name    string
		reqBody models.Metric
		want    want
	}{
		{
			name: "gauge status 200",
			reqBody: models.Metric{
				ID:    "test_name",
				MType: models.Gauge,
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: want{
				statusCode: http.StatusOK,
				metric: models.Metric{
					ID:    "test_name",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(1),
				},
			},
		},
		{
			name: "counter status 200",
			reqBody: models.Metric{
				ID:    "test_name",
				MType: models.Counter,
				Delta: func(v int64) *int64 { return &v }(1),
			},
			want: want{
				statusCode: 200,
				metric: models.Metric{
					ID:    "test_name",
					MType: models.Counter,
					Delta: func(v int64) *int64 { return &v }(6),
				},
			},
		},

		{
			name: "status 400 without ID",
			reqBody: models.Metric{
				MType: models.Gauge,
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: want{
				statusCode: http.StatusBadRequest,
				text:       "metric name is required\n",
			},
		},
		{
			name: "counter status 400 with invalid_type",
			reqBody: models.Metric{
				ID:    "test_name",
				MType: "counter",
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: want{
				statusCode: http.StatusBadRequest,
				text:       "metric test_name with type counter must have filled delta\n",
			},
		},
		{
			name: "gauge status 400 with invalid_value",
			reqBody: models.Metric{
				ID:    "test_name",
				MType: models.Gauge,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				text:       "metric test_name with type gauge must have filled value\n",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	metricsStore.EXPECT().SaveMetric(
		gomock.Any(),
		gomock.Any(),
	).Return(nil).AnyTimes()
	delta := int64(5)
	metricsStore.EXPECT().GetMetric(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(
		&models.Metric{Delta: &delta},
		nil,
	).AnyTimes()
	mh := NewMetricsHandlers(metricsStore, "")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "/update/"
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(test.reqBody)
			if err != nil {
				log.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodPost, url, &buf)
			w := httptest.NewRecorder()
			mh.UpdateMetricJSONHandler()(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
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
				assert.Equal(t, test.want.text, string(body))
			}
		})
	}
	t.Run("invalid request body", func(t *testing.T) {
		url := "/update/"
		buf := bytes.NewBufferString("invalid request body")
		r := httptest.NewRequest(http.MethodPost, url, buf)
		w := httptest.NewRecorder()
		mh.UpdateMetricJSONHandler()(w, r)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Invalid request payload\n", string(body))
	})
}

func TestUpdateMetricsJSONHandler(t *testing.T) {
	type want struct {
		statusCode  int
		metricsList []models.Metric
		text        string
	}
	tests := []struct {
		name    string
		reqBody []models.Metric
		want    want
	}{
		{
			name: "status 200",
			reqBody: []models.Metric{
				{
					ID:    "test_counter",
					MType: models.Counter,
					Delta: func(v int64) *int64 { return &v }(1),
				},
				{
					ID:    "test_gauge",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(2),
				},
				{
					ID:    "test_counter",
					MType: models.Counter,
					Delta: func(v int64) *int64 { return &v }(3),
				},
				{
					ID:    "test_gauge",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(4),
				},
			},
			want: want{
				statusCode: http.StatusOK,
				metricsList: []models.Metric{
					{
						ID:    "test_gauge",
						MType: models.Gauge,
						Value: func(v float64) *float64 { return &v }(4),
					},
					{
						ID:    "test_counter",
						MType: models.Counter,
						Delta: func(v int64) *int64 { return &v }(9),
					},
				},
			},
		},
		{
			name: "status 400 without ID",
			reqBody: []models.Metric{
				{
					MType: models.Counter,
					Delta: func(v int64) *int64 { return &v }(1),
				},
				{
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(2),
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				text:       "metric name is required\nmetric name is required\n",
			},
		},
		{
			name: "status 400 with invalid_type",
			reqBody: []models.Metric{
				{
					ID:    "test_counter",
					MType: "invalid_type",
					Delta: func(v int64) *int64 { return &v }(1),
				},
				{
					ID:    "test_gauge",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(2),
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				text:       "invalid metric type for test_counter: invalid_type\n",
			},
		},
		{
			name: "status 400 with invalid_value",
			reqBody: []models.Metric{
				{
					ID:    "test_counter",
					MType: models.Counter,
				},
				{
					ID:    "test_gauge",
					MType: models.Gauge,
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				text:       "metric test_counter with type counter must have filled delta\nmetric test_gauge with type gauge must have filled value\n",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	metricsStore.EXPECT().SaveMetrics(
		gomock.Any(),
		gomock.Any(),
	).Return(nil).AnyTimes()
	delta := int64(5)
	metricsStore.EXPECT().GetMetric(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(
		&models.Metric{Delta: &delta},
		nil,
	).AnyTimes()
	mh := NewMetricsHandlers(metricsStore, "")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "/updates/"
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(test.reqBody)
			if err != nil {
				log.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodPost, url, &buf)
			w := httptest.NewRecorder()
			mh.UpdateMetricsJSONHandler()(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			if resp.StatusCode < 400 {
				var metricsList []models.Metric
				if err := json.NewDecoder(resp.Body).Decode(&metricsList); err != nil {
					assert.NoError(t, err)
				}
				for _, wantMetric := range test.want.metricsList {
					for _, metric := range metricsList {
						if metric.ID == wantMetric.ID && metric.MType == wantMetric.MType {
							assert.Equal(t, metric.Value, metric.Value)
							assert.Equal(t, metric.Delta, metric.Delta)
							break
						}
					}
				}
			} else {
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, test.want.text, string(body))
			}
		})
	}
	t.Run("invalid request body", func(t *testing.T) {
		url := "/updates/"
		buf := bytes.NewBufferString("invalid request body")
		r := httptest.NewRequest(http.MethodPost, url, buf)
		w := httptest.NewRecorder()
		mh.UpdateMetricJSONHandler()(w, r)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Invalid request payload\n", string(body))
	})
}

func TestGetMetricHandler(t *testing.T) {
	type wantResp struct {
		status   int
		respBody string
	}
	tests := []struct {
		name        string
		context     *chi.Context
		metricStore *models.Metric
		errStore    error
		want        wantResp
	}{
		{
			name: "gauge status 200",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", models.Gauge)
				rctx.URLParams.Add("metricName", "test_name")
				return rctx
			}(),
			metricStore: &models.Metric{
				ID:    "test_name",
				MType: models.Gauge,
				Value: func(v float64) *float64 { return &v }(1),
			},
			want: wantResp{
				status:   200,
				respBody: "1",
			},
		},
		{
			name: "status 404",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", models.Gauge)
				rctx.URLParams.Add("metricName", "")
				return rctx
			}(),
			errStore: errors.NewErrorWithHTTPStatus(
				nil,
				"Metric not found",
				http.StatusNotFound,
			),
			want: wantResp{
				status:   404,
				respBody: "Metric not found\n",
			},
		},
		{
			name: "status 404 with invalid_type",
			context: func() *chi.Context {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", "invalid_type")
				rctx.URLParams.Add("metricName", "test_name")
				return rctx
			}(),
			errStore: errors.NewErrorWithHTTPStatus(
				nil,
				"Metric not found",
				http.StatusNotFound,
			),
			want: wantResp{
				status:   404,
				respBody: "Metric not found\n",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	mh := NewMetricsHandlers(metricsStore, "")
	for _, test := range tests {
		metricsStore.EXPECT().GetMetric(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.metricStore, test.errStore)
		t.Run(test.name, func(t *testing.T) {
			url := "/value/{metricType}/{metricName}"
			r := httptest.NewRequest(http.MethodPost, url, nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, test.context))
			w := httptest.NewRecorder()
			mh.GetMetricHandler()(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			assert.Equal(t, test.want.status, resp.StatusCode)
			assert.Equal(t, test.want.respBody, string(body))
		})
	}
}

func TestGetMetricJSONHandler(t *testing.T) {
	tests := []struct {
		name             string
		reqBody          models.Metric
		saveMetricReturn saveMetricReturn
		want             want
	}{
		{
			name: "gauge status 200",
			reqBody: models.Metric{
				ID:    "test_name",
				MType: models.Gauge,
			},
			saveMetricReturn: saveMetricReturn{
				metric: &models.Metric{
					ID:    "test_name",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(1),
				},
			},
			want: want{
				statusCode: http.StatusOK,
				metric: models.Metric{
					ID:    "test_name",
					MType: models.Gauge,
					Value: func(v float64) *float64 { return &v }(1),
				},
			},
		},
		{
			name: "status 404 without ID",
			reqBody: models.Metric{
				MType: models.Gauge,
			},
			saveMetricReturn: saveMetricReturn{
				err: errors.NewErrorWithHTTPStatus(
					nil,
					"Metric not found",
					http.StatusNotFound,
				),
			},
			want: want{
				statusCode: http.StatusNotFound,
				text:       "Metric not found\n",
			},
		},
		{
			name: "status 404 with invalid_type",
			reqBody: models.Metric{
				ID:    "test_name",
				MType: "invalid_type",
			},
			saveMetricReturn: saveMetricReturn{
				err: errors.NewErrorWithHTTPStatus(
					nil,
					"Metric not found",
					http.StatusNotFound,
				),
			},
			want: want{
				statusCode: http.StatusNotFound,
				text:       "Metric not found\n",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	mh := NewMetricsHandlers(metricsStore, "")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsStore.EXPECT().GetMetric(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(
				test.saveMetricReturn.metric,
				test.saveMetricReturn.err,
			)
			url := "/value/"
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(test.reqBody)
			if err != nil {
				log.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodPost, url, &buf)
			w := httptest.NewRecorder()
			mh.GetMetricJSONHandler()(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			if resp.StatusCode < 400 {
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
				assert.Equal(t, test.want.text, string(body))
			}
		})
	}
}

func ExampleMetricsHandlers_UpdateMetricHandler() {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", models.Counter)
	rctx.URLParams.Add("metricName", "test_name")
	rctx.URLParams.Add("metricValue", "1")

	req, _ := http.NewRequest(http.MethodPost, "/update/{metricType}/{metricName}/{metricValue}", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	metricsStore.EXPECT().SaveMetric(gomock.Any(), gomock.Any()).Return(nil)
	delta := int64(5)
	metricsStore.EXPECT().GetMetric(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.Metric{Delta: &delta}, nil)
	mh := NewMetricsHandlers(metricsStore, "")

	mh.UpdateMetricHandler()(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	fmt.Println(res.StatusCode)

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))

	// Output:
	// 200
	// 6
}

func ExampleMetricsHandlers_GetMetricHandler() {
	metricName := "test_name"
	metricType := models.Counter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", metricType)
	rctx.URLParams.Add("metricName", metricName)
	req, _ := http.NewRequest(http.MethodGet, "/value/{metricType}/{metricName}", nil)
	rr := httptest.NewRecorder()

	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	metricsStore := mocks.NewMockIMetricsStore(ctrl)
	delta := int64(5)
	metric := models.Metric{
		ID:    metricName,
		MType: metricType,
		Delta: &delta,
	}
	metricsStore.EXPECT().GetMetric(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metric, nil)
	mh := NewMetricsHandlers(metricsStore, "")

	mh.GetMetricHandler()(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	fmt.Println(res.StatusCode)

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))

	// Output:
	// 200
	// 5
}
