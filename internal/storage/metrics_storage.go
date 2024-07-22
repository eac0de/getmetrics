package storage

import (
	"context"
	"sync"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/models"
)

type MetricsStorage struct {
	mu      *sync.Mutex
	storage MetricsStorer
}

func NewMetricsStorage(ctx context.Context, config config.HTTPServerConfig) *MetricsStorage {
	var storage MetricsStorer
	storage, err := NewDatabaseStorage(ctx, config.DatabaseDSN)
	if err != nil {
		storage, err = NewFileStorage(ctx, config.FileStoragePath, config.StoreInterval)
		if err != nil {
			storage = NewMemoryStorage()
		}
	}
	metric := MetricsStorage{
		storage: storage,
	}
	return &metric
}

func (ms *MetricsStorage) Save(ctx context.Context, metric *models.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.storage.Save(ctx, metric)
}

func (ms *MetricsStorage) SaveMany(ctx context.Context, metricsList []*models.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.storage.SaveMany(ctx, metricsList)
}

func (ms *MetricsStorage) Get(ctx context.Context, metricName string, metricType string) (*models.Metrics, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.storage.Get(ctx, metricName, metricType)
}

func (ms *MetricsStorage) GetAll(ctx context.Context) ([]*models.Metrics, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.storage.GetAll(ctx)
}

func (ms *MetricsStorage) Close() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.storage.Close()
}

func (ms *MetricsStorage) Ping(ctx context.Context) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.storage.Ping(ctx)
}
