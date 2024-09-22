package router

import (
	"context"
	"encoding/json"
	stderr "errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/pkg/errors"
	"github.com/eac0de/getmetrics/pkg/hasher"
	"github.com/eac0de/getmetrics/pkg/middlewares"
	"github.com/go-chi/chi/v5"
)

type IDatabase interface {
	PingContext(ctx context.Context) error
}

type IMetricsStore interface {
	SaveMetric(ctx context.Context, metric models.Metric) error
	SaveMetrics(ctx context.Context, metricsList []models.Metric) error
	GetMetric(ctx context.Context, metricName string, metricType string) (*models.Metric, error)
	ListAllMetrics(ctx context.Context) ([]*models.Metric, error)
}

type Router struct {
	chi.Router
	MetricsStore IMetricsStore
	Database     IDatabase
	SecretKey    string
}

func New(metricsStore IMetricsStore, database IDatabase, secretKey string) *Router {
	r := &Router{
		Router:       chi.NewRouter(),
		MetricsStore: metricsStore,
		Database:     database,
		SecretKey:    secretKey,
	}

	r.Use(middlewares.LoggerMiddleware)
	r.Use(middlewares.GetCheckSignMiddleware(secretKey))
	contentTypesForCompress := "application/json text/html"
	r.Use(middlewares.GetGzipMiddleware(contentTypesForCompress))
	r.Get("/", r.ShowMetricsSummaryHandler())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", r.UpdateMetricHandler())
	r.Post("/update/", r.UpdateMetricJSONHandler())
	r.Post("/updates/", r.UpdateMetricsJSONHandler())

	r.Get("/value/{metricType}/{metricName}", r.GetMetricHandler())
	r.Post("/value/", r.GetMetricJSONHandler())
	r.Get("/ping", r.PingHandler())
	return r
}

func (router *Router) UpdateMetricHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		metric := models.Metric{
			ID:    metricName,
			MType: metricType,
		}
		switch metric.MType {
		case models.Counter:
			delta, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid value for counter metric", http.StatusBadRequest)
				return
			}
			metric.Delta = &delta
		case models.Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid value for gauge metric", http.StatusBadRequest)
				return
			}
			metric.Value = &value
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		err := router.MetricsStore.SaveMetric(r.Context(), metric)
		if err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		data := []byte(metricValue)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		router.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (router *Router) UpdateMetricJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metric
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		err := models.ValidateMetric(metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = router.MetricsStore.SaveMetric(r.Context(), metric)
		if err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		data, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, "Invalid server data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		router.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (router *Router) UpdateMetricsJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metricsList []models.Metric
		if err := json.NewDecoder(r.Body).Decode(&metricsList); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		var errsList []error
		for _, metric := range metricsList {
			err := models.ValidateMetric(metric)
			if err != nil {
				errsList = append(errsList, err)
			}
		}
		if len(errsList) > 0 {
			err := stderr.Join(errsList...)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		metricsList = models.MergeMetricsList(metricsList)
		err := router.MetricsStore.SaveMetrics(r.Context(), metricsList)
		if err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		data, err := json.Marshal(metricsList)
		if err != nil {
			http.Error(w, "Invalid server data", http.StatusInternalServerError)

			return
		}
		w.Header().Set("Content-Type", "application/json")
		router.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (router *Router) GetMetricHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		metric, err := router.MetricsStore.GetMetric(r.Context(), metricName, metricType)
		if err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		var metricStr string
		switch metric.MType {
		case models.Counter:
			metricStr = fmt.Sprintf("%v", *metric.Delta)
		case models.Gauge:
			metricStr = fmt.Sprintf("%v", *metric.Value)
		}
		data := []byte(metricStr)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		router.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (router *Router) GetMetricJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric *models.Metric
		if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		metric, err := router.MetricsStore.GetMetric(r.Context(), metric.ID, metric.MType)
		if err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		data, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, "Invalid server data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		router.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (router *Router) ShowMetricsSummaryHandler() func(http.ResponseWriter, *http.Request) {
	filePath := filepath.Join("templates", "metrics_summary.html")
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal("open template file error: %s", err.Error())
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("read template error: %s", err.Error())
	}
	tmpl, err := template.New("metrics").Parse(string(data))
	if err != nil {
		log.Fatal("parse template error: %s", err.Error())
	}
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := router.MetricsStore.ListAllMetrics(r.Context())
		if err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].ID < metrics[j].ID
		})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err = tmpl.Execute(w, metrics)
		if err != nil {
			http.Error(w, "Rendering template error", http.StatusInternalServerError)
			return
		}
	}
}

func (router *Router) PingHandler() func(http.ResponseWriter, *http.Request) {
	if router.Database == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Database not init", http.StatusInternalServerError)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := router.Database.PingContext(r.Context()); err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (router *Router) addSign(w http.ResponseWriter, data []byte) {
	hash := hasher.HashSumToString(data, router.SecretKey)
	if hash != "" {
		w.Header().Set("HashSHA256", hash)
	}
}
