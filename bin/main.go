package main

import (
	"context"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jace-ys/mobydick-action/bin/pkg/action"
	"github.com/jace-ys/mobydick-action/bin/pkg/worker"
)

var (
	actionCmd    = kingpin.New("action", "Command-line interface for managing this GitHub Action.")
	organisation = actionCmd.Flag("organisation", "Name of organisation in GitHub.").Required().String()
	token        = actionCmd.Flag("token", "Token used for authenticating with GitHub.").Required().String()

	distributeCmd = actionCmd.Command("distribute", "Distribute this GitHub Action to all repositories in the organisation.")
	concurrency   = distributeCmd.Flag("concurrency", "Size of worker pool to perform concurrent work.").Default("5").Int()
	file          = distributeCmd.Flag("file", "Workflow file to commit into repositories.").Default("mobydick.yaml").String()
	version       = distributeCmd.Flag("version", "Version of this GitHub Action to distribute.").Default("v1.0.0").String()
	private       = distributeCmd.Flag("private", "Only distribute this GitHub Action to private repositories.").Default("false").Bool()
	dryRun        = distributeCmd.Flag("dry-run", "Perform a dry run, showing all the repositories that will be committed to.").Default("false").Bool()
)

func main() {
	command := kingpin.MustParse(actionCmd.Parse(os.Args[1:]))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "caller", log.DefaultCaller)

	workflowFile, err := action.NewWorkflowFile(*file, *version)
	if err != nil {
		level.Error(logger).Log("error", err)
		os.Exit(1)
	}

	workerPool := worker.NewWorkerPool(*concurrency)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: *token,
		},
	)
	githubClient := github.NewClient(oauth2.NewClient(ctx, ts))

	actionManager := action.NewActionManager(ctx, logger, *organisation, *dryRun, workflowFile, workerPool, githubClient.Repositories)

	switch command {
	case distributeCmd.FullCommand():
		success, failures, err := actionManager.Distribute(ctx, *private)
		if err != nil {
			level.Error(logger).Log("error", err)
			os.Exit(1)
		}
		level.Info(logger).Log("success", success, "failures", failures)
	}
}
