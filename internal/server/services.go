package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/eac0de/getmetrics/internal/storage"
)

func LoadMetricsFromFile(filename string, m *storage.MetricsStorage) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	metrics := map[string]map[string]interface{}{}
	err = json.Unmarshal(buf, &metrics)
	if err != nil {
		return err
	}
	m.SystemMetrics = metrics
	return nil
}

func SaveMetricsToFile(filename string, m *storage.MetricsStorage) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	metrics := m.GetAll()
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
