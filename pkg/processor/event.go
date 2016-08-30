package processor

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	workq "github.com/iamduo/go-workq"
	"github.com/mfojtik/gitshift/pkg/api"

	gh "github.com/google/go-github/github"
)

func ProcessEvents() error {
	config := api.EnvToConfig()
	client, err := workq.Connect(config["WORKQ_SERVER_SERVICE_HOST"] + ":" + config["WORKQ_SERVER_SERVICE_PORT"])
	if err != nil {
		return err
	}
	defer client.Close()
	for {
		job, err := client.Lease([]string{"process-event"}, 60000)
		if err != nil {
			log.Printf("ERROR:  leasing error, will retry %v", err)
			continue
		}
		event := new(gh.Event)
		if err := json.Unmarshal(job.Payload, event); err != nil {
			log.Printf("ERROR: unable to decode %q", job.ID)
			client.Fail(job.ID, []byte(""))
			continue
		}
		defer client.Complete(job.ID, []byte(""))

		log.Printf("processing %q job %q for %q", "process-event", job.ID, *event.Type)
		// process issue comment events
		if comment, ok := event.Payload().(*gh.IssueCommentEvent); ok {

			// openshift-bot is special
			user := *comment.Comment.User.Login
			if user == "openshift-bot" {
				if err := ProcessOpenShiftBotComment(comment.Issue, comment.Comment); err != nil {
					log.Printf("ERROR: unable to process openshift-bot comment: %v", err)
				}
			}
		}
	}
}

func ProcessOpenShiftBotComment(issue *gh.Issue, comment *gh.IssueComment) error {
	result := new(api.PullRequest)
	currentVersion := api.GetPull(*issue.Number)

	result.Number = *issue.Number
	result.Title = *issue.Title
	result.CreatedAt = *comment.CreatedAt
	result.Author = *issue.User.Login

	if comment.UpdatedAt != nil {
		result.CreatedAt = *comment.UpdatedAt
	}

	if status := parseJenkinsStatus(*comment.Body); len(status) > 0 {
		result.JenkinsTestStatus = status
	}

	if position := parseMergeQueuePosition(*comment.Body); position > 0 {
		result.Position = position
	}

	if isMerge := parseMergeStatus(*comment.Body); isMerge != nil {
		result.Merge = true
	}

	if !result.IsLatest(currentVersion) || result.Equal(currentVersion) {
		return nil
	}

	return api.StorePullRequest(result)
}

func parseBotStatus(status string) string {
	parts := strings.Split(status, " ")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToUpper(strings.TrimSuffix(strings.TrimSpace(parts[1]), ":"))
}

func parseJenkinsStatus(msg string) string {
	if !strings.Contains(msg, "openshift-jenkins/test") {
		return ""
	}
	return parseBotStatus(msg)
}

func parseMergeStatus(msg string) *bool {
	result := true
	if !strings.Contains(msg, "openshift-jenkins/merge") {
		return nil
	}
	return &result
}

func parseMergeQueuePosition(msg string) int {
	if !strings.Contains(msg, "You are in the build queue at position:") {
		return 0
	}
	parts := strings.Split(msg, ":")
	if len(parts) < 3 {
		return 0
	}
	pos, _ := strconv.ParseInt(parts[2], 10, 64)
	return int(pos)
}
