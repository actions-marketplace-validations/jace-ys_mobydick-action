//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
package action

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v29/github"
)

//counterfeiter:generate . RepositoriesService
type RepositoriesService interface {
	ListByOrg(ctx context.Context, org string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
	CreateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error)
}

type actionManager struct {
	organisation        string
	logger              log.Logger
	workflowFile        []byte
	repositoriesService RepositoriesService
}

func NewActionManager(ctx context.Context, organisation string, logger log.Logger, workflowFile []byte, repositories RepositoriesService) *actionManager {
	return &actionManager{
		organisation:        organisation,
		logger:              logger,
		workflowFile:        workflowFile,
		repositoriesService: repositories,
	}
}

func (am *actionManager) DistributeCommand(ctx context.Context, concurrency int, private bool) error {
	repositories, err := am.ListRepositories(ctx, private)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	for i, repository := range repositories {
		level.Info(am.logger).Log("count", i+1, "repository", *repository.FullName)
	}

	return nil
}

func (am *actionManager) ListRepositories(ctx context.Context, private bool) ([]*github.Repository, error) {
	filter := "all"
	if private {
		filter = "private"
	}

	opts := &github.RepositoryListByOrgOptions{
		Type:        filter,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var list []*github.Repository
	for {
		repositories, response, err := am.repositoriesService.ListByOrg(ctx, am.organisation, opts)
		if err != nil {
			return nil, err
		}
		list = append(list, repositories...)
		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return list, nil
}

func (am *actionManager) CreateFile(ctx context.Context, repository, path string, content []byte) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.String("This is my commit message"),
		Content: content,
	}

	_, _, err := am.repositoriesService.CreateFile(ctx, am.organisation, repository, path, opts)
	if err != nil {
		return err
	}

	return nil
}
