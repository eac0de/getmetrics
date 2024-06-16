package handlers

type MetricsStorer interface {
	Save(metricName string, metricValue interface{})
	Get(metricName string) interface{}
	GetAll() map[string]interface{}
}
