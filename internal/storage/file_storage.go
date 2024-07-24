package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

const EmptyFilename = ""

type fileStorage struct {
	*memoryStorage
	Filename     string
	SaveInterval time.Duration
}

func NewFileStorage(ctx context.Context, filename string, saveInterval time.Duration) (*fileStorage, error) {
	if filename == EmptyFilename {
		return nil, fmt.Errorf("filename cannot be an empty string")
	}
	_, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	memoryStorage := NewMemoryStorage()
	fs := fileStorage{
		memoryStorage: memoryStorage,
		Filename:      filename,
		SaveInterval:  saveInterval,
	}
	err = fs.LoadMetrics()
	if err != nil {
		log.Printf("Load metrics error: %s", err.Error())
	}
	go fs.StartSavingMetrics(ctx)
	return &fs, nil
}

func (fs *fileStorage) LoadMetrics() error {
	f, err := os.OpenFile(fs.Filename, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// Проверяем, что файл не пустой
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		// Если файл пустой, возвращаем nil, так как это не ошибка
		return nil
	}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&fs.MetricsMap); err != nil {
		return err
	}
	return nil
}

func (fs *fileStorage) SaveMetrics() error {
	f, err := os.OpenFile(fs.Filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.MarshalIndent(fs.MetricsMap, "", "    ")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	log.Println("Metrics are saved to file")
	return nil
}

func (fs *fileStorage) StartSavingMetrics(ctx context.Context) {
	ticker := time.NewTicker(fs.SaveInterval)
	for {
		select {
		case <-ctx.Done():
			log.Println("SaveMetricsToFile goroutine is shutting down...")
			return
		case <-ticker.C:
			err := fs.SaveMetrics()
			if err != nil {
				log.Printf("Metrics saving to file error: %s", err.Error())
			}
		}

	}

}

func (fs *fileStorage) Close() error {
	err := fs.SaveMetrics()
	if err != nil {
		return err
	}
	log.Println("FileStorage closed correctly")
	return nil
}
