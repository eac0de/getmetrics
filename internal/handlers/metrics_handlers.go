package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func UpdateMetricHandler(m storage.MetricsStorer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}
		metric, err := m.Save(metricType, metricName, metricValue)
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

func UpdateMetricJSONHandler(m storage.MetricsStorer) func(http.ResponseWriter, *http.Request) {
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
		metric, err := m.Save(metricType, metricName, metricValue)
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

func GetMetricHandler(m storage.MetricsStorer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}
		metric := m.Get(metricType, metricName)
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

func GetMetricJSONHandler(m storage.MetricsStorer) func(http.ResponseWriter, *http.Request) {
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
		metric := m.Get(metricType, metricName)
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

type Metric struct {
	Name  string
	Value interface{}
}

type MetricsData struct {
	Metrics []Metric
}

const metricsTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metric Summary</title>
    <style>
        body {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            font-family: Arial, sans-serif;
            background-color: #f0f0f0;
        }
        .container {
            height: 80%;
            text-align: center;
            background-color: #fff;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
            overflow-y: auto;
        }
        .metrics {
            margin-top: 20px;
        }
        h1 {
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Metric Summary</h1>
        <div class="metrics">
            {{range .Metrics}}
                <p><strong>{{.Name}}</strong> - {{.Value}}</p>
            {{end}}
        </div>
    </div>
</body>
</html>
`

func ShowMetricsSummaryHandler(m storage.MetricsStorer) func(http.ResponseWriter, *http.Request) {
	tmpl := template.Must(template.New("metrics").Parse(metricsTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		data := MetricsData{}
		for _, metric := range m.GetAll() {
			for metricName, metricValue := range metric {
				data.Metrics = append(data.Metrics, Metric{Name: metricName, Value: metricValue})
			}
		}
		sort.Slice(data.Metrics, func(i, j int) bool {
			return data.Metrics[i].Name < data.Metrics[j].Name
		})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "error rendering template", http.StatusInternalServerError)
			return
		}
	}
}
