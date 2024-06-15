package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	a "github.com/eac0de/getmetrics/internal/agent"
)

type AgentConfig struct {
	ServerURL      string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func parseFlags(c *AgentConfig) {
	flag.StringVar(&c.ServerURL, "a", "localhost:8080", "server address")
	flag.IntVar(&c.ReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&c.PollInterval, "p", 2, "poll interval in seconds")
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
	pollInterval := time.Duration(agentConfig.PollInterval) * time.Second
	reportInterval := time.Duration(agentConfig.ReportInterval) * time.Second
	fmt.Println("Server URL:", agentConfig.ServerURL)
	fmt.Println("Report Interval:", reportInterval)
	fmt.Println("Poll Interval:", pollInterval)
	agent := a.NewAgent(agentConfig.ServerURL, pollInterval, reportInterval)
	agent.Run()
}
