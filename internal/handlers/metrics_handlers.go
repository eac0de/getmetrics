package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

type metricsHandlerService struct {
	metricsStore MetricsStorer
}

func NewMetricsHandlerService(m MetricsStorer) *metricsHandlerService {
	return &metricsHandlerService{
		metricsStore: m,
	}
}

func (mhs *metricsHandlerService) UpdateMetricHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}
		metric, err := mhs.metricsStore.Save(metricType, metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var value interface{}
		switch metric.MType {
		case storage.Counter:

			value = *metric.Delta
		case storage.Gauge:
			value = *metric.Value
		}
		metricStr := fmt.Sprintf("%v", value)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metricStr))
	}
}

func (mhs *metricsHandlerService) UpdateMetricJSONHandler() func(http.ResponseWriter, *http.Request) {
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
		metricType := newMetric.MType
		metricName := newMetric.ID

		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}
		var metricValue interface{}
		switch metricType {
		case storage.Counter:
			if newMetric.Delta == nil {
				http.Error(w, "for metric type counter field delta is required", http.StatusBadRequest)
				return
			}
			metricValue = *newMetric.Delta
		case storage.Gauge:
			if newMetric.Value == nil {
				http.Error(w, "for metric type gauge field value is required", http.StatusBadRequest)
				return
			}
			metricValue = *newMetric.Value
		default:
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
		metric, err := mhs.metricsStore.Save(metricType, metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		metricJSON, _ := json.Marshal(metric)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(metricJSON)
	}
}

func (mhs *metricsHandlerService) GetMetricHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}
		metric := mhs.metricsStore.Get(metricType, metricName)
		errorMessage := fmt.Sprintf("metric %s not found", metricName)
		if metric == nil {
			http.Error(w, errorMessage, http.StatusNotFound)
			return
		}
		var value interface{}
		switch metric.MType {
		case storage.Counter:
			value = *metric.Delta
		case storage.Gauge:
			value = *metric.Value
		}
		metricStr := fmt.Sprintf("%v", value)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metricStr))
	}
}

func (mhs *metricsHandlerService) GetMetricJSONHandler() func(http.ResponseWriter, *http.Request) {
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
		metricType := newMetric.MType
		metricName := newMetric.ID

		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}
		metric := mhs.metricsStore.Get(metricType, metricName)
		errorMessage := fmt.Sprintf("metric %s not found", metricName)
		if metric == nil {
			http.Error(w, errorMessage, http.StatusNotFound)
			return
		}
		metricJSON, _ := json.Marshal(metric)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(metricJSON)
	}
}

func (mhs *metricsHandlerService) ShowMetricsSummaryHandler() func(http.ResponseWriter, *http.Request) {
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
		metrics := mhs.metricsStore.GetAll()
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].ID < metrics[j].ID
		})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err := tmpl.Execute(w, metrics)
		if err != nil {
			http.Error(w, "error rendering template", http.StatusInternalServerError)
			return
		}
	}
}
