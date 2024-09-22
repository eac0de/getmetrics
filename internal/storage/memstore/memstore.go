package memstore

import (
	"sync"

	"github.com/eac0de/getmetrics/internal/models"
)

type MemoryStore struct {
	mu          sync.Mutex
	MetricsData models.MetricsData
}

func New() *MemoryStore {
	store := MemoryStore{
		MetricsData: models.MetricsData{
			Counter: make(map[string]int64),
			Gauge:   make(map[string]float64),
		},
	}
	return &store
}
