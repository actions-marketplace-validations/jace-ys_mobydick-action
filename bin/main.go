package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	action       = kingpin.New("action", "Command-line interface to manage this GitHub Action.")
	organisation = action.Flag("organisation", "Name of organisation in GitHub.").Required().String()
	token        = action.Flag("token", "Token used for authenticating with GitHub.").Required().String()

	distribute = action.Command("distribute", "Distribute this GitHub Action to all repositories in the organisation.")
	private    = distribute.Flag("private", "Only distribute GitHub Action to private repositories (default: false).").Default("false").Bool()
)

func main() {
	command := kingpin.MustParse(action.Parse(os.Args[1:]))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actionManager := NewActionManager(ctx, *organisation, *token)

	switch command {
	case distribute.FullCommand():
		err := actionManager.DistributeCommand(ctx, *private)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type actionManager struct {
	githubClient *github.Client
	organisation string
}

func NewActionManager(ctx context.Context, organisation, token string) *actionManager {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	githubClient := github.NewClient(oauth2.NewClient(ctx, ts))

	return &actionManager{
		githubClient: githubClient,
		organisation: organisation,
	}
}

func (am *actionManager) DistributeCommand(ctx context.Context, private bool) error {
	repositories, err := am.ListRepositories(ctx, private)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	for i, repository := range repositories {
		log.Printf("%d: %s\n", i+1, *repository.FullName)
	}

	return nil
}

func (am *actionManager) ListRepositories(ctx context.Context, private bool) ([]*github.Repository, error) {
	filter := "all"
	if private {
		filter = "private"
	}

	options := &github.RepositoryListByOrgOptions{
		Type:        filter,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var list []*github.Repository
	for {
		repositories, response, err := am.githubClient.Repositories.ListByOrg(ctx, am.organisation, options)
		if err != nil {
			return nil, err
		}
		list = append(list, repositories...)
		if response.NextPage == 0 {
			break
		}
		options.Page = response.NextPage
	}

	return list, nil
}
