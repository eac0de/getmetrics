package handlers

type MetricsRepository interface {
	Save(metricName string, metricValue interface{})
	Get(metricName string) interface{}
	GetAll() map[string]interface{}
}
