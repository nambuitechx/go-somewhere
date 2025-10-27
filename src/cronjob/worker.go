package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Worker struct {
	rdb *Client
}

func NewWorker(rdb *Client) *Worker {
	return &Worker{rdb: rdb}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		default:
			// Block pop with timeout so we can respond to context cancellation
			res, err := w.rdb.BRPop(ctx, 5*time.Second, QueueList).Result()

			if err != nil {
				// If context canceled, exit
				if ctx.Err() != nil {
					return
				}

				// Redis timeout or other error, continue
				continue
			}

			if len(res) < 2 {
				continue
			}

			b := []byte(res[1])
			job, err := UnmarshalJob(b)

			if err != nil {
				log.Println("Invalid job data")
				continue
			}


			// Execute job — replace with real handler
			log.Println("Executing job")
			execErr := w.execute(job)

			if execErr != nil {
				log.Println("Job failed")
			}
		}
	}
}

func (w *Worker) execute(j *Job) error {
	// This is a stub: map payload and perform action accordingly
	if cmd, ok := j.Payload["cmd"].(string); ok {
		// Simulate work
		fmt.Println("Running command:", cmd)
		time.Sleep(2 * time.Second)
		fmt.Println("Done:", cmd)
		return nil
	}

	return fmt.Errorf("no cmd in payload")
}
