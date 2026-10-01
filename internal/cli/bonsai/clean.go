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
)

func newCleanCmd() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "remove the current worktree and local branch",
		Flags: []cli.Flag{
			printPathFlag(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			// Reject dirty or non-task worktrees before any destructive operation.
			client := newGitClient(c)
			gitDir, worktree, branch, err := resolveTaskWorktree(ctx, client)
			if err != nil {
				return err
			}
			client.In(worktree)
			clean, err := client.IsClean(ctx)
			if err != nil {
				return fmt.Errorf("inspect current worktree: %w", err)
			}
			if !clean {
				return fmt.Errorf("worktree %q has uncommitted changes", worktree)
			}
			// Refresh remote refs, then ask the user to confirm the squash merge:
			// Git ancestry cannot prove that a squash commit contains this work.
			if err := client.FetchPruneOrigin(ctx); err != nil {
				return fmt.Errorf("fetch origin: %w", err)
			}
			if err := client.VerifyOriginMain(ctx); err != nil {
				return err
			}

			// Shell integration captures stdout as a destination path, so keep
			// the confirmation prompt visible on stderr in that mode.
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
				return fmt.Errorf("read confirmation: %w", err)
			}
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				fmt.Fprintln(messageWriter, "Clean cancelled")
				if c.Bool("print-path") {
					fmt.Fprintln(c.Root().Writer, worktree)
				}
				return nil
			}

			// Update main before deleting the task. Any dirty state or failed
			// fast-forward leaves the task worktree and branch untouched.
			worktrees, err := client.In(gitDir).Worktrees(ctx)
			if err != nil {
				return fmt.Errorf("find main worktree: %w", err)
			}
			nextDir := filepath.Dir(gitDir)
			for _, tree := range worktrees {
				if tree.Bare || tree.Branch != "main" {
					continue
				}
				if tree.Prunable {
					return fmt.Errorf("main worktree %q is unavailable", tree.Path)
				}
				client.In(tree.Path)
				clean, err := client.IsClean(ctx)
				if err != nil {
					return fmt.Errorf("inspect main worktree: %w", err)
				}
				if !clean {
					return fmt.Errorf("main worktree %q has uncommitted changes", tree.Path)
				}
				if err := client.FastForwardOriginMain(ctx); err != nil {
					return fmt.Errorf("update main worktree: %w", err)
				}
				nextDir = tree.Path
				break
			}

			// Run removal from the bare repository, outside the worktree being
			// deleted. Squash merges require -D for the local task branch.
			client.In(gitDir)
			if err := client.RemoveWorktree(ctx, worktree); err != nil {
				return fmt.Errorf("remove worktree %q: %w", worktree, err)
			}
			if err := client.DeleteBranchForce(ctx, branch); err != nil {
				return fmt.Errorf("worktree removed, but could not delete branch %q: %w", branch, err)
			}
			if c.Bool("print-path") {
				fmt.Fprintln(c.Root().Writer, nextDir)
				return nil
			}
			fmt.Fprintf(c.Root().Writer, "Removed %s and %s\nChange directory to %s\n", branch, worktree, nextDir)
			return nil
		},
	}
}
