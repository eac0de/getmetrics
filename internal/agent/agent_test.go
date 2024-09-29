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
	var conf config.AgentConfig
	serverURL := "localhost:8080"
	conf.ServerURL = serverURL
	agent := NewAgent(&conf)
	assert.Equal(t, agent.conf.ServerURL, "http://"+serverURL)

}

func TestStartPoll(t *testing.T) {
	var conf config.AgentConfig
	conf.PollInterval = 10 * time.Second
	agent := NewAgent(&conf)

	var wg sync.WaitGroup
	wg.Add(1) // Увеличиваем счетчик

	context, cancel := context.WithCancel(context.Background())

	go func() {
		defer wg.Done() // Уменьшаем счетчик по завершению
		agent.StartPoll(context)
	}()

	cancel()  // Отменяем контекст
	wg.Wait() // Ждем, пока горутина завершится
}

func TestStartSendReport(t *testing.T) {
	var conf config.AgentConfig
	conf.ReportInterval = 10 * time.Second
	agent := NewAgent(&conf)

	var wg sync.WaitGroup
	wg.Add(1) // Увеличиваем счетчик

	context, cancel := context.WithCancel(context.Background())
	go func() {
		defer wg.Done() // Уменьшаем счетчик по завершению
		agent.StartSendReport(context)
	}()

	cancel()  // Отменяем контекст
	wg.Wait() // Ждем, пока горутина завершится
}

func TestCollectMetrics(t *testing.T) {
	var conf config.AgentConfig
	agent := NewAgent(&conf)
	agent.pollCount = 5
	agent.collectMetrics()
	assert.Equal(t, agent.pollCount, int64(6))
}
