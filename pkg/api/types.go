package api

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

type Event struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type PullRequest struct {
	Number            int        `json:"num"`
	Title             string     `json:"title"`
	Author            string     `json:"author"`
	Position          int        `json:"pos"`
	JenkinsTestStatus string     `json:"jenkinsTestStatus"`
	Merge             bool       `json:"merge"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt"`
}

func (p *PullRequest) ToJSON() string {
	result, _ := json.Marshal(p)
	return string(result)
}

func (p *PullRequest) FromJSON(in string) *PullRequest {
	json.Unmarshal([]byte(in), p)
	return p
}

func (p *PullRequest) IsFailure() bool {
	if len(p.JenkinsTestStatus) == 0 {
		return false
	}
	return strings.Contains("FAILURE", p.JenkinsTestStatus)
}

func (p *PullRequest) Equal(pull *PullRequest) bool {
	if pull == nil {
		return false
	}
	copyOld := *p
	copyOld.UpdatedAt = nil
	copyNew := *pull
	copyNew.UpdatedAt = nil
	return reflect.DeepEqual(copyOld, copyNew)
}

func (p *PullRequest) IsLatest(pull *PullRequest) bool {
	if pull == nil {
		return true
	}
	return pull.CreatedAt.Before(p.CreatedAt)
}