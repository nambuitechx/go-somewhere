package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ScheduleZSet = "cron:schedule"
	QueueList = "cron:queue"
	LockPrefix = "cron:lock:"
)


// JobSpec is a simple job declaration used for scheduling
type JobSpec struct {
	Name string
	Interval string
	Payload map[string]string
}


type Scheduler struct {
	rdb *Client
}

func NewScheduler(rdb *Client) *Scheduler {
	return &Scheduler{rdb: rdb}
}

// Schedule inserts a job into the schedule zset. If job exists, we overwrite next_run.
func (s *Scheduler) Schedule(ctx context.Context, js JobSpec) error {
	interval, err := time.ParseDuration(js.Interval)
	if err != nil {
		return err
	}

	next := time.Now().Add(interval).Unix()
	job := Job{
		Name: js.Name,
		Interval: js.Interval,
		NextRun: next,
		Payload: mapFrom(js.Payload),
	}
	b, _ := job.Marshal()

	return s.rdb.ZAdd(ctx, ScheduleZSet, redis.Z{Score: float64(next), Member: b}).Err()
}

func mapFrom(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// Run starts the poller which moves due jobs to the queue and reschedules them
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		case <-ticker.C:
			if err := s.pollOnce(ctx); err != nil {
				log.Println("Poll error")
			}
		}
	}
}

func (s *Scheduler) pollOnce(ctx context.Context) error {
	now := float64(time.Now().Unix())

	// ZRANGEBYSCORE with limit 0..100 to avoid huge batches
	vals, err := s.rdb.ZRangeByScore(ctx, ScheduleZSet, &redis.ZRangeBy{Min: "-inf", Max: fmt.Sprintf("%f", now), Count: 100}).Result()

	if err != nil {
		return err
	}

	if len(vals) == 0 {
		return nil
	}

	for _, v := range vals {
		// Try to acquire a lock using SETNX with small TTL
		var job Job

		if err := json.Unmarshal([]byte(v), &job); err != nil {
			log.Println("Invalid job json")
			// Remove corrupt entry
			s.rdb.ZRem(ctx, ScheduleZSet, v)
			continue
		}

		lockKey := LockPrefix + job.Name
		ok, err := s.rdb.SetNX(ctx, lockKey, "1", 30*time.Second).Result()

		if err != nil {
			log.Println("Lock check err")
			continue
		}

		if !ok {
			// Someone else owns it
			continue
		}

		// Push to queue
		if err := s.rdb.LPush(ctx, QueueList, v).Err(); err != nil {
			log.Println("Push queue")
			// Release lock
			s.rdb.Del(ctx, lockKey)
			continue
		}

		// Reschedule next run
		interval, err := time.ParseDuration(job.Interval)

		if err != nil {
			log.Println("Parse interval")
			// Cleanup
			s.rdb.Del(ctx, lockKey)
			s.rdb.ZRem(ctx, ScheduleZSet, v)
			continue
		}

		next := time.Now().Add(interval).Unix()
		job.NextRun = next
		nb, _ := json.Marshal(job)

		pipe := s.rdb.TxPipeline()
		pipe.ZAdd(ctx, ScheduleZSet, redis.Z{Score: float64(next), Member: nb})
		pipe.ZRem(ctx, ScheduleZSet, v)

		// Release lock after scheduling
		pipe.Del(ctx, lockKey)
		_, err = pipe.Exec(ctx)

		if err != nil {
			log.Println("Reschedule pipe")
		}
	}

	return nil
}

// helper error for clarity
var ErrNotFound = errors.New("not found")
