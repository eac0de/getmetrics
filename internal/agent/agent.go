package agent

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	serverURL      string
	pollInterval   time.Duration
	reportInterval time.Duration
}

func NewAgent(serverURL string, pollInterval time.Duration, reportInterval time.Duration) *Agent {
	serverURL = "http://" + serverURL
	return &Agent{
		serverURL:      serverURL,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
	}
}

func (a *Agent) Run() {
	var pollCount storage.Counter
	var metrics Metrics

	go func() {
		for {
			metrics = a.collectMetrics(&pollCount)
			time.Sleep(a.pollInterval)

		}
	}()

	go func() {
		for {
			a.sendMetrics(metrics)
			time.Sleep(a.reportInterval)
		}
	}()

	select {}
}

type Metrics struct {
	Alloc         storage.Gauge   `json:"alloc"`
	BuckHashSys   storage.Gauge   `json:"buck_hash_sys"`
	Frees         storage.Gauge   `json:"frees"`
	GCCPUFraction storage.Gauge   `json:"gccpufraction"`
	GCSys         storage.Gauge   `json:"gcsys"`
	HeapAlloc     storage.Gauge   `json:"heap_alloc"`
	HeapIdle      storage.Gauge   `json:"heap_idle"`
	HeapInuse     storage.Gauge   `json:"heap_inuse"`
	HeapObjects   storage.Gauge   `json:"heap_objects"`
	HeapReleased  storage.Gauge   `json:"heap_released"`
	HeapSys       storage.Gauge   `json:"heap_sys"`
	LastGC        storage.Gauge   `json:"last_gc"`
	Lookups       storage.Gauge   `json:"lookups"`
	MCacheInuse   storage.Gauge   `json:"mcache_inuse"`
	MCacheSys     storage.Gauge   `json:"mcache_sys"`
	MSpanInuse    storage.Gauge   `json:"mspan_inuse"`
	MSpanSys      storage.Gauge   `json:"mspan_sys"`
	Mallocs       storage.Gauge   `json:"mallocs"`
	NextGC        storage.Gauge   `json:"next_gc"`
	NumForcedGC   storage.Gauge   `json:"num_forced_gc"`
	NumGC         storage.Gauge   `json:"num_gc"`
	OtherSys      storage.Gauge   `json:"other_sys"`
	PauseTotalNs  storage.Gauge   `json:"pause_total_ns"`
	StackInuse    storage.Gauge   `json:"stack_inuse"`
	StackSys      storage.Gauge   `json:"stack_sys"`
	Sys           storage.Gauge   `json:"sys"`
	TotalAlloc    storage.Gauge   `json:"total_alloc"`
	PollCount     storage.Counter `json:"poll_count"`
	RandomValue   storage.Gauge   `json:"random_value"`
}

func (a *Agent) collectMetrics(pollCount *storage.Counter) Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	*pollCount++

	return Metrics{
		Alloc:         storage.Gauge(memStats.Alloc),
		BuckHashSys:   storage.Gauge(memStats.BuckHashSys),
		Frees:         storage.Gauge(memStats.Frees),
		GCCPUFraction: storage.Gauge(memStats.GCCPUFraction),
		GCSys:         storage.Gauge(memStats.GCSys),
		HeapAlloc:     storage.Gauge(memStats.HeapAlloc),
		HeapIdle:      storage.Gauge(memStats.HeapIdle),
		HeapInuse:     storage.Gauge(memStats.HeapInuse),
		HeapObjects:   storage.Gauge(memStats.HeapObjects),
		HeapReleased:  storage.Gauge(memStats.HeapReleased),
		HeapSys:       storage.Gauge(memStats.HeapSys),
		LastGC:        storage.Gauge(memStats.LastGC),
		Lookups:       storage.Gauge(memStats.Lookups),
		MCacheInuse:   storage.Gauge(memStats.MCacheInuse),
		MCacheSys:     storage.Gauge(memStats.MCacheSys),
		MSpanInuse:    storage.Gauge(memStats.MSpanInuse),
		MSpanSys:      storage.Gauge(memStats.MSpanSys),
		Mallocs:       storage.Gauge(memStats.MSpanSys),
		NextGC:        storage.Gauge(memStats.MSpanSys),
		NumForcedGC:   storage.Gauge(memStats.MSpanSys),
		NumGC:         storage.Gauge(memStats.MSpanSys),
		OtherSys:      storage.Gauge(memStats.MSpanSys),
		PauseTotalNs:  storage.Gauge(memStats.MSpanSys),
		StackInuse:    storage.Gauge(memStats.MSpanSys),
		StackSys:      storage.Gauge(memStats.MSpanSys),
		Sys:           storage.Gauge(memStats.MSpanSys),
		TotalAlloc:    storage.Gauge(memStats.MSpanSys),
		PollCount:     *pollCount,
		RandomValue:   storage.Gauge(memStats.MSpanSys),
	}
}

func (a *Agent) sendMetric(metricType string, metricName string, metricValue interface{}) error {
	url := fmt.Sprintf("%s/update/%s/%s/%v", a.serverURL, metricType, metricName, metricValue)
	client := resty.New()
	resp, err := client.R().SetHeader("contentType", "text/plain").Post(url)
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
		metricType := strings.ToLower(reflect.TypeOf(metricValue).Name())
		if err := a.sendMetric(metricType, metricName, metricValue); err != nil {
			fmt.Printf("failed to send metric %s: %s\n", metricName, err.Error())
		}
	}
}
