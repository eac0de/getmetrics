package handlers

import (
	"github.com/eac0de/getmetrics/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

func UpdateMetricHandler(m *storage.MetricsStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/update/")
		parts := strings.Split(path, "/")
		if len(parts) != 3 {
			http.Error(w, "Invalid URI", http.StatusNotFound)
			return
		}
		metricType := parts[0]
		metricName := parts[1]
		metricValue := parts[2]

		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		if metricType == "counter" {
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			oldValue := m.Get(metricName)
			if oldValue != nil {
				value = value + oldValue.(int64)
			}
			m.Save(metricName, storage.Counter(value))
		} else if metricType == "gauge" {
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			m.Save(metricName, storage.Gauge(value))
		} else {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}
