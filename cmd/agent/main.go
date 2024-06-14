package main

import (
	"flag"
	"time"

	a "github.com/eac0de/getmetrics/internal/agent"
)

var serverURL string
var pollInterval time.Duration
var reportInterval time.Duration

func parseFlags() {
	flag.StringVar(&serverURL, "a", "localhost:8080", "server address")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "report interval")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "poll interval")

	flag.Parse()
}

func main() {
	parseFlags()
	agent := a.NewAgent(serverURL, pollInterval, reportInterval)
	agent.Run()
}
