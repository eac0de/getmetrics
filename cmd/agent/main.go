package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/eac0de/getmetrics/app/agent"
	"github.com/eac0de/getmetrics/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agentConfig := config.NewAgentConfig()
	a := agent.NewAgent(agentConfig)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	var wg sync.WaitGroup
	wg.Add(1)
	go a.Run(ctx)
	go a.Stop(ctx, &wg)
	<-sigChan
	cancel()
	wg.Wait()
}
