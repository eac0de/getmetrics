package main

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

var serverUrl string = "http://localhost:8080"

type gauge float64
type counter int64

type Metrics struct {
	Alloc         gauge   `json:"alloc"`
	BuckHashSys   gauge   `json:"buck_hash_sys"`
	Frees         gauge   `json:"frees"`
	GCCPUFraction gauge   `json:"gccpufraction"`
	GCSys         gauge   `json:"gcsys"`
	HeapAlloc     gauge   `json:"heap_alloc"`
	HeapIdle      gauge   `json:"heap_idle"`
	HeapInuse     gauge   `json:"heap_inuse"`
	HeapObjects   gauge   `json:"heap_objects"`
	HeapReleased  gauge   `json:"heap_released"`
	HeapSys       gauge   `json:"heap_sys"`
	LastGC        gauge   `json:"last_gc"`
	Lookups       gauge   `json:"lookups"`
	MCacheInuse   gauge   `json:"mcache_inuse"`
	MCacheSys     gauge   `json:"mcache_sys"`
	MSpanInuse    gauge   `json:"mspan_inuse"`
	MSpanSys      gauge   `json:"mspan_sys"`
	Mallocs       gauge   `json:"mallocs"`
	NextGC        gauge   `json:"next_gc"`
	NumForcedGC   gauge   `json:"num_forced_gc"`
	NumGC         gauge   `json:"num_gc"`
	OtherSys      gauge   `json:"other_sys"`
	PauseTotalNs  gauge   `json:"pause_total_ns"`
	StackInuse    gauge   `json:"stack_inuse"`
	StackSys      gauge   `json:"stack_sys"`
	Sys           gauge   `json:"sys"`
	TotalAlloc    gauge   `json:"total_alloc"`
	PollCount     counter `json:"poll_count"`
	RandomValue   gauge   `json:"random_value"`
}

func main() {
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	var pollCount counter
	var metrics Metrics

	go func() {
		for {
			metrics = collectMetrics(&pollCount)
			time.Sleep(pollInterval)

		}
	}()

	go func() {
		for {
			sendMetrics(metrics)
			time.Sleep(reportInterval)
		}
	}()

	select {}
}

func collectMetrics(pollCount *counter) Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	*pollCount++

	return Metrics{
		Alloc:         gauge(memStats.Alloc),
		BuckHashSys:   gauge(memStats.BuckHashSys),
		Frees:         gauge(memStats.Frees),
		GCCPUFraction: gauge(memStats.GCCPUFraction),
		GCSys:         gauge(memStats.GCSys),
		HeapAlloc:     gauge(memStats.HeapAlloc),
		HeapIdle:      gauge(memStats.HeapIdle),
		HeapInuse:     gauge(memStats.HeapInuse),
		HeapObjects:   gauge(memStats.HeapObjects),
		HeapReleased:  gauge(memStats.HeapReleased),
		HeapSys:       gauge(memStats.HeapSys),
		LastGC:        gauge(memStats.LastGC),
		Lookups:       gauge(memStats.Lookups),
		MCacheInuse:   gauge(memStats.MCacheInuse),
		MCacheSys:     gauge(memStats.MCacheSys),
		MSpanInuse:    gauge(memStats.MSpanInuse),
		MSpanSys:      gauge(memStats.MSpanSys),
		Mallocs:       gauge(memStats.MSpanSys),
		NextGC:        gauge(memStats.MSpanSys),
		NumForcedGC:   gauge(memStats.MSpanSys),
		NumGC:         gauge(memStats.MSpanSys),
		OtherSys:      gauge(memStats.MSpanSys),
		PauseTotalNs:  gauge(memStats.MSpanSys),
		StackInuse:    gauge(memStats.MSpanSys),
		StackSys:      gauge(memStats.MSpanSys),
		Sys:           gauge(memStats.MSpanSys),
		TotalAlloc:    gauge(memStats.MSpanSys),
		PollCount:     *pollCount,
		RandomValue:   gauge(memStats.MSpanSys),
	}
}

func sendMetric(metricType string, metricName string, metricValue interface{}) error {
	url := fmt.Sprintf("%s/update/%s/%s/%v", serverUrl, metricType, metricName, metricValue)
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
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", body)
	}
	return nil
}

func sendMetrics(metrics Metrics) {
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

	for name, value := range values {
		if err := sendMetric(reflect.TypeOf(value).Name(), name, value); err != nil {
			fmt.Printf("failed to send metric %s: %s\n", name, err.Error())
		}
	}
}
