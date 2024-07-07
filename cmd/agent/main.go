package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/getmetrics/internal/agent"
	"github.com/eac0de/getmetrics/internal/config"
)

func main() {
	agentConfig := config.NewAgentConfig()
	a := agent.NewAgent(agentConfig)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {
		a.Run()
	}()
	<-sigChan
	a.Stop()
	log.Println("Agent stopped.")
}
