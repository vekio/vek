package bonsai

import (
	"context"
	"fmt"
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
			// Map the task name to a folder beside .bare.
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

			// Fetch the latest main and create the task branch and worktree.
			if err := client.Run(ctx, "fetch", "origin"); err != nil {
				return err
			}
			worktree := filepath.Join(root, folder)
			if err := client.Run(ctx, "worktree", "add", "-b", branch, worktree, "origin/main"); err != nil {
				return err
			}
			// Shell integration uses this path to enter the new worktree.
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
