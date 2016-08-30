package client

import (
	"log"

	gh "github.com/google/go-github/github"
)

type Github struct {
	*gh.Client
}

func (g *Github) Events() []*gh.Event {
	result, _, err := g.Client.Activity.ListRepositoryEvents("openshift", "origin", &gh.ListOptions{
		Page:    1,
		PerPage: 100,
	})
	if err != nil {
		log.Printf("ERROR: unable to list github events: %v", err)
		return []*gh.Event{}
	}
	return result
}

func (g *Github) Comment(id int) *gh.IssueComment {
	result, _, err := g.Client.Issues.GetComment("openshift", "origin", id)
	if err != nil {
		log.Printf("ERROR: unable to get github comment: %v", err)
		return nil
	}
	return result
}
