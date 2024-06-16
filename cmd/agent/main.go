package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/eac0de/getmetrics/internal/agent"
)

type AgentConfig struct {
	ServerURL      string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

const (
	defaultServerURL      = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
)

func readAgentFlags(c *AgentConfig) {
	var pollInterval int
	var reportInterval int
	flag.StringVar(&c.ServerURL, "a", defaultServerURL, "server address")
	flag.IntVar(&pollInterval, "r", defaultPollInterval, "report interval in seconds")
	flag.IntVar(&reportInterval, "p", defaultReportInterval, "poll interval in seconds")
	flag.Parse()
	c.PollInterval = time.Duration(pollInterval) * time.Second
	c.ReportInterval = time.Duration(reportInterval) * time.Second
}

func readEnvConfig(c *AgentConfig) {
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	agentConfig := new(AgentConfig)
	readAgentFlags(agentConfig)
	readEnvConfig(agentConfig)
	fmt.Println("Server URL:", agentConfig.ServerURL)
	fmt.Println("Report Interval:", agentConfig.ReportInterval)
	fmt.Println("Poll Interval:", agentConfig.PollInterval)
	a := agent.NewAgent(agentConfig.ServerURL, agentConfig.PollInterval, agentConfig.ReportInterval)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {
		a.Run()
	}()
	<-sigChan
	a.Stop()
	log.Println("Agent stopped.")
}
