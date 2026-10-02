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
			// Find the current task and require all changes to be committed.
			client := newGitClient(c)
			_, worktree, branch, err := resolveTaskWorktree(ctx, client)
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
			// Fetch main and count commits behind and ahead of it.
			if err := client.Run(ctx, "fetch", "origin"); err != nil {
				return err
			}
			divergence, err := client.Output(ctx, "rev-list", "--left-right", "--count", "origin/main...HEAD")
			if err != nil {
				return err
			}
			var behind, ahead int
			if _, err := fmt.Sscan(divergence, &behind, &ahead); err != nil {
				return err
			}
			if ahead == 0 {
				return fmt.Errorf("branch %q has no commits to submit", branch)
			}
			if behind > 0 {
				fmt.Fprintf(c.Root().Writer, "Branch is behind origin/main by %d commit(s)\n", behind)
			}
			// Push the current branch and set its upstream without forcing.
			if err := client.Run(ctx, "push", "--set-upstream", "origin", "HEAD"); err != nil {
				return err
			}
			fmt.Fprintf(c.Root().Writer, "Submitted %s to origin; create the pull request manually\n", branch)
			return nil
		},
	}
}
