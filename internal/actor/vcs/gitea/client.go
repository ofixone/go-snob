package gitea

import (
	"fmt"
	"go-snob/internal/actor/ai"
	"strconv"
	"time"

	"resty.dev/v3"
)

const baseTimeout = 10 * time.Second

type Client struct {
	baseUrl string
	token   string
	client  *resty.Client
}

func NewClient(client *resty.Client, baseUrl string, token string) *Client {
	return &Client{token: token, baseUrl: baseUrl, client: client}
}

func (c *Client) newRequest() *resty.Request {
	return c.client.R().SetTimeout(baseTimeout).
		SetHeader("Authorization", fmt.Sprintf("token %s", c.token)).
		SetHeaders(
			map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
		)
}

func (c *Client) AddComment(owner string, repo string, issueID int) (*resty.Response, error) {
	return c.newRequest().
		SetBody(
			map[string]any{
				"body": "Тестовый комментарий",
			},
		).
		SetPathParams(map[string]string{"owner": owner, "repo": repo, "index": strconv.Itoa(issueID)}).
		Post(c.baseUrl + "/repos/{owner}/{repo}/issues/{index}/comments")
}

func (c *Client) CreateReview(
	owner string,
	repo string,
	issueID int,
	commitSHA string,
	review ai.AIReviewResult,
) (*resty.Response, error) {
	var comments []map[string]any
	for _, com := range review.Comments {
		comments = append(
			comments, map[string]any{
				"body":         com.Message,
				"path":         com.File,
				"new_position": com.NewPosition,
				"old_position": com.OldPosition,
			},
		)
	}

	return c.newRequest().
		SetBody(
			map[string]any{
				"body":      review.Summary,
				"event":     review.Verdict,
				"commit_id": commitSHA,
				"comments":  comments,
			},
		).
		SetPathParams(
			map[string]string{
				"owner": owner,
				"repo":  repo,
				"index": strconv.Itoa(issueID),
			},
		).
		Post(c.baseUrl + "/repos/{owner}/{repo}/pulls/{index}/reviews")
}

func (c *Client) GetDiff(owner string, repo string, issueID int) (*resty.Response, error) {
	return c.newRequest().
		SetPathParams(
			map[string]string{
				"owner": owner, "repo": repo, "index": strconv.Itoa(issueID), "diffType": "diff",
			},
		).Get(c.baseUrl + "/repos/{owner}/{repo}/pulls/{index}.diff")
}
