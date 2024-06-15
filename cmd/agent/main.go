package main

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	a "github.com/eac0de/getmetrics/internal/agent"
)

type AgentConfig struct {
	ServerURL      string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

func parseFlags(c *AgentConfig) {
	flag.StringVar(&c.ServerURL, "a", "localhost:8080", "server address")
	flag.DurationVar(&c.ReportInterval, "r", 10*time.Second, "report interval")
	flag.DurationVar(&c.PollInterval, "p", 2*time.Second, "poll interval")

	flag.Parse()
}

func parseEnv(c *AgentConfig) {
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	agentConfig := new(AgentConfig)
	parseFlags(agentConfig)
	parseEnv(agentConfig)
	agent := a.NewAgent(agentConfig.ServerURL, agentConfig.PollInterval, agentConfig.ReportInterval)
	agent.Run()
}
