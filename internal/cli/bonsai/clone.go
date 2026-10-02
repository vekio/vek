package bonsai

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

// Clone into <path>/.bare and normally check out the default branch beside it.
func newCloneCmd() *cli.Command {
	return &cli.Command{
		Name:      "clone",
		Usage:     "clone a repository and create a worktree for its default branch",
		ArgsUsage: "<repository> [<path>]",
		Arguments: []cli.Argument{
			repoURLArg(),
			pathArg(),
		},
		Flags: []cli.Flag{
			bareOnlyFlag(),
			printPathFlag(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			// Choose the clone folder from the path argument or repository name.
			repoURL := c.StringArg("repoURL")
			root, err := resolveCloneRoot(repoURL, c.StringArg("path"))
			if err != nil {
				return err
			}
			// Create a new folder for the bare repository and its worktrees.
			if err := os.MkdirAll(filepath.Dir(root), 0o755); err != nil {
				return err
			}
			if err := os.Mkdir(root, 0o755); err != nil {
				return err
			}

			// Store Git data in .bare without checking out any files.
			client := newGitClient(c)
			gitDir := filepath.Join(root, ".bare")
			if err := client.Run(ctx, "clone", "--bare", "--", repoURL, gitDir); err != nil {
				// Remove the empty clone folder so the user can retry.
				_ = os.Remove(root)
				return err
			}
			// Let Git commands in the clone folder find .bare.
			if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: ./.bare\n"), 0o644); err != nil {
				return err
			}
			// Configure future fetches to create origin/* remote branches.
			client.In(gitDir)
			if err := client.Run(ctx, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*"); err != nil {
				return err
			}

			destination := root
			if !c.Bool("bare-only") {
				// Fetch origin/* and read the default branch from the cloned HEAD.
				if err := client.Run(ctx, "fetch", "origin"); err != nil {
					return err
				}
				defaultBranch, err := client.Output(ctx, "symbolic-ref", "--short", "HEAD")
				if err != nil {
					return err
				}
				// Map branch names like release/main to folders like release-main.
				folder, err := worktreeName(defaultBranch)
				if err != nil {
					return err
				}
				// Create a worktree for the local branch created by the bare clone.
				destination = filepath.Join(root, folder)
				if err := client.Run(ctx, "worktree", "add", destination, defaultBranch); err != nil {
					return err
				}
				// Track the matching branch on origin for pulls and pushes.
				upstream := "origin/" + defaultBranch
				if err := client.In(destination).Run(ctx, "branch", "--set-upstream-to="+upstream, defaultBranch); err != nil {
					return err
				}
			}

			// Shell integration uses this path to enter the new worktree or root.
			if c.Bool("print-path") {
				fmt.Fprintln(c.Root().Writer, destination)
				return nil
			}
			fmt.Fprintf(c.Root().Writer, "Cloned into %s\n", root)
			return nil
		},
	}
}

// resolveCloneRoot returns the absolute path of the repository container.
// Without a path argument, it uses the repository name in the current directory.
func resolveCloneRoot(repoURL, destinationPath string) (string, error) {
	if destinationPath == "" {
		name, err := repoName(repoURL)
		if err != nil {
			return "", err
		}
		destinationPath = name
	}
	root, err := filepath.Abs(destinationPath)
	if err != nil {
		return "", fmt.Errorf("resolve destination: %w", err)
	}
	return root, nil
}

// repoName extracts the final component of the source for the default path.
func repoName(repoURL string) (string, error) {
	repoPath := repoURL
	if strings.Contains(repoURL, "://") {
		parsed, err := url.Parse(repoURL)
		if err != nil {
			return "", fmt.Errorf("invalid repo URL: %w", err)
		}
		repoPath = parsed.Path
	} else if colon := strings.Index(repoURL, ":"); colon >= 0 && !strings.Contains(repoURL[:colon], "/") {
		// Git treats host:path as SSH when no slash precedes the colon.
		repoPath = repoURL[colon+1:]
	}
	name := strings.TrimSuffix(path.Base(strings.TrimRight(repoPath, "/")), ".git")
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf("cannot infer path from repo URL %q; pass a path explicitly", repoURL)
	}
	return name, nil
}
