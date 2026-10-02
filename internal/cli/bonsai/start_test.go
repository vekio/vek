package bonsai

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestStartFromWorktreeAndProjectRoot(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "new main commit")

	t.Chdir(filepath.Join(root, "main"))
	output, err := runStart("feature/42-authentication")
	if err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "feature-42-authentication")
	if !strings.Contains(output, "feature/42-authentication") || !strings.Contains(output, worktree) {
		t.Fatalf("start output = %q", output)
	}
	if got := strings.TrimSpace(runGit(t, worktree, "branch", "--show-current")); got != "feature/42-authentication" {
		t.Fatalf("worktree branch = %q", got)
	}
	if got, want := strings.TrimSpace(runGit(t, worktree, "rev-parse", "HEAD")), strings.TrimSpace(runGit(t, source, "rev-parse", "HEAD")); got != want {
		t.Fatalf("worktree HEAD = %q, origin main = %q", got, want)
	}
	if _, err := runStart("feature-42-authentication"); err == nil || !strings.HasPrefix(err.Error(), "git worktree:") {
		t.Fatalf("colliding worktree name error = %v", err)
	}

	t.Chdir(root)
	if _, err := runStart("fix/57-cache"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, filepath.Join(root, "fix-57-cache"), "branch", "--show-current")); got != "fix/57-cache" {
		t.Fatalf("second worktree branch = %q", got)
	}
	if _, err := runStart("fix/57-cache"); err == nil || !strings.HasPrefix(err.Error(), "git worktree:") {
		t.Fatalf("duplicate task error = %v", err)
	}
}

func TestStartRejectsInvalidLocationAndTask(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runStart("task"); err == nil || !strings.HasPrefix(err.Error(), "git rev-parse:") {
		t.Fatalf("outside repository error = %v", err)
	}

	source := makeSourceRepository(t, "main")
	t.Chdir(source)
	if _, err := runStart("task"); err == nil || !strings.Contains(err.Error(), "not a Bonsai clone") {
		t.Fatalf("ordinary repository error = %v", err)
	}
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone("--bare-only", source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	if _, err := runStart("../escape"); err == nil || !strings.Contains(err.Error(), "invalid task name") {
		t.Fatalf("invalid task name error = %v", err)
	}
}

func TestStartFromBareCloneWithoutMainWorktree(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone("--bare-only", source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, ".bare"))
	if _, err := runStart("task"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, filepath.Join(root, "task"), "branch", "--show-current")); got != "task" {
		t.Fatalf("worktree branch = %q", got)
	}
}

func TestStartPrintPath(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, ".bare"))
	output, err := runStart("--print-path", "feature/42")
	if err != nil {
		t.Fatal(err)
	}
	if output != filepath.Join(root, "feature-42")+"\n" {
		t.Fatalf("start path output = %q", output)
	}
}

func TestStartAllowsAutomaticUpstreamOnFirstPush(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "config", "branch.autoSetupMerge", "always")
	runGit(t, root, "config", "push.autoSetupRemote", "true")
	runGit(t, root, "config", "push.default", "simple")
	t.Chdir(root)
	if _, err := runStart("feature/42"); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "feature-42")
	if got := strings.TrimSpace(runGit(t, worktree, "for-each-ref", "--format=%(upstream)", "refs/heads/feature/42")); got != "" {
		t.Fatalf("new task upstream = %q, want none", got)
	}
	runGit(t, worktree, "push")
	if got := strings.TrimSpace(runGit(t, worktree, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")); got != "origin/feature/42" {
		t.Fatalf("upstream after push = %q", got)
	}
	if got, want := strings.TrimSpace(runGit(t, source, "rev-parse", "refs/heads/feature/42")), strings.TrimSpace(runGit(t, worktree, "rev-parse", "HEAD")); got != want {
		t.Fatalf("published branch = %q, want %q", got, want)
	}
}

func runStart(args ...string) (string, error) {
	var output, errors bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), append([]string{"vek", "bonsai", "start"}, args...))
	return output.String(), err
}
