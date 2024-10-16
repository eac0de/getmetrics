// Package handlers предоставляет HTTP-обработчики для управления метриками.
//
// Этот пакет реализует функциональность для обновления, получения и отображения метрик.
// Обработчики взаимодействуют с хранилищем метрик и поддерживают как текстовый, так и JSON форматы.
// Пакет обеспечивает валидацию метрик и обработку ошибок, а также предоставляет возможность
// отображения метрик в виде HTML-страницы.
//
// Основные функции пакета включают:
// - Обновление метрик через параметры URL и JSON.
// - Получение метрик в текстовом и JSON форматах.
// - Отображение всех метрик на HTML-странице.
package handlers

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
	"github.com/go-chi/chi/v5"
)

// IMetricsStore интерфейс для хранения метрик. Определяет методы для сохранения и получения метрик.
type IMetricsStore interface {
	SaveMetric(ctx context.Context, metric models.Metric) error
	SaveMetrics(ctx context.Context, metricsList []models.Metric) error
	GetMetric(ctx context.Context, metricName string, metricType string) (*models.Metric, error)
	ListAllMetrics(ctx context.Context) ([]*models.Metric, error)
}

// MetricsHandlers представляет набор обработчиков для работы с метриками.
type MetricsHandlers struct {
	MetricsStore IMetricsStore // Хранилище метрик
	SecretKey    string        // Секретный ключ для генерации подписи
}

// NewMetricsHandlers создает новый экземпляр MetricsHandlers.
//
// Принимает на вход интерфейс хранилища метрик и секретный ключ.
func NewMetricsHandlers(metricStore IMetricsStore, secretKey string) *MetricsHandlers {
	return &MetricsHandlers{
		MetricsStore: metricStore,
		SecretKey:    secretKey,
	}
}

