package client

import (
	"fmt"
	"log"
	"os"

	gh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Github struct {
	*gh.Client
}

func NewGithub() (*Github, error) {
	token := os.Getenv("GITHUB_API_KEY")
	if len(token) == 0 {
		return nil, fmt.Errorf("you must set GITHUB_API_KEY")
	}
	return &Github{
		gh.NewClient(
			oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})),
		),
	}, nil
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
