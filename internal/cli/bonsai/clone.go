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

// Clone into <path>/.git and normally check out the default branch beside it.
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
			repoURL := c.StringArg("repoURL")
			root, err := resolveCloneRoot(repoURL, c.StringArg("path"))
			if err != nil {
				return err
			}
			if _, err := os.Lstat(root); err == nil {
				return fmt.Errorf("destination %q already exists", root)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("inspect destination %q: %w", root, err)
			}
			if err := os.MkdirAll(root, 0o755); err != nil {
				return fmt.Errorf("create destination %q: %w", root, err)
			}

			client := newGitClient(c)
			gitDir := filepath.Join(root, ".git")
			if err := client.CloneBare(ctx, repoURL, gitDir); err != nil {
				// Git usually removes .git on failure. Remove the empty parent too
				// so the same destination can be used for another attempt.
				_ = os.Remove(root)
				return fmt.Errorf("clone into %q: %w", gitDir, err)
			}
			client.In(gitDir)
			if err := client.Config(ctx); err != nil {
				return fmt.Errorf("configure clone at %q: %w", root, err)
			}

			destination := root
			if !c.Bool("bare-only") {
				// Bare clones store the source branches locally. Fetch after setting
				// a refspec so origin/* is available to later Bonsai commands.
				if err := client.FetchOrigin(ctx); err != nil {
					return fmt.Errorf("fetch origin: %w", err)
				}
				defaultBranch, err := client.Output(ctx, "symbolic-ref", "--quiet", "--short", "HEAD")
				if err != nil {
					return fmt.Errorf("find default branch: %w", err)
				}
				upstream := "origin/" + defaultBranch
				if _, err := client.Output(ctx, "rev-parse", "--verify", "--quiet", "refs/remotes/"+upstream+"^{commit}"); err != nil {
					return fmt.Errorf("upstream branch %q unavailable: %w", upstream, err)
				}
				folder, err := worktreeName(defaultBranch)
				if err != nil {
					return fmt.Errorf("invalid default branch %q: %w", defaultBranch, err)
				}
				// The bare clone already created this branch locally; do not use -b.
				destination = filepath.Join(root, folder)
				if err := client.Run(ctx, "worktree", "add", destination, defaultBranch); err != nil {
					return fmt.Errorf("create worktree for %q: %w", defaultBranch, err)
				}
				if err := client.In(destination).Run(ctx, "branch", "--set-upstream-to="+upstream, defaultBranch); err != nil {
					return fmt.Errorf("set upstream %q: %w", upstream, err)
				}
			}

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
	if strings.TrimSpace(repoURL) == "" {
		return "", fmt.Errorf("repo URL is required")
	}
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
