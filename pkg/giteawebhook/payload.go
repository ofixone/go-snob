package giteawebhook

import "fmt"

type Action string

const (
	ActionReviewRequest        Action = "review_requested"
	ActionReviewRequestRemoved Action = "review_request_removed"
)

var validActions = map[Action]struct{}{
	ActionReviewRequest:        {},
	ActionReviewRequestRemoved: {},
}

func (a Action) Validate() error {
	if _, ok := validActions[a]; !ok {
		return fmt.Errorf("invalid Action: %q", a)
	}
	return nil
}

type PullRequest struct {
	ID      int    `json:"id"`
	DiffURL string `json:"diff_url"`
	Head    struct {
		SHA string `json:"sha"`
	} `json:"head"`
}

type Owner struct {
	Login string `json:"login"`
}

type Repository struct {
	Name  string `json:"name"`
	Owner Owner  `json:"owner"`
}

type Payload struct {
	Action      Action      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
}
