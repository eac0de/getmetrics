package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/pkg/compressor"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	conf   *config.AgentConfig
	client *resty.Client
	done   chan struct{}
}

func NewAgent(conf *config.AgentConfig) *Agent {
	conf.ServerURL = "http://" + conf.ServerURL
	client := resty.New()
	return &Agent{
		conf:   conf,
		client: client,
	}
}

func (a *Agent) Stop() {
	close(a.done)
}

func (a *Agent) Run() {
	var pollCount int64
	var metrics Metrics

	a.done = make(chan struct{})

	go func() {
		for {
			select {
			case <-a.done:
				log.Println("Poll goroutine is shutting down...")
				return
			default:
				metrics = a.collectMetrics(&pollCount)
				time.Sleep(a.conf.PollInterval)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-a.done:
				log.Println("Report goroutine is shutting down...")
				return
			default:
				a.sendMetrics(metrics)
				time.Sleep(a.conf.ReportInterval)
			}
		}
	}()

	log.Println("Agent is running. Press Ctrl+C to stop.")
	<-a.done // Блокируемся до закрытия канала done
}

type Metrics struct {
	Alloc         float64 `json:"alloc"`
	BuckHashSys   float64 `json:"buck_hash_sys"`
	Frees         float64 `json:"frees"`
	GCCPUFraction float64 `json:"gccpufraction"`
	GCSys         float64 `json:"gcsys"`
	HeapAlloc     float64 `json:"heap_alloc"`
	HeapIdle      float64 `json:"heap_idle"`
	HeapInuse     float64 `json:"heap_inuse"`
	HeapObjects   float64 `json:"heap_objects"`
	HeapReleased  float64 `json:"heap_released"`
	HeapSys       float64 `json:"heap_sys"`
	LastGC        float64 `json:"last_gc"`
	Lookups       float64 `json:"lookups"`
	MCacheInuse   float64 `json:"mcache_inuse"`
	MCacheSys     float64 `json:"mcache_sys"`
	MSpanInuse    float64 `json:"mspan_inuse"`
	MSpanSys      float64 `json:"mspan_sys"`
	Mallocs       float64 `json:"mallocs"`
	NextGC        float64 `json:"next_gc"`
	NumForcedGC   float64 `json:"num_forced_gc"`
	NumGC         float64 `json:"num_gc"`
	OtherSys      float64 `json:"other_sys"`
	PauseTotalNs  float64 `json:"pause_total_ns"`
	StackInuse    float64 `json:"stack_inuse"`
	StackSys      float64 `json:"stack_sys"`
	Sys           float64 `json:"sys"`
	TotalAlloc    float64 `json:"total_alloc"`
	PollCount     int64   `json:"poll_count"`
	RandomValue   float64 `json:"random_value"`
}

func (a *Agent) collectMetrics(pollCount *int64) Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	*pollCount++

	return Metrics{
		Alloc:         float64(memStats.Alloc),
		BuckHashSys:   float64(memStats.BuckHashSys),
		Frees:         float64(memStats.Frees),
		GCCPUFraction: float64(memStats.GCCPUFraction),
		GCSys:         float64(memStats.GCSys),
		HeapAlloc:     float64(memStats.HeapAlloc),
		HeapIdle:      float64(memStats.HeapIdle),
		HeapInuse:     float64(memStats.HeapInuse),
		HeapObjects:   float64(memStats.HeapObjects),
		HeapReleased:  float64(memStats.HeapReleased),
		HeapSys:       float64(memStats.HeapSys),
		LastGC:        float64(memStats.LastGC),
		Lookups:       float64(memStats.Lookups),
		MCacheInuse:   float64(memStats.MCacheInuse),
		MCacheSys:     float64(memStats.MCacheSys),
		MSpanInuse:    float64(memStats.MSpanInuse),
		MSpanSys:      float64(memStats.MSpanSys),
		Mallocs:       float64(memStats.MSpanSys),
		NextGC:        float64(memStats.MSpanSys),
		NumForcedGC:   float64(memStats.MSpanSys),
		NumGC:         float64(memStats.MSpanSys),
		OtherSys:      float64(memStats.MSpanSys),
		PauseTotalNs:  float64(memStats.MSpanSys),
		StackInuse:    float64(memStats.MSpanSys),
		StackSys:      float64(memStats.MSpanSys),
		Sys:           float64(memStats.MSpanSys),
		TotalAlloc:    float64(memStats.MSpanSys),
		PollCount:     *pollCount,
		RandomValue:   float64(memStats.MSpanSys),
	}
}

// request format /update/
func (a *Agent) sendMetric(metricName string, metricValue interface{}) error {
	url := fmt.Sprintf("%s/update/", a.conf.ServerURL)
	metric := models.Metrics{
		ID: metricName,
	}
	metricType := reflect.ValueOf(metricValue).Type().Name()
	switch metricType {
	case "int64":
		metric.MType = "counter"
		value, ok := metricValue.(int64)
		if !ok {
			return fmt.Errorf("invalid value for type counter")
		}
		metric.Delta = &value
	case "float64":
		metric.MType = "gauge"
		value, ok := metricValue.(float64)
		if !ok {
			return fmt.Errorf("invalid value for type gauge")
		}
		metric.Value = &value
	default:
		return fmt.Errorf("invalid type of value")
	}
	metricJSON, err := json.Marshal(metric)
	if err != nil {
		return err
	}
	metricGzip, err := compressor.GzipData(metricJSON)
	if err != nil {
		return err
	}
	resp, err := a.client.
		R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(metricGzip).
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error: %s", resp.Error())
	}
	return nil
}

func (a *Agent) sendMetrics(metrics Metrics) {
	values := map[string]interface{}{
		"Alloc":         metrics.Alloc,
		"BuckHashSys":   metrics.BuckHashSys,
		"Frees":         metrics.Frees,
		"GCCPUFraction": metrics.GCCPUFraction,
		"GCSys":         metrics.GCSys,
		"HeapAlloc":     metrics.HeapAlloc,
		"HeapIdle":      metrics.HeapIdle,
		"HeapInuse":     metrics.HeapInuse,
		"HeapObjects":   metrics.HeapObjects,
		"HeapReleased":  metrics.HeapReleased,
		"HeapSys":       metrics.HeapSys,
		"LastGC":        metrics.LastGC,
		"Lookups":       metrics.Lookups,
		"MCacheInuse":   metrics.MCacheInuse,
		"MCacheSys":     metrics.MCacheSys,
		"MSpanInuse":    metrics.MSpanInuse,
		"MSpanSys":      metrics.MSpanSys,
		"Mallocs":       metrics.Mallocs,
		"NextGC":        metrics.NextGC,
		"NumForcedGC":   metrics.NumForcedGC,
		"NumGC":         metrics.NumGC,
		"OtherSys":      metrics.OtherSys,
		"PauseTotalNs":  metrics.PauseTotalNs,
		"StackInuse":    metrics.StackInuse,
		"StackSys":      metrics.StackSys,
		"Sys":           metrics.Sys,
		"TotalAlloc":    metrics.TotalAlloc,
		"PollCount":     metrics.PollCount,
		"RandomValue":   metrics.RandomValue,
	}

	for metricName, metricValue := range values {
		if err := a.sendMetric(metricName, metricValue); err != nil {
			fmt.Printf("failed to send metric %s: %s\n", metricName, err.Error())
		}
	}
}
