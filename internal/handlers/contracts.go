package handlers

type MetricsStorer interface {
	Save(metricType string, metricName string, metricValue interface{}) error
	Get(metricType string, metricName string) interface{}
	GetAll() map[string]map[string]interface{}
}
