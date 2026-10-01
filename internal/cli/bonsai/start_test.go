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
	if _, err := runStart("feature-42-authentication"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("colliding worktree name error = %v", err)
	}

	t.Chdir(root)
	if _, err := runStart("fix/57-cache"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, filepath.Join(root, "fix-57-cache"), "branch", "--show-current")); got != "fix/57-cache" {
		t.Fatalf("second worktree branch = %q", got)
	}
	if _, err := runStart("fix/57-cache"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate task error = %v", err)
	}
}

func TestStartRejectsInvalidLocationAndTask(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runStart("task"); err == nil || !strings.Contains(err.Error(), "find Bonsai repository") {
		t.Fatalf("outside repository error = %v", err)
	}

	source := makeSourceRepository(t, "main")
	t.Chdir(source)
	if _, err := runStart("task"); err == nil || !strings.Contains(err.Error(), "not a bare repository") {
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
	t.Chdir(filepath.Join(root, ".git"))
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
	t.Chdir(filepath.Join(root, ".git"))
	output, err := runStart("--print-path", "feature/42")
	if err != nil {
		t.Fatal(err)
	}
	if output != filepath.Join(root, "feature-42")+"\n" {
		t.Fatalf("start path output = %q", output)
	}
}

func runStart(args ...string) (string, error) {
	var output, errors bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), append([]string{"vek", "bonsai", "start"}, args...))
	return output.String(), err
}
