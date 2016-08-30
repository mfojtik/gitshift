package processor

import (
	"log"
	"strconv"
	"strings"

	workq "github.com/iamduo/go-workq"
	"github.com/mfojtik/gitshift/pkg/api"
	apiclient "github.com/mfojtik/gitshift/pkg/client"
)

func ProcessComments() error {
	config := api.EnvToConfig()
	client, err := workq.Connect(config["WORKQ_SERVER_SERVICE_HOST"] + ":" + config["WORKQ_SERVER_SERVICE_PORT"])
	if err != nil {
		return err
	}
	github, err := apiclient.NewGithub()
	if err != nil {
		return err
	}
	defer client.Close()
	for {
		job, err := client.Lease([]string{"process-comment"}, 60000)
		if err != nil {
			log.Printf("ERROR:  leasing error, will retry %v", err)
			continue
		}
		parts := strings.Split(string(job.Payload), ":")
		if len(parts) != 2 {
			log.Printf("ERROR: unknown payload: %q", string(job.Payload))
			continue
		}
		commentStringID, prStringNumber := parts[1], parts[0]
		prNumber, _ := strconv.ParseInt(prStringNumber, 10, 64)
		commentID, _ := strconv.ParseInt(commentStringID, 10, 64)

		log.Printf("processing %q job %d for %d", "process-comment", job.ID, prNumber, commentID)

		currentPull := api.GetPull(int(prNumber))
		if currentPull == nil {
			log.Printf("pull %d is not cached (SKIP)", prNumber)
			client.Complete(job.ID, []byte(""))
			continue
		}

		comment := github.Comment(int(commentID))
		if comment == nil {
			if currentPull.MergeCommentID == int(commentID) {
				currentPull.MergeCommentID = 0
			}
			if currentPull.JenkinsTestCommentID == int(commentID) {
				currentPull.JenkinsTestCommentID = 0
			}
			// This can race
			if err := api.StorePullRequest(currentPull); err != nil {
				client.Fail(job.ID, []byte(""))
				continue
			}
			log.Printf("pull comment %d is not found (SKIP)(REMOVED)", commentID)
			client.Complete(job.ID, []byte(""))
			continue
		}

		result := *currentPull

		if status := parseMergeStatus(*comment.Body); len(status) > 0 {
			result.MergeStatus = status
			result.MergeURL = parseJenkinsURL(*comment.Body)
			result.MergeCommentID = *comment.ID
		}

		if status := parseJenkinsStatus(*comment.Body); len(status) > 0 {
			result.JenkinsTestStatus = status
			result.JenkinsTestURL = parseJenkinsURL(*comment.Body)
			result.JenkinsTestCommentID = *comment.ID
		}

		if position := parseMergeQueuePosition(*comment.Body); position > 0 {
			result.Position = position
		}

		if currentPull.Equal(&result) {
			log.Printf("version: %#+v is equal to version we have: %#+v (SKIP)", currentPull, &result)
			continue
		}
		if err := api.StorePullRequest(&result); err != nil {
			client.Fail(job.ID, []byte(""))
			continue
		}
		log.Printf("PR#%d updated: %+v", result.Number, result)
		client.Complete(job.ID, []byte(""))
	}
}
