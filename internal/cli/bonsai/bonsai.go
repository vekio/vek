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
			newCheckoutCmd(),
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

// Show the current worktree, branch, and changes.
func newStatusCmd() *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "show the current worktree, branch, changes, and divergence from origin/main",
		Arguments: []cli.Argument{},
		Flags:     []cli.Flag{},
		Action: func(ctx context.Context, c *cli.Command) error {
			// TODO: Read the worktree, status, and origin/main divergence.
			_, err := fmt.Fprintln(c.Root().Writer, "status")
			return err
		},
	}
}
