package bonsai

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
	"github.com/vekio/vek/internal/git"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "bonsai",
		Usage: "manage Git branches and worktrees for development tasks",
		Commands: []*cli.Command{
			newInitCmd(),
			newCloneCmd(),
			newStartCmd(),
			newListCmd(),
			newStatusCmd(),
			newSubmitCmd(),
			newCleanCmd(),
		},
	}
}

func newGitClient(c *cli.Command) *git.Client {
	stdout := c.Root().Writer
	if c.Bool("print-path") {
		stdout = io.Discard
	}
	return git.New(stdout, c.Root().ErrWriter)
}

// Resume la rama y el estado del worktree actual.
// bonsai status
func newStatusCmd() *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "show the current worktree, branch, changes, and divergence from origin/main",
		Arguments: []cli.Argument{},
		Flags:     []cli.Flag{},
		Action: func(ctx context.Context, c *cli.Command) error {
			// Ejecutar desde un worktree; --show-toplevel falla en el bare.
			// git rev-parse --show-toplevel
			// git branch --show-current
			// git status --porcelain=v1
			// git rev-parse --path-format=absolute --git-common-dir
			// El padre del .git común da el nombre del repositorio.
			// Comprobar origin/main; si falta, mostrar el estado sin divergencia.
			// git rev-list --left-right --count origin/main...HEAD
			// El primer número es Behind; el segundo, Ahead.
			// git log --oneline origin/main..HEAD
			// Salida esperada:
			// Repository    overmind
			// Worktree      42-authentication
			// Branch        feature/42-authentication
			// Base          origin/main
			//
			// Status        clean
			// Ahead         3
			// Behind        1
			//
			// Commits
			// abc123  feat: add authentication
			// def456  test: add authentication tests
			// 987abc  fix: token validation

			fmt.Println("status")
			return nil
		},
	}
}
