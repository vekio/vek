package bonsai

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/vekio/vek/internal/git"
)

// resolveBonsaiGitDir locates a Bonsai clone from its root, bare repository, or a worktree.
func resolveBonsaiGitDir(ctx context.Context, client *git.Client) (string, error) {
	gitDir, err := client.Output(ctx, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("find Bonsai repository: %w", err)
	}
	if filepath.Base(gitDir) != ".git" {
		return "", fmt.Errorf("Git directory %q is not a Bonsai clone", gitDir)
	}
	bare, err := client.Output(ctx, "--git-dir="+gitDir, "rev-parse", "--is-bare-repository")
	if err != nil {
		return "", fmt.Errorf("inspect Bonsai repository: %w", err)
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
	worktree, branch, err = client.CurrentWorktree(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("run from a Bonsai task worktree: %w", err)
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
