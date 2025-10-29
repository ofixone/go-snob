package gitea

import (
	"fmt"
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
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		})
}

// AddComment добавляет комментарий к pull request / issue
// owner - название проекта, где хранится репозиторий
// repo - название репозитория
// issueID - идентификатор pull request / issue
func (c *Client) AddComment(owner string, repo string, issueID int) (*resty.Response, error) {
	return c.newRequest().
		SetBody(map[string]any{
			"body": "Тестовый комментарий",
		}).
		SetPathParams(map[string]string{"owner": owner, "repo": repo, "index": strconv.Itoa(issueID)}).
		Post(c.baseUrl + "/repos/{owner}/{repo}/issues/{index}/comments")
}
