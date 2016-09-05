package processor

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/mfojtik/gitshift/pkg/api"
	"github.com/mfojtik/gitshift/pkg/client"

	gh "github.com/google/go-github/github"
)

func ProcessEvents() error {
	workq, _, _, err := client.GetAll()
	if err != nil {
		return err
	}
	defer workq.Close()
	for {
		job, err := workq.Lease([]string{"process-event"}, 60000)
		if err != nil {
			log.Printf("ERROR:  leasing error, will retry %v", err)
			continue
		}
		event := new(gh.Event)
		if err := json.Unmarshal(job.Payload, event); err != nil {
			log.Printf("ERROR: unable to decode %q", job.ID)
			workq.Complete(job.ID, []byte(""))
			continue
		}
		workq.Complete(job.ID, []byte(""))

		// process issue comment events
		if comment, ok := event.Payload().(*gh.IssueCommentEvent); ok {
			if err := ProcessOpenShiftBotComment(comment.Issue, comment.Comment); err != nil {
				log.Printf("ERROR: unable to process openshift-bot comment: %v", err)
			}
			if err := ProcessApprovedComment(comment.Issue, comment.Comment); err != nil {
				log.Printf("ERROR: unable to process LGTM comment: %v", err)
			}
			if err := ProcessMilestone(comment.Issue); err != nil {
				log.Printf("ERROR: unable to process milestone: %v", err)
			}
		}
	}
}

func ProcessApprovedComment(issue *gh.Issue, comment *gh.IssueComment) error {
	if !strings.Contains(strings.ToLower(*comment.Body), "lgtm") {
		return nil
	}
	currentVersion := client.GetPull(*issue.Number)
	if currentVersion == nil {
		log.Printf("not tracking %d, skipping", *issue.Number)
		return nil
	}
	currentVersion.Approved = true
	return client.StorePullRequest(currentVersion)
}

func ProcessMilestone(issue *gh.Issue) error {
	if issue.Milestone == nil {
		return nil
	}
	currentVersion := client.GetPull(*issue.Number)
	if currentVersion == nil {
		return nil
	}
	milestone := *issue.Milestone.Title
	if len(milestone) == 0 {
		return nil
	}
	if currentVersion.Milestone == milestone {
		return nil
	}
	currentVersion.Milestone = milestone
	return client.StorePullRequest(currentVersion)
}

func ProcessOpenShiftBotComment(issue *gh.Issue, comment *gh.IssueComment) error {
	if *comment.User.Login != "openshift-bot" {
		return nil
	}
	// Ignore "Evaluated for origin merge up to 5d4ca73" comments
	if strings.Contains(*comment.Body, "Evaluated for") {
		return nil
	}
	result := new(api.PullRequest)
	currentVersion := client.GetPull(*issue.Number)

	result.Number = *issue.Number
	result.Title = *issue.Title
	result.Author = *issue.User.Login

	updateFromComment(comment, result)

	if !result.IsLatest(currentVersion) {
		log.Printf("version: %#+v is older than version we have: %#+v (SKIP)", result, currentVersion)
		return nil
	}

	if result.Equal(currentVersion) {
		log.Printf("version: %#+v is equal to version we have: %#+v (SKIP)", result, currentVersion)
		return nil
	}

	return client.StorePullRequest(result)
}
