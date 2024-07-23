package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/getmetrics/internal/agent"
	"github.com/eac0de/getmetrics/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agentConfig := config.NewAgentConfig()
	a := agent.NewAgent(agentConfig)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go a.Run(ctx)
	<-sigChan
	a.Stop(cancel)
}
