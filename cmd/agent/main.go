package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/getmetrics/internal/agent"
	"github.com/eac0de/getmetrics/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewAgentConfig()

	a := agent.NewAgent(cfg)
	go a.StartPoll(ctx)
	go a.StartSendReport(ctx)

	log.Println("Agent is running. Press Ctrl+C to stop")

	<-ctx.Done() // Блокируемся до закрытия канала done
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	<-sigChan
}
