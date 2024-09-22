package fileservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eac0de/getmetrics/internal/storage/memstore"
)

type FileService struct {
	MemoryStorage *memstore.MemoryStore
	FilePath      string
}

func New(memoryStorage *memstore.MemoryStore, filePath string) (*FileService, error) {
	if filePath == "" {
		return nil, fmt.Errorf("filePath cannot be an empty string")
	}
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.Decode(&memoryStorage.MetricsData)

	return &FileService{
		MemoryStorage: memoryStorage,
		FilePath:      filePath,
	}, nil
}

func (fs *FileService) SaveMetrics() error {
	f, err := os.OpenFile(fs.FilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.MarshalIndent(fs.MemoryStorage.MetricsData, "", "    ")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	log.Println("Metric are saved to file")
	return nil
}

func (fs *FileService) StartSavingMetrics(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			log.Println("StartSavingMetrics goroutine is shutting down...")
			return
		case <-ticker.C:
			err := fs.SaveMetrics()
			if err != nil {
				log.Printf("Metric saving to file error: %s", err.Error())
			}
		}

	}
}
