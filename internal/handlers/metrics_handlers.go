package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/eac0de/getmetrics/pkg/hasher"
	"github.com/go-chi/chi/v5"
)

type MetricsHandlerService struct {
	metricsStorage storage.MetricsStorer
	key            string
}

func NewMetricsHandlerService(ms storage.MetricsStorer, key string) *MetricsHandlerService {
	return &MetricsHandlerService{
		metricsStorage: ms,
		key:            key,
	}
}

func (mhs *MetricsHandlerService) UpdateMetricHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		metric := models.Metrics{
			ID:    metricName,
			MType: metricType,
		}
		var value interface{}
		switch metric.MType {
		case models.Counter:
			Delta, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid value for counter metric", http.StatusBadRequest)
				return
			}
			value = Delta
			metric.Delta = &Delta
		case models.Gauge:
			Value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid value for gauge metric", http.StatusBadRequest)
				return
			}
			value = Value
			metric.Value = &Value
		}
		_, err := mhs.metricsStorage.Save(r.Context(), metric)
		if err != nil {
			var ewhs *storage.ErrorWithHTTPStatus
			if errors.As(err, &ewhs) {
				http.Error(w, ewhs.Error(), ewhs.HTTPStatus)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		metricStr := fmt.Sprintf("%v", value)
		data := []byte(metricStr)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		hash := hasher.HashSumToString(data, mhs.key)
		if hash != "" {
			w.Header().Set("HashSHA256", hash)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (mhs *MetricsHandlerService) UpdateMetricJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newMetric, err := mhs.metricsStorage.Save(r.Context(), metric)
		if err != nil {
			var ewhs *storage.ErrorWithHTTPStatus
			if errors.As(err, &ewhs) {
				http.Error(w, ewhs.Error(), ewhs.HTTPStatus)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		metricJSON, _ := json.Marshal(newMetric)
		w.Header().Set("Content-Type", "application/json")
		hash := hasher.HashSumToString(metricJSON, mhs.key)
		if hash != "" {
			w.Header().Set("HashSHA256", hash)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(metricJSON)
	}
}

func (mhs *MetricsHandlerService) UpdateManyMetricsJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metricList []models.Metrics
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &metricList); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newMetricsList, err := mhs.metricsStorage.SaveMany(r.Context(), metricList)
		if err != nil {
			var ewhs *storage.ErrorWithHTTPStatus
			if errors.As(err, &ewhs) {
				http.Error(w, ewhs.Error(), ewhs.HTTPStatus)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		metricsJSON, _ := json.Marshal(newMetricsList)
		w.Header().Set("Content-Type", "application/json")
		hash := hasher.HashSumToString(metricsJSON, mhs.key)
		if hash != "" {
			w.Header().Set("HashSHA256", hash)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(metricsJSON)
	}
}

func (mhs *MetricsHandlerService) GetMetricHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		metric, err := mhs.metricsStorage.Get(r.Context(), metricName, metricType)
		if err != nil {
			var ewhs *storage.ErrorWithHTTPStatus
			if errors.As(err, &ewhs) {
				http.Error(w, ewhs.Error(), ewhs.HTTPStatus)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		var value interface{}
		switch metric.MType {
		case models.Counter:
			value = *metric.Delta
		case models.Gauge:
			value = *metric.Value
		}
		metricStr := fmt.Sprintf("%v", value)
		data := []byte(metricStr)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		hash := hasher.HashSumToString(data, mhs.key)
		if hash != "" {
			w.Header().Set("HashSHA256", hash)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (mhs *MetricsHandlerService) GetMetricJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newMetric models.Metrics
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &newMetric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		metricName := newMetric.ID
		metricType := newMetric.MType
		metric, err := mhs.metricsStorage.Get(r.Context(), metricName, metricType)
		if err != nil {
			var ewhs *storage.ErrorWithHTTPStatus
			if errors.As(err, &ewhs) {
				http.Error(w, ewhs.Error(), ewhs.HTTPStatus)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		metricJSON, _ := json.Marshal(metric)
		w.Header().Set("Content-Type", "application/json")
		hash := hasher.HashSumToString(metricJSON, mhs.key)
		if hash != "" {
			w.Header().Set("HashSHA256", hash)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(metricJSON)
	}
}

func (mhs *MetricsHandlerService) ShowMetricsSummaryHandler() func(http.ResponseWriter, *http.Request) {
	fpath := filepath.Join("templates", "metrics_summary.html")
	file, err := os.OpenFile(fpath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("read template error(1): %s", err.Error())
	}
	temp, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("read template error(2): %s", err.Error())
	}
	metricsTemplate := string(temp)
	tmpl := template.Must(template.New("metrics").Parse(metricsTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := mhs.metricsStorage.GetAll(r.Context())
		if err != nil {
			var ewhs *storage.ErrorWithHTTPStatus
			if errors.As(err, &ewhs) {
				http.Error(w, ewhs.Error(), ewhs.HTTPStatus)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
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

func (mhs *MetricsHandlerService) PingHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := mhs.metricsStorage.Ping(r.Context()); err != nil {
			var ewhs *storage.ErrorWithHTTPStatus
			if errors.As(err, &ewhs) {
				http.Error(w, ewhs.Error(), ewhs.HTTPStatus)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
