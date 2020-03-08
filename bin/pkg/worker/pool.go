//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
package worker

import (
	"context"
	"sync"
)

//counterfeiter:generate . Job
type Job interface {
	Process(ctx context.Context) error
}

type Result struct {
	Job Job
	Err error
}

type WorkerPool struct {
	concurrency int
	jobsChan    chan Job
	resultsChan chan Result
	waitGroup   sync.WaitGroup
}

func NewWorkerPool(concurrency int) *WorkerPool {
	return &WorkerPool{
		concurrency: concurrency,
		jobsChan:    make(chan Job, concurrency),
	}
}

func (p *WorkerPool) Work(ctx context.Context, jobs []Job) []Result {
	p.resultsChan = make(chan Result, len(jobs))

	p.waitGroup.Add(p.concurrency)
	for i := 0; i < p.concurrency; i++ {
		go func() {
			defer p.waitGroup.Done()
			p.startWorker(ctx)
		}()
	}

	go func() {
		defer close(p.jobsChan)
		for _, job := range jobs {
			p.jobsChan <- job
		}
	}()

	go func() {
		defer close(p.resultsChan)
		p.waitGroup.Wait()
	}()

	return p.Results()
}

func (p *WorkerPool) Results() []Result {
	var results []Result
	for r := range p.resultsChan {
		results = append(results, r)
	}
	return results
}

func (p *WorkerPool) startWorker(ctx context.Context) {
	for job := range p.jobsChan {
		err := job.Process(ctx)
		p.resultsChan <- Result{job, err}
	}
}
