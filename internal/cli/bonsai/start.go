package bonsai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/urfave/cli/v3"
)

func newStartCmd() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "create a branch named by the task name and a worktree from origin/main",
		Description: "For example, feature/42 creates branch feature/42 in worktree feature-42.",
		ArgsUsage:   "<task-name>",
		Arguments: []cli.Argument{
			taskNameArg(),
		},
		Flags: []cli.Flag{
			printPathFlag(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			branch := c.StringArg("task-name")
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
			root := filepath.Dir(gitDir)

			worktree := filepath.Join(root, folder)
			if _, err := os.Lstat(worktree); err == nil {
				return fmt.Errorf("worktree path %q already exists", worktree)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("inspect worktree path %q: %w", worktree, err)
			}
			if err := client.AddWorktree(ctx, branch, worktree); err != nil {
				return fmt.Errorf("start task %q: %w", branch, err)
			}
			if c.Bool("print-path") {
				fmt.Fprintln(c.Root().Writer, worktree)
				return nil
			}
			fmt.Fprintf(c.Root().Writer, "Started %s in %s\n", branch, worktree)
			return nil
		},
	}
}

// worktreeName derives a safe single directory name from the task name.
func worktreeName(taskName string) (string, error) {
	if taskName == "" || strings.ContainsAny(taskName, "\\\x00") {
		return "", fmt.Errorf("invalid task name %q", taskName)
	}
	for _, r := range taskName {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return "", fmt.Errorf("invalid task name %q", taskName)
		}
	}
	for _, segment := range strings.Split(taskName, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("invalid task name %q", taskName)
		}
	}
	name := strings.ReplaceAll(taskName, "/", "-")
	if !filepath.IsLocal(name) {
		return "", fmt.Errorf("invalid task name %q", taskName)
	}
	return name, nil
}
