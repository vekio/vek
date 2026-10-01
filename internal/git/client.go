package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Client executes Git commands without invoking a shell.
type Client struct {
	dir    string
	stdout io.Writer
	stderr io.Writer
}

// New sends Git stdout and stderr to their respective writers.
func New(stdout, stderr io.Writer) *Client {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	return &Client{stdout: stdout, stderr: stderr}
}

// In sets the working directory for subsequent commands.
func (c *Client) In(directory string) *Client {
	c.dir = directory
	return c
}

func (c *Client) Run(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("git command is required")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.dir
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", args[0], err)
	}
	return nil
}

// Output returns trimmed Git stdout for commands whose result must be parsed.
func (c *Client) Output(ctx context.Context, args ...string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("git command is required")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return "", fmt.Errorf("git %s: %w: %s", args[0], err, message)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CloneBare clones repoURL into gitDir without checking out a worktree.
func (c *Client) CloneBare(ctx context.Context, repoURL, gitDir string) error {
	return c.Run(ctx, "clone", "--bare", "--", repoURL, gitDir)
}

// Config sets the refspec used by future fetches from origin.
func (c *Client) Config(ctx context.Context) error {
	return c.Run(ctx, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
}

// AddWorktree fetches origin and creates branch and its worktree from origin/main.
func (c *Client) AddWorktree(ctx context.Context, branch, path string) error {
	if err := c.FetchOrigin(ctx); err != nil {
		return fmt.Errorf("fetch origin: %w", err)
	}
	return c.Run(ctx, "worktree", "add", "-b", branch, path, "origin/main")
}
