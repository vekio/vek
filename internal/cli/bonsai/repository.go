package bonsai

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/vekio/vek/internal/git"
)

// resolveBonsaiGitDir locates a Bonsai clone from its root, bare repository, or a worktree.
func resolveBonsaiGitDir(ctx context.Context, client *git.Client) (string, error) {
	// Find .bare from the clone folder or any of its worktrees.
	gitDir, err := client.Output(ctx, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	if filepath.Base(gitDir) != ".bare" {
		return "", fmt.Errorf("Git directory %q is not a Bonsai clone", gitDir)
	}
	// Check the common repository, not the current worktree.
	bare, err := client.Output(ctx, "--git-dir="+gitDir, "rev-parse", "--is-bare-repository")
	if err != nil {
		return "", err
	}
	if bare != "true" {
		return "", fmt.Errorf("Git directory %q is not a bare repository", gitDir)
	}
	return gitDir, nil
}

// resolveTaskWorktree accepts only a task worktree created by Bonsai.
func resolveTaskWorktree(ctx context.Context, client *git.Client) (gitDir, worktree, branch string, err error) {
	gitDir, err = resolveBonsaiGitDir(ctx, client)
	if err != nil {
		return "", "", "", err
	}
	// Read the current worktree and its branch without changing directories.
	worktree, err = client.Output(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", "", err
	}
	branch, err = client.Output(ctx, "branch", "--show-current")
	if err != nil {
		return "", "", "", err
	}
	if branch == "" {
		return "", "", "", fmt.Errorf("current worktree has a detached HEAD")
	}
	if branch == "main" {
		return "", "", "", fmt.Errorf("main is not a task branch")
	}
	folder, err := worktreeName(branch)
	if err != nil {
		return "", "", "", fmt.Errorf("current branch is not a Bonsai task: %w", err)
	}
	if filepath.Dir(worktree) != filepath.Dir(gitDir) || filepath.Base(worktree) != folder {
		return "", "", "", fmt.Errorf("worktree %q does not match Bonsai task %q", worktree, branch)
	}
	return gitDir, worktree, branch, nil
}
