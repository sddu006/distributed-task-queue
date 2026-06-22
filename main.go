package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task-queue/api"
	"task-queue/queue"
	"task-queue/store"
	"task-queue/worker"
)

func main() {
	fmt.Println("Starting Task Queue System...")

	// 1. Create the store with AOF persistence
	s, err := store.NewStore("aof.log")
	if err != nil {
		log.Fatal("Failed to create store:", err)
	}
	defer s.Close()

	// 2. Create the queue with buffer size 100
	q := queue.NewQueue(100)

	// 3. Create and start worker pool with 5 workers
	wp := worker.NewWorkerPool(5, q, s)
	wp.Start()

	// 4. Create and start API server
	server := api.NewServer(s, q)

	// 5. Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go server.Run("8080")

	<-quit
	fmt.Println("\nShutting down...")
	wp.Stop()
	fmt.Println("Goodbye!")
}
