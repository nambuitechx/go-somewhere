package main

import (
	"encoding/json"
	"time"
)

// Job is the persisted job definition stored in the schedule zset.
type Job struct {
	Name string `json:"name"`
	Interval string `json:"interval"` // Duration string like "5m"
	NextRun int64 `json:"next_run"` // Unix timestamp
	Payload map[string]interface{} `json:"payload,omitempty"`
}

func (j *Job) Marshal() ([]byte, error) {
	return json.Marshal(j)
}

func UnmarshalJob(b []byte) (*Job, error) {
	var j Job
	if err := json.Unmarshal(b, &j); err != nil {
		return nil, err
	}
	return &j, nil
}

func NowUnix() int64 { return time.Now().Unix() }
