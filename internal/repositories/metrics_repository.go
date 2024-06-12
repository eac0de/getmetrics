package repositories

type MetricsRepository interface {
	Save(metricName string, metricValue interface{})
	Get(metricName string) interface{}
}
