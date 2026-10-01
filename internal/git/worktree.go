package git

import (
	"context"
	"fmt"
	"strings"
)

// Worktree describes one entry from git worktree list.
type Worktree struct {
	Path     string
	Branch   string
	Bare     bool
	Detached bool
	Prunable bool
}

// Worktrees lists the worktrees known to Git, including the bare repository.
func (c *Client) Worktrees(ctx context.Context) ([]Worktree, error) {
	output, err := c.Output(ctx, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, err
	}
	return parseWorktrees(output)
}

func parseWorktrees(output string) ([]Worktree, error) {
	if output == "" {
		return nil, nil
	}
	if !strings.HasSuffix(output, "\x00\x00") {
		return nil, fmt.Errorf("incomplete git worktree list output")
	}
	var worktrees []Worktree
	for record := range strings.SplitSeq(strings.TrimSuffix(output, "\x00\x00"), "\x00\x00") {
		var worktree Worktree
		for field := range strings.SplitSeq(record, "\x00") {
			switch {
			case strings.HasPrefix(field, "worktree "):
				worktree.Path = strings.TrimPrefix(field, "worktree ")
			case strings.HasPrefix(field, "branch refs/heads/"):
				worktree.Branch = strings.TrimPrefix(field, "branch refs/heads/")
			case field == "bare":
				worktree.Bare = true
			case field == "detached":
				worktree.Detached = true
			case field == "prunable" || strings.HasPrefix(field, "prunable "):
				worktree.Prunable = true
			}
		}
		if worktree.Path == "" {
			return nil, fmt.Errorf("git worktree list entry has no path")
		}
		worktrees = append(worktrees, worktree)
	}
	return worktrees, nil
}
