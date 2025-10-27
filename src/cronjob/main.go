package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main()  {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup redis
	addr := getEnv("REDIS_ADDR", "redis:6379")
	rdb := NewClient(Config{Addr: addr})
	defer rdb.Close()

	// Start scheduler and worker
	s := NewScheduler(rdb)
	w := NewWorker(rdb)

	go s.Run(ctx)
	go w.Run(ctx)

	// Demo: schedule a sample job if environment says so
	if getEnv("SCHEDULE_DEMO", "true") == "true" {
		job := JobSpec{
			Name: "cleanup_demo",
			Interval: "10s",
			Payload: map[string]string{"cmd": "cleanup_files"},
		}
		
		if err := s.Schedule(ctx, job); err != nil {
			log.Println("Failed schedule demo job")
		}
	}

	// Wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("Shutting down...")
	cancel()

	// Give goroutines some time to stop
	time.Sleep(500 * time.Millisecond)
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	
	return d
}
