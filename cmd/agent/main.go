package main

import (
	. "github.com/eac0de/getmetrics/internal/agent"
	"time"
)

func main() {
	agent := NewAgent("127.0.0.1", "8080", 2*time.Second, 10*time.Second)
	agent.Run()
}
