package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func UpdateMetricHandler(m MetricsStorer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}

		if metricType == "counter" {
			i, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			value := storage.Counter(i)
			oldValue := m.Get(metricName)
			if _, ok := oldValue.(storage.Counter); ok {
				value = value + oldValue.(storage.Counter)
			}
			m.Save(metricName, storage.Counter(value))
		} else if metricType == "gauge" {
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "invalid gauge value", http.StatusBadRequest)
				return
			}
			m.Save(metricName, storage.Gauge(value))
		} else {
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

func GetMetricHandler(m MetricsStorer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}
		metric := m.Get(metricName)
		errorMessage := fmt.Sprintf("metric %s not found", metricName)
		if metric == nil {
			http.Error(w, errorMessage, http.StatusNotFound)
			return
		}
		if metricType == "counter" {
			if _, ok := metric.(storage.Counter); !ok {
				http.Error(w, "invalid metric type", http.StatusBadRequest)
				return
			}
		} else if metricType == "gauge" {
			if _, ok := metric.(storage.Gauge); !ok {
				http.Error(w, "invalid metric type", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
		metricStr := fmt.Sprintf("%v", metric)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metricStr))
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

func ShowMetricsSummaryHandler(m MetricsStorer) func(http.ResponseWriter, *http.Request) {
	tmpl := template.Must(template.New("metrics").Parse(metricsTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		data := MetricsData{}
		for metricName, metricValue := range m.GetAll() {
			data.Metrics = append(data.Metrics, Metric{Name: metricName, Value: metricValue})
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "error rendering template", http.StatusInternalServerError)
			return
		}
	}
}
