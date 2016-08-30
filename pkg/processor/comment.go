package processor

import (
	"log"
	"strings"

	gh "github.com/google/go-github/github"
	"github.com/mfojtik/gitshift/pkg/api"
	"github.com/mfojtik/gitshift/pkg/client"
)

func ProcessComments() error {
	workq, github, _, err := client.GetAll()
	if err != nil {
		return err
	}
	defer workq.Close()
	for {
		job, err := workq.Lease([]string{"process-comment"}, 60000)
		if err != nil {
			log.Printf("ERROR:  leasing error, will retry %v", err)
			continue
		}
		parts := strings.Split(string(job.Payload), ":")
		if len(parts) != 2 {
			log.Printf("ERROR: unknown payload: %q", string(job.Payload))
			continue
		}

		// Get the current version of pull request from redis or complete this job
		// if there is no version stored.
		current := client.GetPull(api.StringToInt(parts[0]))
		if current == nil {
			log.Printf("nothing to update for #%s because it is not cached", parts[0])
			workq.Complete(job.ID, []byte(""))
			continue
		}

		// Get the fresh comment body from Github.
		// TODO: This cost Github API request and need to be limited
		comment := github.Comment(api.StringToInt(parts[1]))
		if comment == nil {
			workq.Complete(job.ID, []byte(""))
			continue
		}

		result := *current
		updateFromComment(comment, &result)

		if current.Equal(&result) {
			log.Printf("skipping update (%d): %s == %s", current.Number, current, &result)
			continue
		}

		if err := client.StorePullRequest(&result); err != nil {
			workq.Fail(job.ID, []byte(""))
			log.Printf("ERROR: failed to store pull request: %v (will retry)", err)
			continue
		}

		log.Printf("updated: %s", result.Number, result)
		workq.Complete(job.ID, []byte(""))
	}
}

func updateFromComment(comment *gh.IssueComment, pr *api.PullRequest) {
	if comment.UpdatedAt != nil {
		pr.CreatedAt = *comment.UpdatedAt
	}

	if status := parseMergeStatus(*comment.Body); len(status) > 0 {
		pr.MergeStatus = status
		pr.MergeURL = parseJenkinsURL(*comment.Body)
		pr.MergeCommentID = *comment.ID
	}

	if status := parseJenkinsStatus(*comment.Body); len(status) > 0 {
		pr.JenkinsTestStatus = status
		pr.JenkinsTestURL = parseJenkinsURL(*comment.Body)
		pr.JenkinsTestCommentID = *comment.ID
	}

	if position := parseMergeQueuePosition(*comment.Body); position > 0 {
		pr.Position = position
	}
}

func parseBotStatus(status string) string {
	parts := strings.Split(status, " ")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToUpper(strings.TrimSuffix(strings.TrimSpace(parts[1]), ":"))
}

func parseJenkinsStatus(commentBody string) string {
	if !strings.Contains(commentBody, "openshift-jenkins/test") {
		return ""
	}
	return parseBotStatus(commentBody)
}

func parseJenkinsURL(commentBody string) string {
	if !strings.Contains(commentBody, "continuous-integration/openshift-jenkins") {
		return ""
	}
	if parts := strings.Split(commentBody, "("); len(parts) != 2 {
		return ""
	} else {
		return strings.TrimSuffix(parts[1], ")")
	}
}

func parseMergeStatus(commentBody string) string {
	if !strings.Contains(commentBody, "openshift-jenkins/merge") {
		return ""
	}
	return parseBotStatus(commentBody)
}

func parseMergeQueuePosition(commentBody string) int {
	if !strings.Contains(commentBody, "You are in the build queue at position:") {
		return 0
	}
	if parts := strings.Split(commentBody, ":"); len(parts) < 3 {
		return 0
	} else {
		return api.StringToInt(parts[2])
	}
}
