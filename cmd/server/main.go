package main

import (
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

type MemStorage struct {
	Metrics map[string]interface{}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]interface{}),
	}
}

func (m *MemStorage) SaveMetric(metricName string, metricValue interface{}) {
	m.Metrics[metricName] = metricValue
}

func (m *MemStorage) GetMetric(metricName string) interface{} {
	value, ok := m.Metrics[metricName]
	if !ok {
		return nil
	}
	return value
}

func (m *MemStorage) UpdateHandler(w http.ResponseWriter, r *http.Request) {
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
		oldValue := m.GetMetric(metricName)
		if oldValue != nil {
			value = value + oldValue.(int64)
		}
		m.SaveMetric(metricName, value)
	} else if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		m.SaveMetric(metricName, value)
	} else {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func main() {
	storage := NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", storage.UpdateHandler)
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		return
	}
}
