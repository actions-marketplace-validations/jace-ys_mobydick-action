package action_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/go-github/v29/github"
	"github.com/stretchr/testify/assert"

	"github.com/jace-ys/actions-mobydick/bin/pkg/action"
	"github.com/jace-ys/actions-mobydick/bin/pkg/action/actionfakes"
	"github.com/jace-ys/actions-mobydick/bin/pkg/worker"
	"github.com/jace-ys/actions-mobydick/bin/pkg/workflow"
)

func TestActionManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewNopLogger()

	workerPool := worker.NewWorkerPool(1)

	t.Run("ListRepositories", func(t *testing.T) {
		var workflowFile *workflow.WorkflowFile

		t.Run("Error", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.ListByOrgReturnsOnCall(0, fakeRepositories(0), &github.Response{NextPage: 0}, fmt.Errorf("could not list repositories"))

			actionManager := action.NewActionManager(ctx, logger, "organisation", workflowFile, workerPool, repositoriesService)
			repositories, err := actionManager.ListRepositories(ctx, true)

			assert.Equal(t, 1, repositoriesService.ListByOrgCallCount())
			assert.Error(t, err)
			assert.Equal(t, 0, len(repositories))
		})

		t.Run("PagesOne", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.ListByOrgReturnsOnCall(0, fakeRepositories(2), &github.Response{NextPage: 0}, nil)

			actionManager := action.NewActionManager(ctx, logger, "organisation", workflowFile, workerPool, repositoriesService)
			repositories, err := actionManager.ListRepositories(ctx, true)

			assert.Equal(t, 1, repositoriesService.ListByOrgCallCount())
			assert.NoError(t, err)
			assert.Equal(t, 2, len(repositories))
		})

		t.Run("PagesMultiple", func(t *testing.T) {
			repositoriesService := new(actionfakes.FakeRepositoriesService)
			repositoriesService.ListByOrgReturnsOnCall(0, fakeRepositories(2), &github.Response{NextPage: 1}, nil)
			repositoriesService.ListByOrgReturnsOnCall(1, fakeRepositories(2), &github.Response{NextPage: 0}, nil)

			actionManager := action.NewActionManager(ctx, logger, "organisation", workflowFile, workerPool, repositoriesService)
			repositories, err := actionManager.ListRepositories(ctx, true)

			assert.Equal(t, 2, repositoriesService.ListByOrgCallCount())
			assert.NoError(t, err)
			assert.Equal(t, 4, len(repositories))
		})
	})
}

func fakeRepositories(num int) []*github.Repository {
	return make([]*github.Repository, num)
}
