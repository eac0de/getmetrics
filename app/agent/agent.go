package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/models"
	"github.com/eac0de/getmetrics/pkg/compressor"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	conf      *config.AgentConfig
	client    *resty.Client
	metrics   *Metrics
	pollCount int64
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
	log.Println("Agent stopped")
}

func (a *Agent) Run(ctx context.Context) {

	go a.StartPoll(ctx)

	go a.StartSendReport(ctx)

	log.Println("Agent is running. Press Ctrl+C to stop")
	<-ctx.Done() // Блокируемся до закрытия канала done
}

func (a *Agent) StartPoll(ctx context.Context) {
	ticker := time.NewTicker(a.conf.PollInterval)
	for {
		select {
		case <-ctx.Done():
			log.Println("Poll goroutine is shutting down...")
			return
		case <-ticker.C:
			a.metrics = a.collectMetrics()
		}
	}
}

func (a *Agent) StartSendReport(ctx context.Context) {
	ticker := time.NewTicker(a.conf.ReportInterval)
	for {
		select {
		case <-ctx.Done():
			log.Println("Goroutine sending reports has been shut down...")
			return
		case <-ticker.C:
			attempts, err := SendMetricsWithRetry(a.sendMetrics, a.metrics)
			if err != nil {
				log.Printf("Not a single attempt has been successful, attemts count: %v\n", attempts)
			}

		}
	}
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

func (a *Agent) collectMetrics() *Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	a.pollCount++
	return &Metrics{
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
		PollCount:     a.pollCount,
		RandomValue:   float64(memStats.MSpanSys),
	}
}

func (a *Agent) sendMetrics(metrics *Metrics) error {
	values := models.MetricsDict{
		Gauge: map[string]float64{
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
			"RandomValue":   metrics.RandomValue,
		},
		Counter: map[string]int64{
			"PollCount": metrics.PollCount,
		},
	}
	metricsList := []models.Metrics{}
	for metricName, metricValue := range values.Gauge {
		metricsList = append(metricsList, models.Metrics{ID: metricName, MType: models.Gauge, Value: &metricValue})
	}
	for metricName, metricDelta := range values.Counter {
		metricsList = append(metricsList, models.Metrics{ID: metricName, MType: models.Counter, Delta: &metricDelta})
	}
	metricsListJSON, err := json.Marshal(metricsList)
	if err != nil {
		return err
	}
	metricGzip, err := compressor.GzipData(metricsListJSON)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/updates/", a.conf.ServerURL)
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
		return fmt.Errorf("send metrics error: %s", resp.Body())
	}
	return nil
}

func SendMetricsWithRetry(sendMetricsFunc func(*Metrics) error, metrics *Metrics) (uint8, error) {
	var err error
	var attemtsCount uint8
	for waitTime := 1; waitTime <= 5; waitTime += 2 {
		attemtsCount += 1
		err = sendMetricsFunc(metrics)
		if err == nil {
			return attemtsCount, nil
		}
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
	return attemtsCount, err
}
