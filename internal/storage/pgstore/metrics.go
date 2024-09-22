package pgstore

import (
	"context"
	"database/sql"
	stderr "errors"
	"net/http"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/pkg/errors"
)

func (store *PostgresqlStore) SaveMetric(ctx context.Context, metric models.Metric) error {
	if metric.MType == models.Counter {
		var delta int64
		query := "SELECT delta FROM metrics WHERE type=$1 AND id=$2"
		err := store.GetContext(ctx, &delta, query, metric.MType, metric.ID)
		if err != nil {
			if !stderr.Is(err, sql.ErrNoRows) {
				return err
			}
		}
		*metric.Delta = delta + *metric.Delta
	}
	query := `
	INSERT INTO metrics (id, type, delta, value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id, type)
	DO UPDATE SET delta = $3, value = $4
	`
	_, err := store.ExecContext(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value)
	if err != nil {
		return err
	}
	return nil
}

func (store *PostgresqlStore) SaveMetrics(ctx context.Context, metricsList []models.Metric) error {
	tx, err := store.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	var errsList []error
	query := `
	INSERT INTO metrics (id, type, delta, value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id, type)
	DO UPDATE SET delta = $3, value = $4
	`
	for _, metric := range metricsList {
		if metric.MType == models.Counter {
			var delta int64
			query := "SELECT delta FROM metrics WHERE type=$1 AND id=$2"
			err := store.GetContext(ctx, &delta, query, metric.MType, metric.ID)
			if err != nil {
				if !stderr.Is(err, sql.ErrNoRows) {
					errsList = append(errsList, err)
					continue
				}
			}
			*metric.Delta = delta + *metric.Delta
		}
		_, err = tx.ExecContext(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			errsList = append(errsList, err)
		}
	}
	if len(errsList) > 0 {
		err = stderr.Join(errsList...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (store *PostgresqlStore) GetMetric(ctx context.Context, metricName string, metricType string) (*models.Metric, error) {
	query := "SELECT id, type, delta, value FROM metrics WHERE type=$1 AND id=$2"
	var metric *models.Metric
	err := store.GetContext(ctx, metric, query)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return nil, errors.NewErrorWithHTTPStatus(
				err,
				"Metric not found",
				http.StatusNotFound,
			)
		}
		return nil, err
	}
	return metric, nil
}

func (store *PostgresqlStore) ListAllMetrics(ctx context.Context) ([]*models.Metric, error) {
	var metricsList []*models.Metric
	query := "SELECT id, type, delta, value FROM metrics"
	err := store.SelectContext(ctx, &metricsList, query)
	if err != nil {
		return nil, err
	}
	return metricsList, nil
}
