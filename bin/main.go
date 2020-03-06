package main

import (
	"context"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jace-ys/actions-mobydick/bin/action"
)

var (
	actionCmd    = kingpin.New("action", "Command-line interface to manage this GitHub Action.")
	organisation = actionCmd.Flag("organisation", "Name of organisation in GitHub.").Required().String()
	token        = actionCmd.Flag("token", "Token used for authenticating with GitHub.").Required().String()

	distributeCmd = actionCmd.Command("distribute", "Distribute this GitHub Action to all repositories in the organisation.")
	concurrency   = distributeCmd.Flag("concurrency", "Size of worker pool to perform concurrent work.").Default("5").Int()
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

	actionManager := action.NewActionManager(ctx, *organisation, logger, githubClient.Repositories)

	switch command {
	case distributeCmd.FullCommand():
		err := actionManager.DistributeCommand(ctx, *concurrency, *private)
		if err != nil {
			level.Error(logger).Log("error", err)
			os.Exit(1)
		}
	}
}
