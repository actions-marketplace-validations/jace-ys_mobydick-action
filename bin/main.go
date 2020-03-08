package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jace-ys/actions-mobydick/bin/pkg/action"
)

var (
	actionCmd    = kingpin.New("action", "Command-line interface to manage this GitHub Action.")
	organisation = actionCmd.Flag("organisation", "Name of organisation in GitHub.").Required().String()
	token        = actionCmd.Flag("token", "Token used for authenticating with GitHub.").Required().String()

	distributeCmd = actionCmd.Command("distribute", "Distribute this GitHub Action to all repositories in the organisation.")
	concurrency   = distributeCmd.Flag("concurrency", "Size of worker pool to perform concurrent work.").Default("5").Int()
	file          = distributeCmd.Flag("file", "Workflow file to commit into repositories.").Default("mobydick.yaml").String()
	private       = distributeCmd.Flag("private", "Only distribute GitHub Action to private repositories (default: false).").Default("false").Bool()
)

func main() {
	command := kingpin.MustParse(actionCmd.Parse(os.Args[1:]))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "caller", log.DefaultCaller)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: *token,
		},
	)
	githubClient := github.NewClient(oauth2.NewClient(ctx, ts))

	workflowFile, err := ioutil.ReadFile(*file)
	if err != nil {
		level.Error(logger).Log("error", err)
		os.Exit(1)
	}

	actionManager := action.NewActionManager(ctx, *organisation, logger, workflowFile, githubClient.Repositories)

	switch command {
	case distributeCmd.FullCommand():
		err := actionManager.DistributeCommand(ctx, *concurrency, *private)
		if err != nil {
			level.Error(logger).Log("error", err)
			os.Exit(1)
		}
	}
}
