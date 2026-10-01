package bonsai

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v3"
	"github.com/vekio/vek/internal/tui"
)

func newListCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "list worktrees, branches, and working tree changes",
		Action: func(ctx context.Context, c *cli.Command) error {
			client := newGitClient(c)
			gitDir, err := resolveBonsaiGitDir(ctx, client)
			if err != nil {
				return err
			}
			root := filepath.Dir(gitDir)
			worktrees, err := client.In(gitDir).Worktrees(ctx)
			if err != nil {
				return fmt.Errorf("list worktrees: %w", err)
			}

			type row struct{ name, branch, status string }
			rows := make([]row, 0, len(worktrees))
			for _, worktree := range worktrees {
				if worktree.Bare {
					continue
				}
				name, err := filepath.Rel(root, worktree.Path)
				if err != nil || !filepath.IsLocal(name) {
					name = worktree.Path
				}
				branch := worktree.Branch
				if worktree.Detached {
					branch = "(detached HEAD)"
				} else if branch == "" {
					branch = "(unknown)"
				}
				status := "prunable"
				if !worktree.Prunable {
					changes, err := client.In(worktree.Path).Output(ctx, "status", "--porcelain=v1", "-z")
					if err != nil {
						return fmt.Errorf("inspect worktree %q: %w", worktree.Path, err)
					}
					status = "clean"
					if changes != "" {
						status = "modified"
					}
				}
				rows = append(rows, row{name, branch, status})
			}
			tableRows := make([][]string, 0, len(rows))
			for _, row := range rows {
				tableRows = append(tableRows, []string{row.name, row.branch, row.status})
			}
			table, err := tui.NewTable([]string{"WORKTREE", "BRANCH", "STATUS"}, tableRows)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(c.Root().Writer, table.View())
			return err
		},
	}
}
