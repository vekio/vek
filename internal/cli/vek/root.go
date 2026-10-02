package vek

import (
	"github.com/urfave/cli/v3"
	"github.com/vekio/vek/internal/cli/bonsai"
	gitignore "github.com/vekio/vek/internal/cli/gitignore"
	"github.com/vekio/vek/internal/cli/sshalias"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:                  "vek",
		Version:               buildVersion(),
		Usage:                 "vekio power user cli",
		EnableShellCompletion: true,
		Commands:              registerCommands(),
	}
}

func registerCommands() []*cli.Command {
	return []*cli.Command{
		gitignore.NewCmd(),
		sshalias.NewCmd(),
		bonsai.NewCmd(),
	}
}
