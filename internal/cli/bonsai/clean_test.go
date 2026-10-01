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

func TestCleanRequiresConfirmationAndKeepsRemoteBranch(t *testing.T) {
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
	runGit(t, worktree, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "task commit")
	if _, err := runSubmit(); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "squash merge")
	if output, err := runClean("n\n"); err != nil || !strings.Contains(output, "Clean cancelled") {
		t.Fatalf("declined clean = %q, %v", output, err)
	}
	if _, err := os.Stat(worktree); err != nil {
		t.Fatalf("worktree removed after declining: %v", err)
	}
	output, err := runClean("y\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "Removed feature/42") || !strings.Contains(output, filepath.Join(root, "main")) {
		t.Fatalf("clean output = %q", output)
	}
	if _, err := os.Stat(worktree); !os.IsNotExist(err) {
		t.Fatalf("worktree remains after clean: %v", err)
	}
	if got := strings.TrimSpace(runGit(t, filepath.Join(root, ".git"), "branch", "--list", "feature/42")); got != "" {
		t.Fatalf("local task branch remains: %q", got)
	}
	if got, want := strings.TrimSpace(runGit(t, filepath.Join(root, "main"), "rev-parse", "HEAD")), strings.TrimSpace(runGit(t, source, "rev-parse", "HEAD")); got != want {
		t.Fatalf("main HEAD = %q, want %q", got, want)
	}
	if got := strings.TrimSpace(runGit(t, source, "branch", "--list", "feature/42")); got == "" {
		t.Fatal("remote task branch was deleted")
	}
}

func TestCleanStopsForDirtyTaskOrMain(t *testing.T) {
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
	dirtyTask := filepath.Join(worktree, "dirty.txt")
	if err := os.WriteFile(dirtyTask, []byte("change"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runClean("y\n"); err == nil || !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("dirty task clean error = %v", err)
	}
	if err := os.Remove(dirtyTask); err != nil {
		t.Fatal(err)
	}
	dirtyMain := filepath.Join(root, "main", "dirty.txt")
	if err := os.WriteFile(dirtyMain, []byte("change"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runClean("y\n"); err == nil || !strings.Contains(err.Error(), "main worktree") {
		t.Fatalf("dirty main clean error = %v", err)
	}
	if _, err := os.Stat(worktree); err != nil {
		t.Fatalf("task worktree removed after failed clean: %v", err)
	}
}

func TestCleanPrintPathWithoutMainWorktree(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone("--bare-only", source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, ".git"))
	if _, err := runStart("feature/42"); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "feature-42")
	t.Chdir(worktree)
	output, err := runClean("n\n", "--print-path")
	if err != nil || output != worktree+"\n" {
		t.Fatalf("cancelled clean path output = %q, %v", output, err)
	}
	output, err = runClean("y\n", "--print-path")
	if err != nil {
		t.Fatal(err)
	}
	if output != root+"\n" {
		t.Fatalf("clean path output = %q", output)
	}
}

func TestCleanPreservesTaskWhenMainCannotFastForward(t *testing.T) {
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
	runGit(t, filepath.Join(root, "main"), "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "local main commit")
	runGit(t, source, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "remote main commit")
	t.Chdir(worktree)
	if _, err := runClean("y\n"); err == nil || !strings.Contains(err.Error(), "update main worktree") {
		t.Fatalf("non-fast-forward clean error = %v", err)
	}
	if _, err := os.Stat(worktree); err != nil {
		t.Fatalf("task worktree removed after failed main update: %v", err)
	}
	if got := strings.TrimSpace(runGit(t, filepath.Join(root, ".git"), "branch", "--list", "feature/42")); got == "" {
		t.Fatal("task branch was deleted after failed main update")
	}
}

func runClean(answer string, args ...string) (string, error) {
	var output, errors bytes.Buffer
	root := &cli.Command{Name: "vek", Reader: strings.NewReader(answer), Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), append([]string{"vek", "bonsai", "clean"}, args...))
	return output.String(), err
}
