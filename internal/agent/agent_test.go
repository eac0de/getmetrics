package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	var cfg config.AgentConfig
	serverURL := "localhost:8080"
	cfg.ServerURL = serverURL
	agent := NewAgent(&cfg)
	assert.Equal(t, agent.cfg.ServerURL, "http://"+serverURL)

}

func TestStartPoll(t *testing.T) {
	var cfg config.AgentConfig
	cfg.PollInterval = 10 * time.Second
	agent := NewAgent(&cfg)

	var wg sync.WaitGroup
	wg.Add(1) // Увеличиваем счетчик

	context, cancel := context.WithCancel(context.Background())

	go func() {
		agent.StartPoll(context, &wg)
	}()

	cancel()  // Отменяем контекст
	wg.Wait() // Ждем, пока горутина завершится
}

func TestStartSendReport(t *testing.T) {
	var cfg config.AgentConfig
	cfg.ReportInterval = 10 * time.Second
	agent := NewAgent(&cfg)

	var wg sync.WaitGroup
	wg.Add(1) // Увеличиваем счетчик

	context, cancel := context.WithCancel(context.Background())
	go func() {
		agent.StartSendReport(context, &wg)
	}()

	cancel()  // Отменяем контекст
	wg.Wait() // Ждем, пока горутина завершится
}

func TestCollectMetrics(t *testing.T) {
	var cfg config.AgentConfig
	agent := NewAgent(&cfg)
	agent.pollCount = 5
	agent.collectMetrics()
	assert.Equal(t, agent.pollCount, int64(6))
}
