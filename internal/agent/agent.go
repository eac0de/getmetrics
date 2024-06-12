package agent

import (
	"fmt"
	. "github.com/eac0de/getmetrics/internal/storage"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Agent struct {
	serverUrl      string
	pollInterval   time.Duration
	reportInterval time.Duration
}

func NewAgent(serverHost string, serverPort string, pollInterval time.Duration, reportInterval time.Duration) *Agent {
	serverUrl := "http://" + serverHost + ":" + serverPort
	return &Agent{
		serverUrl:      serverUrl,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
	}
}

func (a *Agent) Run() {
	var pollCount Counter
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
	Alloc         Gauge   `json:"alloc"`
	BuckHashSys   Gauge   `json:"buck_hash_sys"`
	Frees         Gauge   `json:"frees"`
	GCCPUFraction Gauge   `json:"gccpufraction"`
	GCSys         Gauge   `json:"gcsys"`
	HeapAlloc     Gauge   `json:"heap_alloc"`
	HeapIdle      Gauge   `json:"heap_idle"`
	HeapInuse     Gauge   `json:"heap_inuse"`
	HeapObjects   Gauge   `json:"heap_objects"`
	HeapReleased  Gauge   `json:"heap_released"`
	HeapSys       Gauge   `json:"heap_sys"`
	LastGC        Gauge   `json:"last_gc"`
	Lookups       Gauge   `json:"lookups"`
	MCacheInuse   Gauge   `json:"mcache_inuse"`
	MCacheSys     Gauge   `json:"mcache_sys"`
	MSpanInuse    Gauge   `json:"mspan_inuse"`
	MSpanSys      Gauge   `json:"mspan_sys"`
	Mallocs       Gauge   `json:"mallocs"`
	NextGC        Gauge   `json:"next_gc"`
	NumForcedGC   Gauge   `json:"num_forced_gc"`
	NumGC         Gauge   `json:"num_gc"`
	OtherSys      Gauge   `json:"other_sys"`
	PauseTotalNs  Gauge   `json:"pause_total_ns"`
	StackInuse    Gauge   `json:"stack_inuse"`
	StackSys      Gauge   `json:"stack_sys"`
	Sys           Gauge   `json:"sys"`
	TotalAlloc    Gauge   `json:"total_alloc"`
	PollCount     Counter `json:"poll_count"`
	RandomValue   Gauge   `json:"random_value"`
}

func (a *Agent) collectMetrics(pollCount *Counter) Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	*pollCount++

	return Metrics{
		Alloc:         Gauge(memStats.Alloc),
		BuckHashSys:   Gauge(memStats.BuckHashSys),
		Frees:         Gauge(memStats.Frees),
		GCCPUFraction: Gauge(memStats.GCCPUFraction),
		GCSys:         Gauge(memStats.GCSys),
		HeapAlloc:     Gauge(memStats.HeapAlloc),
		HeapIdle:      Gauge(memStats.HeapIdle),
		HeapInuse:     Gauge(memStats.HeapInuse),
		HeapObjects:   Gauge(memStats.HeapObjects),
		HeapReleased:  Gauge(memStats.HeapReleased),
		HeapSys:       Gauge(memStats.HeapSys),
		LastGC:        Gauge(memStats.LastGC),
		Lookups:       Gauge(memStats.Lookups),
		MCacheInuse:   Gauge(memStats.MCacheInuse),
		MCacheSys:     Gauge(memStats.MCacheSys),
		MSpanInuse:    Gauge(memStats.MSpanInuse),
		MSpanSys:      Gauge(memStats.MSpanSys),
		Mallocs:       Gauge(memStats.MSpanSys),
		NextGC:        Gauge(memStats.MSpanSys),
		NumForcedGC:   Gauge(memStats.MSpanSys),
		NumGC:         Gauge(memStats.MSpanSys),
		OtherSys:      Gauge(memStats.MSpanSys),
		PauseTotalNs:  Gauge(memStats.MSpanSys),
		StackInuse:    Gauge(memStats.MSpanSys),
		StackSys:      Gauge(memStats.MSpanSys),
		Sys:           Gauge(memStats.MSpanSys),
		TotalAlloc:    Gauge(memStats.MSpanSys),
		PollCount:     *pollCount,
		RandomValue:   Gauge(memStats.MSpanSys),
	}
}

func (a *Agent) sendMetric(metricType string, metricName string, metricValue interface{}) error {
	url := fmt.Sprintf("%s/update/%s/%s/%v", a.serverUrl, metricType, metricName, metricValue)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", body)
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
