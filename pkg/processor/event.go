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
	// Ignore "Evaluated for origin merge up to 5d4ca73" comments
	if strings.Contains(*comment.Body, "Evaluated for") {
		return nil
	}
	result := new(api.PullRequest)
	currentVersion := api.GetPull(*issue.Number)

	result.Number = *issue.Number
	result.Title = *issue.Title
	result.CreatedAt = *comment.CreatedAt
	result.Author = *issue.User.Login

	if comment.UpdatedAt != nil {
		result.CreatedAt = *comment.UpdatedAt
	}

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

	log.Printf("PR#%d got: %#+v", result.Number, result)

	if !result.IsLatest(currentVersion) {
		log.Printf("version: %#+v is older than version we have: %#+v (SKIP)", result, currentVersion)
		return nil
	}

	if result.Equal(currentVersion) {
		log.Printf("version: %#+v is equal to version we have: %#+v (SKIP)", result, currentVersion)
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

func parseJenkinsURL(msg string) string {
	if !strings.Contains(msg, "continuous-integration/openshift-jenkins") {
		return ""
	}
	parts := strings.Split(msg, "(")
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSuffix(parts[1], ")")
}

func parseMergeStatus(msg string) string {
	if !strings.Contains(msg, "openshift-jenkins/merge") {
		return ""
	}
	return parseBotStatus(msg)
}

func parseMergeQueuePosition(msg string) int {
	if !strings.Contains(msg, "You are in the build queue at position:") {
		return 0
	}
	parts := strings.Split(msg, ":")
	if len(parts) < 3 {
		return 0
	}
	pos, _ := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
	return int(pos)
}