// UpdateMetricHandler возвращает HTTP-обработчик для обновления метрики по URL параметрам.
//
// Поддерживаются два типа метрик: counter и gauge. Обновляет значение метрики в хранилище.
func (h *MetricsHandlers) UpdateMetricHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		if metricName == "" {
			http.Error(w, "metric name is required", http.StatusNotFound)
			return
		}
		metric := models.Metric{
			ID:    metricName,
			MType: metricType,
		}
		switch metric.MType {
		case models.Counter:
			delta, err := strconv.ParseInt(metricValue, 10, 64)
			if err == nil {
				metric.Delta = &delta
			}
		case models.Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err == nil {
				metric.Value = &value
			}
		}
		err := h.validateMetric(metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if metric.MType == models.Counter {
			var oldDelta int64
			oldDelta, err = h.getOldDelta(r.Context(), metric.ID)
			if err != nil {
				msg, statusCode := errors.GetMessageAndStatusCode(err)
				http.Error(w, msg, statusCode)
				return
			}
			*metric.Delta += oldDelta
		}
		err = h.MetricsStore.SaveMetric(r.Context(), metric)
		if err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		var answer string
		switch metric.MType {
		case models.Counter:
			answer = fmt.Sprintf("%v", *metric.Delta)
		case models.Gauge:
			answer = fmt.Sprintf("%v", *metric.Value)
		}
		data := []byte(answer)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		h.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// UpdateMetricJSONHandler возвращает HTTP-обработчик для обновления метрики через JSON-пост запрос.
//
// Обрабатывает JSON-запрос с метрикой и обновляет ее значение в хранилище.
func (h *MetricsHandlers) UpdateMetricJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metric
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		err := h.validateMetric(metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if metric.MType == models.Counter {
			var oldDelta int64
			oldDelta, err = h.getOldDelta(r.Context(), metric.ID)
			if err != nil {
				msg, statusCode := errors.GetMessageAndStatusCode(err)
				http.Error(w, msg, statusCode)
				return
			}
			*metric.Delta += oldDelta
		}
		err = h.MetricsStore.SaveMetric(r.Context(), metric)
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
		h.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// UpdateMetricsJSONHandler возвращает HTTP-обработчик для массового обновления метрик через JSON.
//
// Обрабатывает список метрик, выполняет их валидацию и обновляет в хранилище.
func (h *MetricsHandlers) UpdateMetricsJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metricsList []models.Metric
		if err := json.NewDecoder(r.Body).Decode(&metricsList); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		var errsList []error
		for _, metric := range metricsList {
			err := h.validateMetric(metric)
			if err != nil {
				errsList = append(errsList, err)
			}
		}
		if len(errsList) > 0 {
			err := stderr.Join(errsList...)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		metricsList, err := h.mergeMetricsList(r.Context(), metricsList)
		if err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		err = h.MetricsStore.SaveMetrics(r.Context(), metricsList)
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
		h.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// GetMetricHandler возвращает HTTP-обработчик для получения метрики по имени и типу.
//
// Обрабатывает запрос для получения метрики и возвращает её значение в формате текста.
func (h *MetricsHandlers) GetMetricHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")
		metric, err := h.MetricsStore.GetMetric(r.Context(), metricName, metricType)
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
		h.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// GetMetricJSONHandler возвращает HTTP-обработчик для получения метрики через JSON.
//
// Получает метрику из хранилища и возвращает ее в формате JSON.
func (h *MetricsHandlers) GetMetricJSONHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metric
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		metric, err := h.MetricsStore.GetMetric(r.Context(), m.ID, m.MType)
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
		h.addSign(w, data)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// ShowMetricsSummaryHandler возвращает HTTP-обработчик для отображения HTML-страницы со списком всех метрик.
//
// Загружает шаблон и отображает страницу со всеми метриками из хранилища.
func (h *MetricsHandlers) ShowMetricsSummaryHandler() func(http.ResponseWriter, *http.Request) {
	filePath := filepath.Join("templates", "metrics_summary.html")
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("open template file error: %s", err.Error())
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("read template error: %s", err.Error())
	}
	tmpl, err := template.New("metrics").Parse(string(data))
	if err != nil {
		log.Fatalf("parse template error: %s", err.Error())
	}
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := h.MetricsStore.ListAllMetrics(r.Context())
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

func (h *MetricsHandlers) getOldDelta(ctx context.Context, ID string) (int64, error) {
	var oldDelta int64
	metric, err := h.MetricsStore.GetMetric(ctx, ID, models.Counter)
	if err != nil {
		_, statusCode := errors.GetMessageAndStatusCode(err)
		if statusCode != http.StatusNotFound {
			return 0, err
		}
	} else {
		oldDelta = *metric.Delta
	}
	return oldDelta, nil
}

func (h *MetricsHandlers) addSign(w http.ResponseWriter, data []byte) {
	hash := hasher.HashSumToString(data, h.SecretKey)
	if hash != "" {
		w.Header().Set("HashSHA256", hash)
	}
}

func (h *MetricsHandlers) validateMetric(metric models.Metric) error {
	if metric.ID == "" {
		return fmt.Errorf("metric name is required")
	}
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("metric %s with type %s must have filled value", metric.ID, metric.MType)
		}
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("metric %s with type %s must have filled delta", metric.ID, metric.MType)
		}
	default:
		return fmt.Errorf("invalid metric type for %s: %s", metric.ID, metric.MType)
	}
	return nil
}

func (h *MetricsHandlers) mergeMetricsList(ctx context.Context, metricsList []models.Metric) ([]models.Metric, error) {
	metricsMap := models.MetricsData{
		Gauge:   map[string]float64{},
		Counter: map[string]int64{},
	}

	// Обработка метрик
	for _, metric := range metricsList {
		switch metric.MType {
		case models.Gauge:
			metricsMap.Gauge[metric.ID] = *metric.Value
		case models.Counter:
			metricsMap.Counter[metric.ID] += *metric.Delta
		}
	}

	// Формируем результирующий список
	mergeMetricsList := make([]models.Metric, 0, len(metricsMap.Gauge)+len(metricsMap.Counter))

	// Добавляем все gauge метрики
	for ID, value := range metricsMap.Gauge {
		mergeMetricsList = append(mergeMetricsList, models.Metric{ID: ID, MType: models.Gauge, Value: &value})
	}

	// Добавляем все counter метрики
	for ID, value := range metricsMap.Counter {
		oldDelta, err := h.getOldDelta(ctx, ID)
		if err != nil {
			return nil, err
		}
		value += oldDelta
		mergeMetricsList = append(mergeMetricsList, models.Metric{ID: ID, MType: models.Counter, Delta: &value})
	}

	return mergeMetricsList, nil
}
