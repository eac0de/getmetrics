package storage

import (
	"context"

	"github.com/eac0de/getmetrics/internal/models"
)

type MetricsStorer interface {
	Save(ctx context.Context, um *models.UnknownMetrics) (*models.Metrics, error)
	SaveMany(ctx context.Context, umList []*models.UnknownMetrics) ([]*models.Metrics, error)
	Get(ctx context.Context, metricType string, metricName string) (*models.Metrics, error)
	GetAll(ctx context.Context) ([]*models.Metrics, error)
}
