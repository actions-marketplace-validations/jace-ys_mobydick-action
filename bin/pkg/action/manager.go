//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
package action

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v29/github"

	"github.com/jace-ys/actions-mobydick/bin/pkg/worker"
	"github.com/jace-ys/actions-mobydick/bin/pkg/workflow"
)

//counterfeiter:generate . RepositoriesService
type RepositoriesService interface {
	ListByOrg(ctx context.Context, org string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
	CreateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error)
}

type ActionManager struct {
	logger              log.Logger
	organisation        string
	dryRun              bool
	workflowFile        *workflow.WorkflowFile
	workerPool          *worker.WorkerPool
	repositoriesService RepositoriesService
}

func NewActionManager(
	ctx context.Context,
	logger log.Logger,
	organisation string,
	dryRun bool,
	workflowFile *workflow.WorkflowFile,
	workerPool *worker.WorkerPool,
	repositories RepositoriesService,
) *ActionManager {
	return &ActionManager{
		logger:              logger,
		organisation:        organisation,
		dryRun:              dryRun,
		workflowFile:        workflowFile,
		workerPool:          workerPool,
		repositoriesService: repositories,
	}
}

func (am *ActionManager) DistributeCommand(ctx context.Context, private bool) (int, int, error) {
	repositories, err := am.ListRepositories(ctx, private)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list repositories: %w", err)
	}

	var jobs []worker.Job
	for _, repository := range repositories {
		jobs = append(jobs, &createFileJob{
			handler:    am,
			repository: *repository.Name,
			path:       am.workflowFile.Path,
			content:    am.workflowFile.Content,
		})
	}

	results := am.workerPool.Work(ctx, jobs)

	var success []worker.Result
	var failures []worker.Result
	for _, result := range results {
		if result.Err != nil {
			failures = append(failures, result)
		} else {
			success = append(success, result)
		}
	}

	return len(success), len(failures), nil
}

func (am *ActionManager) ListRepositories(ctx context.Context, private bool) ([]*github.Repository, error) {
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

func (am *ActionManager) CreateFile(ctx context.Context, repository, path string, content []byte) error {
	if am.dryRun {
		level.Info(am.logger).Log("event", "create_file.dry_run", "repository", repository)
		return nil
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String("GitHub Actions workflow for Mobydick"),
		Content: content,
	}

	_, _, err := am.repositoriesService.CreateFile(ctx, am.organisation, repository, path, opts)
	if err != nil {
		level.Info(am.logger).Log("event", "create_file.failure", "repository", repository, "error", err)
		return err
	}

	level.Info(am.logger).Log("event", "create_file.success", "repository", repository)
	return nil
}

type createFileJob struct {
	handler    *ActionManager
	repository string
	path       string
	content    []byte
}

func (job *createFileJob) Process(ctx context.Context) error {
	err := job.handler.CreateFile(ctx, job.repository, job.path, job.content)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	return nil
}
