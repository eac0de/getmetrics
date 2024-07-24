package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type (
	SQLModel interface {
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}
	SQLPinger interface {
		PingContext(ctx context.Context) error
	}
)

func Ping(
	ctx context.Context,
	pinger SQLPinger,
) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return pinger.PingContext(ctx)
}

func Migrate(ctx context.Context, sqlDB *sql.DB) error {
	migrationsDir := "./migrations"
	if err := goose.UpContext(ctx, sqlDB, migrationsDir); err != nil {
		return err
	}
	log.Println("Migrations applied successfully")
	return nil
}

func execWithRetry(
	ctx context.Context,
	model SQLModel,
	query string,
	args ...any,
) (sql.Result, error) {
	var err error
	var result sql.Result
	for waitTime := 1; waitTime <= 5; waitTime += 2 {
		result, err = model.ExecContext(ctx, query, args...)
		if err != nil {
			fmt.Println(err.Error())
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				fmt.Println(pgErr.Code)
				if pgerrcode.IsConnectionException(pgErr.Code) {
					fmt.Printf("Database connection error. New attempt in %v sec\n", waitTime)
					time.Sleep(time.Duration(waitTime) * time.Second)
					continue
				}
			}
		}
		break
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func queryWithRetry(
	ctx context.Context,
	model SQLModel,
	query string,
	args ...any,
) (*sql.Rows, error) {
	var result *sql.Rows
	var err error
	for waitTime := 1; waitTime <= 5; waitTime += 2 {
		result, err = model.QueryContext(ctx, query, args...)
		if err != nil {
			fmt.Println(err.Error())
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				fmt.Println(pgErr.Code)
				if pgerrcode.IsConnectionException(pgErr.Code) {
					fmt.Printf("Database connection error. New attempt in %v sec\n", waitTime)
					time.Sleep(time.Duration(waitTime) * time.Second)
					continue
				}
			}
		}
		break
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func InsertOrUpdateMetricIntoDatabase(
	ctx context.Context,
	model SQLModel,
	ID string,
	MType string,
	Delta interface{},
	Value interface{},
) error {
	query := "SELECT COUNT(*) FROM metrics WHERE m_type=$1 AND id=$2"
	rows, err := queryWithRetry(ctx, model, query, MType, ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()
	if rows.Err() != nil {
		return rows.Err()
	}
	var exist int
	err = rows.Scan(&exist)
	if err != nil {
		return err
	}
	if exist > 0 {
		query = "UPDATE metrics SET delta=$3, value=$4 WHERE m_type=$2 AND id=$1"
	} else {
		query = "INSERT INTO metrics (id, m_type, delta, value) VALUES($1,$2,$3,$4)"
	}
	_, err = execWithRetry(ctx, model, query, ID, MType, Delta, Value)
	if err != nil {
		return err
	}
	return nil
}

func SelectMetricFromDatabase(
	ctx context.Context,
	model SQLModel,
	metricName string,
	metricType string,
) (*models.Metrics, error) {
	query := "SELECT id, m_type, delta, value FROM metrics WHERE m_type=$1 AND id=$2"
	rows, err := queryWithRetry(ctx, model, query, metricType, metricName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	exists := rows.Next()
	if !exists {
		return nil, fmt.Errorf("metric %s with type %s not found", metricName, metricType)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	var metric models.Metrics
	var delta sql.NullInt64
	var value sql.NullFloat64
	err = rows.Scan(&metric.ID, &metric.MType, &delta, &value)
	if err != nil {
		return nil, err
	}
	if delta.Valid {
		metric.Delta = &delta.Int64
	}
	if value.Valid {
		metric.Value = &value.Float64
	}
	return &metric, nil
}

func SelectAllMetricsFromDatabase(
	ctx context.Context,
	model SQLModel,
) ([]*models.Metrics, error) {
	var metrics []*models.Metrics
	query := "SELECT id, m_type, delta, value FROM metrics"
	rows, err := queryWithRetry(ctx, model, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m models.Metrics
		var delta sql.NullInt64
		var value sql.NullFloat64
		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return nil, err
		}
		if delta.Valid {
			m.Delta = &delta.Int64
		}
		if value.Valid {
			m.Value = &value.Float64
		}
		metrics = append(metrics, &m)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
