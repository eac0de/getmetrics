package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

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
		return NewErrorWithHTTPStatus(err, 500)
	}
	return nil
}

func (dbs *databaseStorage) Get(ctx context.Context, metricName string, metricType string) (*models.Metrics, error) {
	return database.SelectMetricFromDatabase(ctx, dbs.sqlDB, metricName, metricType)
}

func (dbs *databaseStorage) GetAll(ctx context.Context) ([]*models.Metrics, error) {
	return database.SelectAllMetricsFromDatabase(ctx, dbs.sqlDB)
}

func (dbs *databaseStorage) Save(ctx context.Context, metric *models.Metrics) error {
	return dbs.SaveBySQLModel(ctx, dbs.sqlDB, metric)
}

func (dbs *databaseStorage) SaveMany(ctx context.Context, metricsList []*models.Metrics) error {

	tx, err := dbs.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return NewErrorWithHTTPStatus(err, 500)
	}
	var errList []error
	for _, metric := range metricsList {
		err = dbs.SaveBySQLModel(ctx, tx, metric)
		if err != nil {
			errList = append(errList, err)
		}
	}
	err = errors.Join(errList...)
	if err != nil {
		tx.Rollback()
		return NewErrorWithHTTPStatus(err, 400)
	}
	tx.Commit()
	return nil
}

func (dbs *databaseStorage) SaveBySQLModel(ctx context.Context, sql database.SQLModel, metric *models.Metrics) error {
	if metric.ID == "" {
		NewErrorWithHTTPStatus(fmt.Errorf("Metric name is required"), 404)
	}
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s must have filled value", metric.ID, models.Gauge), 400)
		}
	case models.Counter:
		if metric.Delta == nil {
			return NewErrorWithHTTPStatus(fmt.Errorf("Metric %s with type %s must have filled delta", metric.ID, models.Counter), 400)
		}
		existMetric, err := dbs.Get(ctx, metric.ID, metric.MType)
		oldDelta := int64(0)
		if err == nil {
			oldDelta = *existMetric.Delta
		}
		*metric.Delta += oldDelta
	default:
		return NewErrorWithHTTPStatus(fmt.Errorf("Invalid metric type for %s: %s", metric.ID, models.Counter), 400)
	}
	err := database.InsertOrUpdateMetricIntoDatabase(ctx, dbs.sqlDB, metric.ID, metric.MType, metric.Delta, metric.Value)
	if err != nil {
		return NewErrorWithHTTPStatus(fmt.Errorf("Metric saving error: %s", err.Error()), 500)
	}
	return nil
}
