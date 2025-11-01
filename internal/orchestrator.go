package internal

import (
	"context"
	"go-snob/internal/actor/ai"
	"go-snob/internal/actor/vcs/gitea"
	"go-snob/pkg/giteawebhook"

	"go.uber.org/zap"
)

type Orchestrator struct {
	aiClient    *ai.Client
	giteaClient *gitea.Client
	logger      *zap.Logger
}

func NewOrchestrator(ai *ai.Client, gitea *gitea.Client, logger *zap.Logger) *Orchestrator {
	return &Orchestrator{aiClient: ai, giteaClient: gitea, logger: logger}
}

func (o *Orchestrator) Handler(ctx context.Context, p giteawebhook.Payload) {
	if p.Action != giteawebhook.ActionReviewRequest {
		return
	}

	o.logger.Info("starting getting diff..")
	diff, err := o.giteaClient.GetDiff(p.Repository.Owner.Login, p.Repository.Name, p.PullRequest.ID)
	if err != nil {
		o.logger.Error("failed to get diff", zap.Error(err))
		return
	}
	o.logger.Info("got diff")

	o.logger.Info("starting ai request..")
	aiReview, err := o.aiClient.Send(diff.String())
	if err != nil {
		o.logger.Error("failed to send ai review", zap.Error(err))
		return
	}
	o.logger.Info("got ai review")

	o.logger.Info("starting gitea review..")
	giteReview, err := o.giteaClient.CreateReview(
		p.Repository.Owner.Login,
		p.Repository.Name,
		p.PullRequest.ID,
		p.PullRequest.Head.SHA,
		aiReview,
	)
	if err != nil {
		o.logger.Error("failed to create review", zap.Error(err))
	}
	_ = giteReview
	o.logger.Info("got gitea review")
}
