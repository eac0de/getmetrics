package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/eac0de/getmetrics/internal/models"
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
		row := db.sqlDB.QueryRowContext(
			ctx, "SELECT delta FROM metrics WHERE m_type = $1 AND id = $2", um.MType, um.ID,
		)
		var oldValue sql.NullInt64
		row.Scan(&oldValue)
		if oldValue.Valid {
			metricValueInt += oldValue.Int64
		}
		metric.Delta = &metricValueInt
		metric.MType = Counter
	default:
		// Обработка некорректного типа метрики
		return nil, fmt.Errorf("invalid metric type")
	}
	metric.ID = um.ID
	var deltaValue, valueValue interface{}
	if metric.Delta != nil {
		deltaValue = *metric.Delta
	}
	if metric.Value != nil {
		valueValue = *metric.Value
	}
	row := db.sqlDB.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM metrics WHERE m_type = $1 AND id = $2",
		metric.MType, metric.ID,
	)
	var exist int64
	err = row.Scan(&exist)
	if err != nil {
		return nil, fmt.Errorf("metric scan error")
	}
	if exist > 0 {
		_, err = db.sqlDB.ExecContext(
			ctx,
			"UPDATE metrics SET delta=$1, value=$2 WHERE m_type =$3  AND id =$4",
			deltaValue, valueValue, metric.MType, metric.ID,
		)
	} else {
		_, err = db.sqlDB.ExecContext(
			ctx,
			"INSERT INTO metrics (id, m_type, delta, value) VALUES($1,$2,$3,$4)",
			metric.ID, metric.MType, deltaValue, valueValue,
		)
	}
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
	for metricName, metricValue := range ms.Counter {
		row := db.sqlDB.QueryRowContext(
			ctx,
			"SELECT delta FROM metrics WHERE m_type = $1 AND id = $2",
			Counter, metricName,
		)
		var oldValue sql.NullInt64
		row.Scan(&oldValue)
		if oldValue.Valid {
			metricValue += oldValue.Int64
		}
		err = db.insertOrUpdateMetrics(ctx, tx, metricName, Counter, metricValue, nil)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		metricsList = append(metricsList, &models.Metrics{ID: metricName, MType: Counter, Delta: &metricValue})
	}
	for metricName, metricValue := range ms.Gauge {
		err = db.insertOrUpdateMetrics(ctx, tx, metricName, Gauge, metricValue, nil)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		metricsList = append(metricsList, &models.Metrics{ID: metricName, MType: Gauge, Value: &metricValue})
	}
	tx.Commit()
	return metricsList, nil
}

type ExecSqlModel interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (db *DatabaseSQL) insertOrUpdateMetrics(
	ctx context.Context,
	sqlModel ExecSqlModel,
	ID string,
	MType string,
	Delta interface{},
	Value interface{},
) error {
	row := db.sqlDB.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM metrics WHERE m_type=$1 AND id=$2",
		MType, ID,
	)
	var exist int64
	err := row.Scan(&exist)
	if err != nil {
		return err
	}
	if exist > 0 {
		_, err = sqlModel.ExecContext(
			ctx,
			"UPDATE metrics SET delta=$1, value=$2 WHERE m_type=$3 AND id=$4",
			Delta, Value, MType, ID,
		)
	} else {
		_, err = sqlModel.ExecContext(
			ctx,
			"INSERT INTO metrics (id, m_type, delta, value) VALUES($1,$2,$3,$4)",
			ID, MType, Delta, Value,
		)
	}
	return nil
}

func (db *DatabaseSQL) Get(ctx context.Context, metricType string, metricName string) (*models.Metrics, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	var metric models.Metrics
	row := db.sqlDB.QueryRowContext(
		ctx,
		"SELECT id, m_type, delta, value FROM metrics WHERE m_type=$1 AND id=$2",
		metricType, metricName,
	)
	var delta sql.NullInt64
	var value sql.NullFloat64
	err := row.Scan(&metric.ID, &metric.MType, &delta, &value)
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
	var metrics []*models.Metrics
	rows, err := db.sqlDB.QueryContext(
		ctx,
		"SELECT id, m_type, delta, value FROM metrics",
	)
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
