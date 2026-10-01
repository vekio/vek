package bonsai

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestSubmitPushesCommittedTaskWithoutRebase(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, "main"))
	if _, err := runStart("feature/42"); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "feature-42")
	t.Chdir(worktree)
	if _, err := runSubmit(); err == nil || !strings.Contains(err.Error(), "no commits") {
		t.Fatalf("submit without commits error = %v", err)
	}
	runGit(t, worktree, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "task commit")
	if err := os.WriteFile(filepath.Join(worktree, "untracked.txt"), []byte("change"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runSubmit(); err == nil || !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("dirty submit error = %v", err)
	}
	if err := os.Remove(filepath.Join(worktree, "untracked.txt")); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "new main commit")
	output, err := runSubmit()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "behind origin/main by 1") || !strings.Contains(output, "Submitted feature/42") {
		t.Fatalf("submit output = %q", output)
	}
	if got, want := strings.TrimSpace(runGit(t, source, "rev-parse", "refs/heads/feature/42")), strings.TrimSpace(runGit(t, worktree, "rev-parse", "HEAD")); got != want {
		t.Fatalf("remote branch = %q, want %q", got, want)
	}
	if got := strings.TrimSpace(runGit(t, worktree, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")); got != "origin/feature/42" {
		t.Fatalf("upstream = %q", got)
	}
}

func TestSubmitRejectsMainAndDetachedHead(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, "main"))
	if _, err := runSubmit(); err == nil || !strings.Contains(err.Error(), "main is not a task") {
		t.Fatalf("submit main error = %v", err)
	}
	if _, err := runStart("feature/42"); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "feature-42")
	runGit(t, worktree, "switch", "--detach", "--quiet")
	t.Chdir(worktree)
	if _, err := runSubmit(); err == nil || !strings.Contains(err.Error(), "detached HEAD") {
		t.Fatalf("submit detached HEAD error = %v", err)
	}
}

func runSubmit() (string, error) {
	var output, errors bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), []string{"vek", "bonsai", "submit"})
	return output.String(), err
}
