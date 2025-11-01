package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"resty.dev/v3"
)

const baseTimeout = 30 * time.Second

type Client struct {
	client *resty.Client
	logger *zap.Logger

	baseUrl string
	token   string

	systemPrompt string
}

func NewClient(client *resty.Client, logger *zap.Logger, baseUrl string, token string, systemPrompt string) *Client {
	return &Client{
		client:       client,
		logger:       logger,
		token:        token,
		baseUrl:      baseUrl,
		systemPrompt: systemPrompt,
	}
}

func (c *Client) newRequest() *resty.Request {
	return c.client.R().SetTimeout(baseTimeout).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.token)).
		SetHeaders(
			map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
		)
}

type AIReviewResult struct {
	// Общая оценка
	Verdict  string    `json:"verdict,omitempty"` // "APPROVED", "REQUEST_CHANGES", "COMMENT"
	Summary  string    `json:"summary"`
	Comments []Comment `json:"comments"`
}

type ModelResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"` // <- здесь JSON как строка
			// Остальные поля можно пропустить, если не нужны
		} `json:"message"`
	} `json:"choices"`
}

type Comment struct {
	File        string `json:"file"`
	NewPosition int    `json:"new_position"`
	OldPosition int    `json:"old_position"`
	Message     string `json:"message"`
}

func (c *Client) Send(message string) (AIReviewResult, error) {
	r, err := c.newRequest().
		SetBody(
			map[string]any{
				"model":             "Qwen/Qwen3-Coder-480B-A35B-Instruct",
				"max_tokens":        2000,
				"temperature":       .1,
				"top_p":             .8,
				"frequency_penalty": .5,
				"presence_penalty":  0,
				"messages": []map[string]string{
					{
						"role":    "system",
						"content": c.systemPrompt,
					},
					{
						"role":    "user",
						"content": message,
					},
				},
				"response_format": map[string]any{
					"type": "json_schema",
					"json_schema": map[string]any{
						"name":   "ai_review_result",
						"strict": true,
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"verdict": map[string]any{
									"type": "string",
									"enum": []string{"APPROVED", "REQUEST_CHANGES", "COMMENT"},
									"description": "Ставишь APPROVED если в целом все ок, нет критичных проблем. Ставишь" +
										" REQUEST CHANGES если критичные проблемы есть. Ставишь COMMENT, если не определился ",
								},
								"summary": map[string]any{
									"type": "string",
								},
								"comments": map[string]any{
									"type": "array",
									"items": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"file": map[string]any{
												"type": "string",
											},
											"new_position": map[string]any{
												"type":        "integer",
												"description": "номер строки в diff после изменений, если комментируем новую строку",
											},
											"old_position": map[string]any{
												"type":        "integer",
												"description": "номер строки в diff до изменений, если комментируем удалённую строку",
											},
											"message": map[string]any{
												"type": "string",
											},
										},
										"required": []string{"file", "new_position", "old_position", "message"},
									},
								},
							},
							"required": []string{"summary", "comments", "verdict"},
						},
					},
				},
			},
		).
		Post(c.baseUrl)
	if err != nil {
		return AIReviewResult{}, fmt.Errorf("send request: %w", err)
	}
	if r.StatusCode() != http.StatusOK {
		return AIReviewResult{}, fmt.Errorf(
			"send request: %w",
			fmt.Errorf("unexpected status code: %v", r.StatusCode()),
		)
	}

	var resp ModelResponse
	if err := json.Unmarshal(r.Bytes(), &resp); err != nil {
		return AIReviewResult{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(resp.Choices) > 0 {
		contentStr := resp.Choices[0].Message.Content

		var review AIReviewResult
		if err := json.Unmarshal([]byte(contentStr), &review); err != nil {
			return AIReviewResult{}, fmt.Errorf("unmarshal response: %w", err)
		}
		return review, nil
	}

	return AIReviewResult{}, fmt.Errorf("empty response")
}
