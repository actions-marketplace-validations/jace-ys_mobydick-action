[![ci-badge]][ci-workflow]

[ci-badge]: https://github.com/jace-ys/mobydick-action/workflows/.github/workflows/ci.yml/badge.svg
[ci-workflow]: https://github.com/jace-ys/mobydick-action/actions?query=workflow%3A.github%2Fworkflows%2Fci.yml

# Mobydick Action

GitHub Action to validate that Docker images are compatible with [Dependabot's](https://dependabot.com/) update strategy.

## About

Mobydick Action is a Rust binary packaged as a Docker container action. It verifies that any Dockerfiles in a repository uses images that are compatible with [Dependabot's update strategy for Dockerfiles](https://dependabot.com/blog/dependabot-now-supports-docker/).

This will enable Dependabot to do its job of creating pull requests to bump images whenever a new version is released, which helps minimise the risks that outdated images pose and eliminates the dilemma faced when deciding on what version of an image to use.

## CLI

Mobydick Action comes bundled with a command-line tool built in Go for managing this GitHub Action across your organisation. You will need Go 1.14 to build the CLI binary.

1. Clone the repository and build the CLI binary:

```
make --directory bin
```

2. Using the CLI:

```
$ bin/action --help
usage: action --organisation=ORGANISATION --token=TOKEN [<flags>] <command> [<args> ...]

Command-line interface for managing this GitHub Action.

Flags:
  --help                       Show context-sensitive help (also try --help-long and --help-man).
  --organisation=ORGANISATION  Name of organisation in GitHub.
  --token=TOKEN                Token used for authenticating with GitHub.

Commands:
  help [<command>...]
    Show help.

  distribute [<flags>]
    Distribute this GitHub Action to all repositories in the organisation.
```

- `bin/action distribute`:

  Used to distribute Mobydick Action to all repositories in a GitHub organisation as a workflow file in the `.github/workflows` folder (currently commits directly to the default branch). See `bin/action distribute --help` for more info. Configure `bin/mobydick.yaml` for your own use cases.
