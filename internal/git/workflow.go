package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// CurrentWorktree returns the current worktree root and branch.
// A detached HEAD has an empty branch name.
func (c *Client) CurrentWorktree(ctx context.Context) (path, branch string, err error) {
	path, err = c.Output(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", fmt.Errorf("find current worktree: %w", err)
	}
	branch, err = c.Output(ctx, "branch", "--show-current")
	if err != nil {
		return "", "", fmt.Errorf("find current branch: %w", err)
	}
	return path, branch, nil
}

// IsClean reports whether the working tree has no staged, unstaged, or
// untracked changes.
func (c *Client) IsClean(ctx context.Context) (bool, error) {
	changes, err := c.Output(ctx, "status", "--porcelain=v1", "-z")
	if err != nil {
		return false, err
	}
	return changes == "", nil
}

// FetchOrigin updates remote-tracking refs from origin.
func (c *Client) FetchOrigin(ctx context.Context) error {
	return c.Run(ctx, "fetch", "origin")
}

// FetchPruneOrigin also removes remote-tracking refs deleted from origin.
func (c *Client) FetchPruneOrigin(ctx context.Context) error {
	return c.Run(ctx, "fetch", "--prune", "origin")
}

// VerifyOriginMain requires origin/main to resolve to a commit.
func (c *Client) VerifyOriginMain(ctx context.Context) error {
	if _, err := c.Output(ctx, "rev-parse", "--verify", "--quiet", "refs/remotes/origin/main^{commit}"); err != nil {
		return fmt.Errorf("origin/main is unavailable: %w", err)
	}
	return nil
}

// OriginMainDivergence returns commits ahead of and behind origin/main.
// Git prints the behind count first for origin/main...HEAD.
func (c *Client) OriginMainDivergence(ctx context.Context) (ahead, behind int, err error) {
	output, err := c.Output(ctx, "rev-list", "--left-right", "--count", "origin/main...HEAD")
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(output)
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("unexpected divergence output %q", output)
	}
	behind, err = strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse behind count %q: %w", fields[0], err)
	}
	ahead, err = strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse ahead count %q: %w", fields[1], err)
	}
	return ahead, behind, nil
}

// PushCurrent publishes HEAD to origin and sets its upstream without force.
func (c *Client) PushCurrent(ctx context.Context) error {
	return c.Run(ctx, "push", "--set-upstream", "origin", "HEAD")
}

// FastForwardOriginMain merges origin/main into the current branch only when
// Git can fast-forward it.
func (c *Client) FastForwardOriginMain(ctx context.Context) error {
	return c.Run(ctx, "merge", "--ff-only", "origin/main")
}

// RemoveWorktree removes a registered worktree without forcing dirty changes.
func (c *Client) RemoveWorktree(ctx context.Context, path string) error {
	return c.Run(ctx, "worktree", "remove", path)
}

// DeleteBranchForce deletes only the local branch. It is needed after a squash
// merge, which does not preserve the task branch as an ancestor of main.
func (c *Client) DeleteBranchForce(ctx context.Context, branch string) error {
	return c.Run(ctx, "branch", "-D", branch)
}
