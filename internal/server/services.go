package server

import (
	"encoding/json"
	"io"
	"os"

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
