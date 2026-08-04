package vek

import (
	"github.com/urfave/cli/v3"
	gitignore "github.com/vekio/vek/internal/cli/gitignore"
	sshalias "github.com/vekio/vek/internal/cli/ssh-alias"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:                  "vek",
		Version:               "v0.0.1",
		Usage:                 "vekio power user cli",
		EnableShellCompletion: true,
		Commands:              registerCommands(),
	}
}

func registerCommands() []*cli.Command {
	return []*cli.Command{
		gitignore.NewCmd(),
		sshalias.NewCmd(),
	}
}
