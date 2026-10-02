package bonsai

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/vekio/vek/internal/git"
)

func newCleanCmd() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "remove the current worktree and local branch",
		Flags: []cli.Flag{
			printPathFlag(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			// Find the current task and require a clean worktree.
			client := newGitClient(c)
			gitDir, worktree, branch, err := resolveTaskWorktree(ctx, client)
			if err != nil {
				return err
			}
			client.In(worktree)
			changes, err := client.Output(ctx, "status", "--porcelain=v1", "-z")
			if err != nil {
				return err
			}
			if changes != "" {
				return fmt.Errorf("worktree %q has uncommitted changes", worktree)
			}
			// Refresh origin before updating main.
			if err := client.Run(ctx, "fetch", "--prune", "origin"); err != nil {
				return err
			}
			// Ask about the squash merge; Git history cannot confirm it.
			// Keep prompts on stderr when stdout is used for the next path.
			messageWriter := c.Root().Writer
			if c.Bool("print-path") {
				messageWriter = c.Root().ErrWriter
			}
			fmt.Fprintf(messageWriter, "Branch: %s\nWorktree: %s\n", branch, worktree)
			fmt.Fprintln(messageWriter, "Bonsai cannot verify a squash merge.")
			fmt.Fprint(messageWriter, "Confirm this work was squash-merged into main? [y/N] ")
			reader := c.Root().Reader
			if reader == nil {
				reader = os.Stdin
			}
			answer, err := bufio.NewReader(reader).ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				fmt.Fprintln(messageWriter, "Clean cancelled")
				if c.Bool("print-path") {
					fmt.Fprintln(c.Root().Writer, worktree)
				}
				return nil
			}

			// Find the main worktree and update it before removing the task.
			output, err := client.In(gitDir).Output(ctx, "worktree", "list", "--porcelain", "-z")
			if err != nil {
				return err
			}
			worktrees, err := git.ParseWorktrees(output)
			if err != nil {
				return err
			}
			nextDir := filepath.Dir(gitDir)
			for _, tree := range worktrees {
				if tree.Bare || tree.Branch != "main" {
					continue
				}
				client.In(tree.Path)
				changes, err := client.Output(ctx, "status", "--porcelain=v1", "-z")
				if err != nil {
					return err
				}
				if changes != "" {
					return fmt.Errorf("main worktree %q has uncommitted changes", tree.Path)
				}
				if err := client.Run(ctx, "merge", "--ff-only", "origin/main"); err != nil {
					return err
				}
				nextDir = tree.Path
				break
			}

			// Remove the worktree from .bare, outside the folder being deleted.
			client.In(gitDir)
			if err := client.Run(ctx, "worktree", "remove", worktree); err != nil {
				return err
			}
			// Delete the local branch; squash merges require -D.
			if err := client.Run(ctx, "branch", "-D", branch); err != nil {
				return err
			}
			// Return to main, or the clone folder when no main worktree exists.
			if c.Bool("print-path") {
				fmt.Fprintln(c.Root().Writer, nextDir)
				return nil
			}
			fmt.Fprintf(c.Root().Writer, "Removed %s and %s\nChange directory to %s\n", branch, worktree, nextDir)
			return nil
		},
	}
}
