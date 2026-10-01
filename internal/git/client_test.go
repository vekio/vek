package git

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientStreamsOutputAndChangesDirectory(t *testing.T) {
	ctx := context.Background()
	var output, errors bytes.Buffer
	client := New(&output, &errors)
	first := filepath.Join(t.TempDir(), "first")
	second := filepath.Join(t.TempDir(), "second")
	for _, dir := range []string{first, second} {
		if err := client.In(filepath.Dir(dir)).Run(ctx, "init", "-q", dir); err != nil {
			t.Fatal(err)
		}
	}

	if err := client.In(first).Run(ctx, "rev-parse", "--show-toplevel"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != first {
		t.Fatalf("first repository output = %q", got)
	}
	output.Reset()
	if err := client.In(second).Run(ctx, "rev-parse", "--show-toplevel"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != second {
		t.Fatalf("second repository output = %q", got)
	}
	output.Reset()
	if err := client.Run(ctx, "rev-parse", "--verify", "missing-ref"); err == nil {
		t.Fatal("expected Git error")
	}
	if output.Len() != 0 {
		t.Fatalf("Git stderr leaked to stdout: %q", output.String())
	}
	if !strings.Contains(errors.String(), "fatal:") {
		t.Fatalf("Git stderr was not forwarded: %q", errors.String())
	}
}

func TestAddWorktreeStartsFromOriginMain(t *testing.T) {
	ctx := context.Background()
	var output, errors bytes.Buffer
	client := New(&output, &errors)
	parent := t.TempDir()
	source := filepath.Join(parent, "source")
	if err := client.In(parent).Run(ctx, "init", "-q", "-b", "main", source); err != nil {
		t.Fatal(err)
	}
	client.In(source)
	if err := client.Run(ctx, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "initial"); err != nil {
		t.Fatal(err)
	}
	bare := filepath.Join(parent, "bare.git")
	if err := client.In(parent).CloneBare(ctx, source, bare); err != nil {
		t.Fatal(err)
	}
	client.In(bare)
	if err := client.Config(ctx); err != nil {
		t.Fatal(err)
	}
	if err := client.Run(ctx, "show-ref", "--verify", "--quiet", "refs/remotes/origin/main"); err == nil {
		t.Fatal("origin/main exists before AddWorktree fetch")
	}
	invalidWorktree := filepath.Join(parent, "escape")
	if err := client.AddWorktree(ctx, "../escape", invalidWorktree); err == nil {
		t.Fatalf("invalid branch error = %v", err)
	}
	if _, err := os.Stat(invalidWorktree); !os.IsNotExist(err) {
		t.Fatalf("invalid branch created a worktree: %v", err)
	}
	if err := client.Run(ctx, "show-ref", "--verify", "--quiet", "refs/remotes/origin/main"); err != nil {
		t.Fatal("AddWorktree did not fetch origin before worktree add")
	}

	worktree := filepath.Join(parent, "task")
	if err := client.AddWorktree(ctx, "feature/task", worktree); err != nil {
		t.Fatal(err)
	}
	if err := client.Run(ctx, "show-ref", "--verify", "--quiet", "refs/remotes/origin/main"); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := client.In(worktree).Run(ctx, "branch", "--show-current"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != "feature/task" {
		t.Fatalf("worktree branch = %q", got)
	}
}
