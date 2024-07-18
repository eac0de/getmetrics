package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type DatabaseSQL struct {
	sqlDB *sql.DB
	mu    sync.Mutex
}

func NewDatabaseSQL(ctx context.Context, databaseDSN string) (*DatabaseSQL, error) {
	sqlDB, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	db := &DatabaseSQL{
		sqlDB: sqlDB,
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	err = db.Migrate(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DatabaseSQL) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return db.sqlDB.PingContext(ctx)
}

func (db *DatabaseSQL) Close() error {
	return db.sqlDB.Close()
}

func (db *DatabaseSQL) Save(ctx context.Context, um *models.UnknownMetrics) (*models.Metrics, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	var err error
	var metric models.Metrics
	metric.ID = um.ID
	switch um.MType {
	case Gauge:
		metricValueFloat, ok := um.DeltaValue.(float64)
		if !ok {
			valueStr, ok := um.DeltaValue.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value type for guage metric(1)")
			}
			metricValueFloat, err = strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value type for guage metric(2)")
			}
		}
		metric.Value = &metricValueFloat
		metric.MType = Gauge
	case Counter:
		metricValueInt, ok := um.DeltaValue.(int64)
		if !ok {
			valueStr, ok := um.DeltaValue.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value type for counter metric(1)")
			}
			metricValueInt, err = strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value type for counter metric(2)")
			}
		}
		query := "SELECT delta FROM metrics WHERE m_type = $1 AND id = $2"
		rows, err := db.queryWithRetry(
			ctx, db.sqlDB, query, um.MType, um.ID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		rows.Next()
		var oldValue sql.NullInt64
		rows.Scan(&oldValue)
		if oldValue.Valid {
			metricValueInt += oldValue.Int64
		}
		metric.Delta = &metricValueInt
		metric.MType = Counter
	default:
		// Обработка некорректного типа метрики
		return nil, fmt.Errorf("invalid metric type")
	}
	var delta, value interface{}
	if metric.Delta != nil {
		delta = *metric.Delta
	}
	if metric.Value != nil {
		delta = *metric.Value
	}
	err = db.insertOrUpdateMetric(ctx, db.sqlDB, metric.ID, metric.MType, delta, value)
	if err != nil {
		return nil, fmt.Errorf("metric saving error")
	}
	return &metric, nil
}

func (db *DatabaseSQL) SaveMany(ctx context.Context, umList []*models.UnknownMetrics) ([]*models.Metrics, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	ms := models.SystemMetrics{Gauge: map[string]float64{}, Counter: map[string]int64{}}
	for _, um := range umList {
		var err error
		switch um.MType {
		case Gauge:
			metricValueFloat, ok := um.DeltaValue.(float64)
			if !ok {
				valueStr, ok := um.DeltaValue.(string)
				if !ok {
					return nil, fmt.Errorf("invalid value type for guage metric(1)")
				}
				metricValueFloat, err = strconv.ParseFloat(valueStr, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid value type for guage metric(2)")
				}
			}
			ms.Gauge[um.ID] = metricValueFloat
		case Counter:
			metricValueInt, ok := um.DeltaValue.(int64)
			if !ok {
				valueStr, ok := um.DeltaValue.(string)
				if !ok {
					return nil, fmt.Errorf("invalid value type for counter metric(1)")
				}
				metricValueInt, err = strconv.ParseInt(valueStr, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid value type for counter metric(2)")
				}
			}
			oldValue, ok := ms.Counter[um.ID]
			if !ok {
				oldValue = 0
			}
			ms.Counter[um.ID] = metricValueInt + oldValue
		default:
			// Обработка некорректного типа метрики
			return nil, fmt.Errorf("invalid metric type")
		}
	}
	tx, err := db.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	metricsList := []*models.Metrics{}
	var errList []error
	for metricName, metricValue := range ms.Counter {
		query := "SELECT delta FROM metrics WHERE m_type = $1 AND id = $2"
		rows, err := db.queryWithRetry(
			ctx, db.sqlDB, query, Counter, metricName,
		)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		defer rows.Close()
		rows.Next()
		var oldValue sql.NullInt64
		rows.Scan(&oldValue)
		if oldValue.Valid {
			metricValue += oldValue.Int64
		}
		err = db.insertOrUpdateMetric(ctx, tx, metricName, Counter, metricValue, nil)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		metricsList = append(metricsList, &models.Metrics{ID: metricName, MType: Counter, Delta: &metricValue})
	}
	for metricName, metricValue := range ms.Gauge {
		err = db.insertOrUpdateMetric(ctx, tx, metricName, Gauge, nil, metricValue)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		metricsList = append(metricsList, &models.Metrics{ID: metricName, MType: Gauge, Value: &metricValue})
	}
	err = errors.Join(errList...)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return metricsList, nil
}

type ExecSQLModel interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type QuerySQLModel interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func (db *DatabaseSQL) execWithRetry(ctx context.Context, model ExecSQLModel, query string, args ...any) (sql.Result, error) {
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

func (db *DatabaseSQL) queryWithRetry(ctx context.Context, model QuerySQLModel, query string, args ...any) (*sql.Rows, error) {
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

func (db *DatabaseSQL) insertOrUpdateMetric(
	ctx context.Context,
	sqlModel ExecSQLModel,
	ID string,
	MType string,
	Delta interface{},
	Value interface{},
) error {
	query := "SELECT COUNT(*) FROM metrics WHERE m_type=$1 AND id=$2"
	rows, err := db.queryWithRetry(ctx, db.sqlDB, query, MType, ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()
	var exist int64
	err = rows.Scan(&exist)
	if err != nil {
		return err
	}
	if exist > 0 {
		query = "UPDATE metrics SET delta=$3, value=$4 WHERE m_type=$2 AND id=$1"
	} else {
		query = "INSERT INTO metrics (id, m_type, delta, value) VALUES($1,$2,$3,$4)"
	}
	_, err = db.execWithRetry(ctx, sqlModel, query, ID, MType, Delta, Value)
	if err != nil {
		return err
	}
	return nil
}

func (db *DatabaseSQL) Get(ctx context.Context, metricType string, metricName string) (*models.Metrics, error) {
	query := "SELECT id, m_type, delta, value FROM metrics WHERE m_type=$1 AND id=$2"
	rows, err := db.queryWithRetry(ctx, db.sqlDB, query, metricType, metricName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	exists := rows.Next()
	if !exists {
		return nil, fmt.Errorf("metric %s with type %s not found", metricName, metricType)
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

func (db *DatabaseSQL) GetAll(ctx context.Context) ([]*models.Metrics, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	var metrics []*models.Metrics
	query := "SELECT id, m_type, delta, value FROM metrics"
	rows, err := db.queryWithRetry(ctx, db.sqlDB, query)
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

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (db *DatabaseSQL) Migrate(ctx context.Context) error {
	migrationsDir := "./migrations"
	// Применение миграций
	if err := goose.UpContext(ctx, db.sqlDB, migrationsDir); err != nil {
		return err
	}
	log.Println("Migrations applied successfully!")
	return nil
}
