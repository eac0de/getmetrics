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

type MetricsFileService struct {
	filename    string
	metricStore storage.MetricsStorer
}

func (mfs *MetricsFileService) LoadMetrics() error {
	f, err := os.OpenFile(mfs.filename, os.O_CREATE|os.O_RDONLY, 0666)
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
	metrics := []models.Metrics{}
	if err := decoder.Decode(&metrics); err != nil {
		return err
	}

	// Сохраняем метрики в хранилище
	for _, metric := range metrics {
		var value interface{}
		switch metric.MType {
		case storage.Gauge:
			value = *metric.Value
		case storage.Counter:
			value = *metric.Delta
		}
		_, err := mfs.metricStore.Save(metric.MType, metric.ID, value)
		if err != nil {
			fmt.Printf("save metric error: %s", err.Error())
		}
	}
	return nil
}

func (mfs *MetricsFileService) SaveMetrics() error {
	f, err := os.OpenFile(mfs.filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	metrics := mfs.metricStore.GetAll()
	if metrics == nil {
		fmt.Println("empty metrics list")
		return nil
	}
	data, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	fmt.Println("metrics have been preserved")
	return nil
}

func (mfs *MetricsFileService) SaveMetricsToFileGorutine(s *MetricsServer) {
	if s.conf.FileStoragePath == "" {
		return
	}
	for {
		select {
		case <-s.exit:
			log.Println("SaveMetricsToFile goroutine is shutting down...")
			return
		default:
			time.Sleep(s.conf.StoreInterval)
			err := mfs.SaveMetrics()
			if err != nil {
				fmt.Printf("metrics saving error: %s", err.Error())
			}
		}

	}

}
