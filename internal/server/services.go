package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/internal/storage"
)

func LoadMetricsFromFile(filename string, m *storage.MetricsStorage) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0666)
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

	// Декодируем JSON
	decoder := json.NewDecoder(f)
	metrics := new(models.SystemMetrics)
	if err := decoder.Decode(&metrics); err != nil {
		return err
	}

	// Сохраняем метрики в хранилище
	m.SystemMetrics = *metrics
	return nil
}

func SaveMetricsToFile(filename string, m *storage.MetricsStorage) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	metrics := m.SystemMetrics
	data, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func SaveMetricsToFileGorutine(s *MetricsServer, m *storage.MetricsStorage) {
	for {
		select {
		case <-s.exit:
			log.Println("SaveMetricsToFile goroutine is shutting down...")
			return
		default:
			time.Sleep(s.storeInterval)
			err := SaveMetricsToFile(s.fileStoragePath, m)
			if err != nil {
				fmt.Printf("metrics saving error: %s", err.Error())
			}
			fmt.Println("metrics have been preserved")
		}

	}

}
