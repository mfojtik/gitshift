package client

import (
	"time"

	workq "github.com/iamduo/go-workq"
	uuid "github.com/satori/go.uuid"
)

type Job struct {
	UUID string
	Name string
}

func AddJob(name string, payload []byte, ttr, ttl time.Duration) (*Job, error) {
	jobClient, _, _, err := GetAll()
	if err != nil {
		return nil, err
	}
	defer jobClient.Close()
	job := &Job{
		UUID: uuid.NewV4().String(),
		Name: name,
	}
	if err := jobClient.Add(&workq.BgJob{
		ID:          job.UUID,
		Name:        job.Name,
		TTR:         int(ttr / 1000000),
		TTL:         int(ttl / 1000000),
		Payload:     payload,
		Priority:    0,
		MaxAttempts: 3,
		MaxFails:    3,
	}); err != nil {
		return nil, err
	}
	return job, nil
}
