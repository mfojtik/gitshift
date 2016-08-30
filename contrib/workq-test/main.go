package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	workq "github.com/iamduo/go-workq"
	"github.com/satori/go.uuid"
)

func envToMap() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		if len(parts) < 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result
}

func main() {
	config := envToMap()
	counter := 0
	go func() {
		for {
			client, err := workq.Connect(config["WORKQ_SERVER_SERVICE_HOST"] + ":" + config["WORKQ_SERVER_SERVICE_PORT"])
			if err != nil {
				log.Printf("ERROR: Failed to connect to workq-server: %v", err)
				os.Exit(1)
			}
			log.Printf("Awaiting incoming 'sample-job' jobs ...")
			job, err := client.Lease([]string{"sample-job"}, 60000)
			if err != nil {
				log.Printf("ERROR: Leasing job failed: %v", err)
			}
			log.Printf("Leased Job: ID: %s, Name: %s, Payload: %s", job.ID, job.Name, string(job.Payload))

			// Do the actual work here
			time.Sleep(1 * time.Second)
			if err := client.Complete(job.ID, []byte(fmt.Sprintf("completed #%d", job.Payload))); err != nil {
				log.Printf("ERROR: Failed to complete job: %v", err)
			}
			client.Close()
		}
	}()

	for {
		client, err := workq.Connect(config["WORKQ_SERVER_SERVICE_HOST"] + ":" + config["WORKQ_SERVER_SERVICE_PORT"])
		if err != nil {
			log.Printf("ERROR: Failed to connect to workq-server: %v", err)
			os.Exit(1)
		}
		counter++
		jobID := uuid.NewV4()
		job := &workq.FgJob{
			ID:      jobID.String(),
			Name:    fmt.Sprintf("sample-job"),
			TTR:     5000,  // 5 second time-to-run limit
			Timeout: 60000, // Wait up to 60 seconds for a worker to pick up.
			Payload: []byte(fmt.Sprintf("ping #%d", counter)),
		}
		log.Printf("Running job %d", counter)
		result, err := client.Run(job)
		if err != nil {
			log.Printf("ERROR: Scheduling job #%d failed: %v", err)
		}
		log.Printf("#%d: Success: %t, Result: %s", counter, result.Success, string(result.Result))
		client.Close()
	}
}
