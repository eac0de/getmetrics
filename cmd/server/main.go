package main

import (
	. "github.com/eac0de/getmetrics/internal/server"
)

func main() {
	server := NewServer("127.0.0.1", "8080")
	server.Run()
}
