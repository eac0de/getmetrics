package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/eac0de/getmetrics/internal/agent"
	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/pkg/utils"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Build version: %s\n", utils.GetValueOrDefault(buildVersion))
	fmt.Printf("Build date: %s\n", utils.GetValueOrDefault(buildDate))
	fmt.Printf("Build commit: %s\n", utils.GetValueOrDefault(buildCommit))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatal(err)
	}
	a := agent.NewAgent(cfg)
	var wg sync.WaitGroup
	wg.Add(2)
	go a.StartPoll(ctx, &wg)
	go a.StartSendReport(ctx, &wg)

	log.Println("Agent is running. Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigChan
	cancel()
	wg.Wait()
}
