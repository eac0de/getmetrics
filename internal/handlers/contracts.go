package handlers

import "github.com/eac0de/getmetrics/internal/models"

type MetricsStorer interface {
	Save(metricType string, metricName string, metricValue interface{}) (*models.Metrics, error)
	Get(metricType string, metricName string) *models.Metrics
	GetAll() map[string]map[string]interface{}
}
