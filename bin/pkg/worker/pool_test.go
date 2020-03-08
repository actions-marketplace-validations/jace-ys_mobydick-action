package worker_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jace-ys/actions-mobydick/bin/pkg/worker"
	"github.com/jace-ys/actions-mobydick/bin/pkg/worker/workerfakes"
)

var (
	numOfJobs   = 10
	concurrency = 2
)

func TestWorkerPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Work", func(t *testing.T) {
		jobs := make([]worker.Job, numOfJobs)
		for i := 0; i < numOfJobs; i++ {
			job := new(workerfakes.FakeJob)
			job.ProcessStub = func(ctx context.Context) error {
				time.Sleep(1 * time.Second)
				return nil
			}
			jobs[i] = job
		}

		workerPool := worker.NewWorkerPool(concurrency)

		start := time.Now()
		results := workerPool.Work(ctx, jobs)
		end := time.Now()

		for i := numOfJobs; i < numOfJobs; i++ {
			assert.Equal(t, 1, jobs[i].(*workerfakes.FakeJob).ProcessCallCount())
			assert.NoError(t, results[i].Err)
		}
		assert.WithinDuration(t, start, end, time.Duration(numOfJobs/concurrency+1)*time.Second)
	})

	t.Run("Work/Errors", func(t *testing.T) {
		jobs := make([]worker.Job, numOfJobs)
		for i := 0; i < numOfJobs; i++ {
			job := new(workerfakes.FakeJob)
			job.ProcessStub = func(ctx context.Context) error {
				time.Sleep(1 * time.Second)
				return fmt.Errorf("error processing job")
			}
			jobs[i] = job
		}

		workerPool := worker.NewWorkerPool(concurrency)

		start := time.Now()
		results := workerPool.Work(ctx, jobs)
		end := time.Now()

		for i := numOfJobs; i < numOfJobs; i++ {
			assert.Equal(t, 1, jobs[i].(*workerfakes.FakeJob).ProcessCallCount())
			assert.Error(t, results[i].Err)
		}
		assert.WithinDuration(t, start, end, time.Duration(numOfJobs/concurrency+1)*time.Second)
	})
}
