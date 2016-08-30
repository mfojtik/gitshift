package client

import (
	"time"

	workq "github.com/iamduo/go-workq"
	"github.com/mfojtik/gitshift/pkg/api"
	uuid "github.com/satori/go.uuid"
)

type Job struct {
	UUID string
	Name string
}

func AddJob(name string, payload []byte, ttr, ttl time.Duration) (*Job, error) {
	config := api.EnvToConfig()
	client, err := workq.Connect(config["WORKQ_SERVER_SERVICE_HOST"] + ":" + config["WORKQ_SERVER_SERVICE_PORT"])
	if err != nil {
		return nil, err
	}
	defer client.Close()
	job := &Job{
		UUID: uuid.NewV4().String(),
		Name: name,
	}
	if err := client.Add(&workq.BgJob{
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
