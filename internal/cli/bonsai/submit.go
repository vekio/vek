package bonsai

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func newSubmitCmd() *cli.Command {
	return &cli.Command{
		Name:  "submit",
		Usage: "push the current clean branch to origin",
		Action: func(ctx context.Context, c *cli.Command) error {
			// Only a clean task worktree can be published.
			client := newGitClient(c)
			_, worktree, branch, err := resolveTaskWorktree(ctx, client)
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
			// Compare against the latest origin/main before deciding whether there
			// are task commits to push. Falling behind does not trigger a rebase.
			if err := client.FetchOrigin(ctx); err != nil {
				return fmt.Errorf("fetch origin: %w", err)
			}
			if err := client.VerifyOriginMain(ctx); err != nil {
				return err
			}
			ahead, behind, err := client.OriginMainDivergence(ctx)
			if err != nil {
				return fmt.Errorf("compare branch with origin/main: %w", err)
			}
			if ahead == 0 {
				return fmt.Errorf("branch %q has no commits to submit", branch)
			}
			if behind > 0 {
				fmt.Fprintf(c.Root().Writer, "Branch is behind origin/main by %d commit(s)\n", behind)
			}
			// A normal push lets Git reject non-fast-forward updates; never force it.
			if err := client.PushCurrent(ctx); err != nil {
				return fmt.Errorf("push branch %q: %w", branch, err)
			}
			fmt.Fprintf(c.Root().Writer, "Submitted %s to origin; create the pull request manually\n", branch)
			return nil
		},
	}
}
