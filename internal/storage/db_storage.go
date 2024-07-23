package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/eac0de/getmetrics/internal/database"
	"github.com/eac0de/getmetrics/internal/models"
)

type databaseStorage struct {
	sqlDB *sql.DB
}

func NewDatabaseStorage(ctx context.Context, DSN string) (*databaseStorage, error) {
	sqlDB, err := sql.Open("pgx", DSN)
	if err != nil {
		return nil, err
	}
	dbs := databaseStorage{
		sqlDB: sqlDB,
	}
	err = database.Ping(ctx, dbs.sqlDB)
	if err != nil {
		return nil, err
	}
	err = database.Migrate(ctx, dbs.sqlDB)
	if err != nil {
		return nil, err
	}
	return &dbs, nil
}

func (dbs *databaseStorage) Close() error {
	err := dbs.sqlDB.Close()
	if err != nil {
		return nil
	}
	log.Println("DatabaseStorage closed correctly")
	return nil
}

func (dbs *databaseStorage) Ping(ctx context.Context) error {
	err := database.Ping(ctx, dbs.sqlDB)
	if err != nil {
		return NewErrorWithHTTPStatus(err, http.StatusInternalServerError)
	}
	return nil
}

func (dbs *databaseStorage) Get(ctx context.Context, metricName string, metricType string) (*models.Metrics, error) {
	return database.SelectMetricFromDatabase(ctx, dbs.sqlDB, metricName, metricType)
}

func (dbs *databaseStorage) GetAll(ctx context.Context) ([]*models.Metrics, error) {
	return database.SelectAllMetricsFromDatabase(ctx, dbs.sqlDB)
}

func (dbs *databaseStorage) Save(ctx context.Context, metric models.Metrics) (*models.Metrics, error) {
	return dbs.SaveBySQLModel(ctx, dbs.sqlDB, metric)
}

func (dbs *databaseStorage) SaveMany(ctx context.Context, metricsList []models.Metrics) ([]*models.Metrics, error) {
	metricsList, err := dbs.MergeMetricsList(metricsList)
	if err != nil {
		return nil, err
	}
	tx, err := dbs.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, NewErrorWithHTTPStatus(err, http.StatusInternalServerError)
	}
	var errList []error
	var newMetricsList []*models.Metrics
	for _, metric := range metricsList {
		newMetric, err := dbs.SaveBySQLModel(ctx, tx, metric)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		newMetricsList = append(newMetricsList, newMetric)
	}
	err = errors.Join(errList...)
	if err != nil {
		tx.Rollback()
		return nil, NewErrorWithHTTPStatus(err, http.StatusBadRequest)
	}
	tx.Commit()
	return newMetricsList, nil
}

func (dbs *databaseStorage) SaveBySQLModel(ctx context.Context, sql database.SQLModel, metric models.Metrics) (*models.Metrics, error) {
	if metric.ID == "" {
		return nil, NewErrorWithHTTPStatus(fmt.Errorf("Metric name is required"), http.StatusNotFound)
	}
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return nil, NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s must have filled value", metric.ID, models.Gauge), http.StatusBadRequest)
		}
	case models.Counter:
		if metric.Delta == nil {
			return nil, NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s must have filled delta", metric.ID, models.Counter), http.StatusBadRequest)
		}
		existMetric, err := dbs.Get(ctx, metric.ID, metric.MType)
		oldDelta := int64(0)
		if err == nil {
			oldDelta = *existMetric.Delta
		}
		delta := *metric.Delta + oldDelta
		metric.Delta = &delta
	default:
		return nil, NewErrorWithHTTPStatus(fmt.Errorf("Invalid metric type for %s: %s", metric.ID, metric.MType), http.StatusBadRequest)
	}
	var delta, value interface{}
	if metric.Delta != nil {
		delta = *metric.Delta
	}
	if metric.Value != nil {
		value = *metric.Value
	}
	err := database.InsertOrUpdateMetricIntoDatabase(ctx, dbs.sqlDB, metric.ID, metric.MType, delta, value)
	if err != nil {
		return nil, NewErrorWithHTTPStatus(fmt.Errorf("Metric saving error: %s", err.Error()), http.StatusInternalServerError)
	}
	return &metric, nil
}

func (dbs *databaseStorage) MergeMetricsList(metricsList []models.Metrics) ([]models.Metrics, error) {
	metricsMap := models.MetricsMap{
		Gauge:   map[string]float64{},
		Counter: map[string]int64{},
	}
	var errList []error
	for _, metric := range metricsList {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				errList = append(errList, NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s must have filled value", metric.ID, models.Gauge), http.StatusBadRequest))
				continue
			}
			metricsMap.Gauge[metric.ID] = *metric.Value
		case models.Counter:
			if metric.Delta == nil {
				errList = append(errList, NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s must have filled delta", metric.ID, models.Counter), http.StatusBadRequest))
				continue
			}
			oldDelta, ok := metricsMap.Counter[metric.ID]
			if !ok {
				oldDelta = int64(0)
			}
			delta := *metric.Delta + oldDelta
			metricsMap.Counter[metric.ID] = delta
		default:
			errList = append(errList, NewErrorWithHTTPStatus(fmt.Errorf("Invalid metric type for %s: %s", metric.ID, models.Counter), http.StatusBadRequest))
		}
	}
	err := errors.Join(errList...)
	if err != nil {
		return nil, NewErrorWithHTTPStatus(err, http.StatusBadRequest)
	}
	var mergeMetricsList []models.Metrics
	for ID, value := range metricsMap.Gauge {
		mergeMetricsList = append(mergeMetricsList, models.Metrics{ID: ID, MType: models.Gauge, Value: &value})
	}
	for ID, value := range metricsMap.Counter {
		mergeMetricsList = append(mergeMetricsList, models.Metrics{ID: ID, MType: models.Counter, Delta: &value})
	}
	return mergeMetricsList, nil
}
