package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/getmetrics/internal/agent"
	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/pkg/semaphore"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewAgentConfig()

	a := agent.NewAgent(cfg)
	go a.StartPoll(ctx)
	go a.StartPoll2(ctx)

	sph := semaphore.NewSemaphore(cfg.RateLimit)

	go a.StartSendReport(ctx, sph)
	go a.StartSendReport2(ctx, sph)

	log.Println("Agent is running. Press Ctrl+C to stop")

	<-ctx.Done() // Блокируемся до закрытия канала done
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	<-sigChan
}
