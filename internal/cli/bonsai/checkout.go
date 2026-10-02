package bonsai

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

func newCheckoutCmd() *cli.Command {
	return &cli.Command{
		Name:      "checkout",
		Usage:     "create a worktree for an existing local or origin branch",
		ArgsUsage: "<branch>",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "branch", Required: true},
		},
		Flags: []cli.Flag{
			printPathFlag(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			// Map the branch name to a folder beside .bare.
			branch := c.StringArg("branch")
			folder, err := worktreeName(branch)
			if err != nil {
				return err
			}
			client := newGitClient(c)
			gitDir, err := resolveBonsaiGitDir(ctx, client)
			if err != nil {
				return err
			}
			client.In(gitDir)
			worktree := filepath.Join(filepath.Dir(gitDir), folder)
			upstream := "origin/" + branch
			// Reuse a local branch without fetching or changing its commits.
			localBranch, err := client.Output(ctx, "branch", "--list", "--", branch)
			if err != nil {
				return err
			}
			if localBranch == "" {
				// Otherwise fetch origin and create a branch with remote tracking.
				if err := client.Run(ctx, "fetch", "--prune", "origin"); err != nil {
					return err
				}
				if err := client.Run(ctx, "worktree", "add", "--track", "-b", branch, worktree, "refs/remotes/"+upstream); err != nil {
					return err
				}
			} else {
				if err := client.Run(ctx, "worktree", "add", worktree, branch); err != nil {
					return err
				}
				// Keep the upstream, or use a known origin branch if none is set.
				currentUpstream, err := client.Output(ctx, "for-each-ref", "--format=%(upstream)", "refs/heads/"+branch)
				if err != nil {
					return err
				}
				if currentUpstream == "" {
					remoteBranch, err := client.Output(ctx, "branch", "--remotes", "--list", "--", upstream)
					if err != nil {
						return err
					}
					if remoteBranch != "" {
						if err := client.Run(ctx, "branch", "--set-upstream-to="+upstream, branch); err != nil {
							return err
						}
					}
				}
			}
			// Shell integration uses this path to enter the new worktree.
			if c.Bool("print-path") {
				fmt.Fprintln(c.Root().Writer, worktree)
				return nil
			}
			fmt.Fprintf(c.Root().Writer, "Created worktree for %s in %s\n", branch, worktree)
			return nil
		},
	}
}
