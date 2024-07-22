package storage

import (
	"context"

	"github.com/eac0de/getmetrics/internal/models"
)

type MetricsStorer interface {
	Save(ctx context.Context, metric *models.Metrics) error
	SaveMany(ctx context.Context, metricsList []*models.Metrics) error
	Get(ctx context.Context, metricName string, metricType string) (*models.Metrics, error)
	GetAll(ctx context.Context) ([]*models.Metrics, error)
	Close() error
	Ping(ctx context.Context) error
}
