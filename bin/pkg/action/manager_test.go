package action_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/go-github/v29/github"
	"github.com/stretchr/testify/assert"

	"github.com/jace-ys/mobydick-action/bin/pkg/action"
	"github.com/jace-ys/mobydick-action/bin/pkg/action/actionfakes"
	"github.com/jace-ys/mobydick-action/bin/pkg/worker"
)

func TestActionManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewNopLogger()

	t.Run("ListRepositories", func(t *testing.T) {
		workerPool := &worker.WorkerPool{}
		workflowFile := &action.WorkflowFile{}

		t.Run("Error", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.ListByOrgReturnsOnCall(0, fakeRepositories(0), &github.Response{NextPage: 0}, fmt.Errorf("could not list repositories"))

			actionManager := action.NewActionManager(ctx, logger, "organisation", false, workflowFile, workerPool, repositoriesService)
			repositories, err := actionManager.ListRepositories(ctx, true)

			assert.Equal(t, 1, repositoriesService.ListByOrgCallCount())
			assert.Error(t, err)
			assert.Equal(t, 0, len(repositories))
		})

		t.Run("PagesOne", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.ListByOrgReturnsOnCall(0, fakeRepositories(2), &github.Response{NextPage: 0}, nil)

			actionManager := action.NewActionManager(ctx, logger, "organisation", false, workflowFile, workerPool, repositoriesService)
			repositories, err := actionManager.ListRepositories(ctx, true)

			assert.Equal(t, 1, repositoriesService.ListByOrgCallCount())
			assert.NoError(t, err)
			assert.Equal(t, 2, len(repositories))
		})

		t.Run("PagesMultiple", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.ListByOrgReturnsOnCall(0, fakeRepositories(2), &github.Response{NextPage: 1}, nil)
			repositoriesService.ListByOrgReturnsOnCall(1, fakeRepositories(2), &github.Response{NextPage: 0}, nil)

			actionManager := action.NewActionManager(ctx, logger, "organisation", false, workflowFile, workerPool, repositoriesService)
			repositories, err := actionManager.ListRepositories(ctx, true)

			assert.Equal(t, 2, repositoriesService.ListByOrgCallCount())
			assert.NoError(t, err)
			assert.Equal(t, 4, len(repositories))
		})
	})

	t.Run("CreateFile", func(t *testing.T) {
		workflowFile := &action.WorkflowFile{
			Path:    "path/to/workflow.yaml",
			Content: []byte("content"),
		}

		t.Run("Error", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.CreateFileReturnsOnCall(0, &github.RepositoryContentResponse{}, &github.Response{}, fmt.Errorf("could not create file"))

			workerPool := worker.NewWorkerPool(1)

			actionManager := action.NewActionManager(ctx, logger, "organisation", false, workflowFile, workerPool, repositoriesService)
			err := actionManager.CreateFile(ctx, "repository", workflowFile.Path, workflowFile.Content)

			assert.Equal(t, 1, repositoriesService.CreateFileCallCount())
			assert.Error(t, err)
		})

		t.Run("DryRun", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)

			workerPool := worker.NewWorkerPool(1)

			actionManager := action.NewActionManager(ctx, logger, "organisation", true, workflowFile, workerPool, repositoriesService)
			err := actionManager.CreateFile(ctx, "repository", workflowFile.Path, workflowFile.Content)

			assert.Equal(t, 0, repositoriesService.CreateFileCallCount())
			assert.NoError(t, err)
		})

		t.Run("Success", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.CreateFileReturnsOnCall(0, &github.RepositoryContentResponse{}, &github.Response{}, nil)

			workerPool := worker.NewWorkerPool(1)

			actionManager := action.NewActionManager(ctx, logger, "organisation", false, workflowFile, workerPool, repositoriesService)
			err := actionManager.CreateFile(ctx, "repository", workflowFile.Path, workflowFile.Content)

			assert.Equal(t, 1, repositoriesService.CreateFileCallCount())
			assert.NoError(t, err)
		})
	})

	t.Run("DistributeCommand", func(t *testing.T) {
		workflowFile := &action.WorkflowFile{
			Path:    "path/to/workflow.yaml",
			Content: []byte("content"),
		}

		t.Run("Failure", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.ListByOrgReturnsOnCall(0, fakeRepositories(1), &github.Response{NextPage: 0}, nil)
			repositoriesService.CreateFileReturnsOnCall(0, &github.RepositoryContentResponse{}, &github.Response{}, fmt.Errorf("failed to create file"))

			workerPool := worker.NewWorkerPool(1)

			actionManager := action.NewActionManager(ctx, logger, "organisation", false, workflowFile, workerPool, repositoriesService)
			success, failures, err := actionManager.Distribute(ctx, true)

			assert.Equal(t, 1, repositoriesService.ListByOrgCallCount())
			assert.Equal(t, 1, repositoriesService.CreateFileCallCount())
			assert.NoError(t, err)
			assert.Equal(t, 0, success)
			assert.Equal(t, 1, failures)
		})

		t.Run("Success", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.ListByOrgReturnsOnCall(0, fakeRepositories(1), &github.Response{NextPage: 0}, nil)
			repositoriesService.CreateFileReturnsOnCall(0, &github.RepositoryContentResponse{}, &github.Response{}, nil)

			workerPool := worker.NewWorkerPool(1)

			actionManager := action.NewActionManager(ctx, logger, "organisation", false, workflowFile, workerPool, repositoriesService)
			success, failures, err := actionManager.Distribute(ctx, true)

			assert.Equal(t, 1, repositoriesService.ListByOrgCallCount())
			assert.Equal(t, 1, repositoriesService.CreateFileCallCount())
			assert.NoError(t, err)
			assert.Equal(t, 1, success)
			assert.Equal(t, 0, failures)
		})
	})
}

func fakeRepositories(num int) []*github.Repository {
	var repositories []*github.Repository
	name := "repository"
	for i := 0; i < num; i++ {
		repositories = append(repositories, &github.Repository{Name: &name})
	}
	return repositories
}
