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

func TestCheckoutExistingRemoteBranch(t *testing.T) {
	for _, beforeClone := range []bool{true, false} {
		name := "branch-created-after-clone"
		if beforeClone {
			name = "branch-present-at-clone"
		}
		t.Run(name, func(t *testing.T) {
			source := makeSourceRepository(t, "main")
			createRemoteBranch := func() {
				runGit(t, source, "switch", "-c", "feature/42")
				runGit(t, source, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "remote task")
				runGit(t, source, "switch", "main")
			}
			if beforeClone {
				createRemoteBranch()
			}
			root := filepath.Join(t.TempDir(), "project")
			if _, err := runClone(source, root); err != nil {
				t.Fatal(err)
			}
			if !beforeClone {
				createRemoteBranch()
				t.Chdir(filepath.Join(root, "main"))
			} else {
				t.Chdir(root)
			}
			output, err := runCheckout("--print-path", "feature/42")
			if err != nil {
				t.Fatal(err)
			}
			worktree := filepath.Join(root, "feature-42")
			if output != worktree+"\n" {
				t.Fatalf("checkout path output = %q", output)
			}
			if got := strings.TrimSpace(runGit(t, worktree, "branch", "--show-current")); got != "feature/42" {
				t.Fatalf("worktree branch = %q", got)
			}
			if got, want := strings.TrimSpace(runGit(t, worktree, "rev-parse", "HEAD")), strings.TrimSpace(runGit(t, source, "rev-parse", "feature/42")); got != want {
				t.Fatalf("worktree HEAD = %q, want %q", got, want)
			}
			if got := strings.TrimSpace(runGit(t, worktree, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")); got != "origin/feature/42" {
				t.Fatalf("worktree upstream = %q", got)
			}
			if _, err := runCheckout("feature/42"); err == nil || !strings.HasPrefix(err.Error(), "git worktree:") {
				t.Fatalf("duplicate checkout error = %v", err)
			}
		})
	}
}

func TestCheckoutPreservesLocalCommitsAndUpstream(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone("--bare-only", source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, ".bare"))
	if _, err := runStart("feature/42"); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "feature-42")
	runGit(t, worktree, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "local task")
	localHead := strings.TrimSpace(runGit(t, worktree, "rev-parse", "HEAD"))
	runGit(t, root, "worktree", "remove", worktree)
	runGit(t, source, "branch", "feature/42")
	output, err := runCheckout("feature/42")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "Created worktree for feature/42") || !strings.Contains(output, worktree) {
		t.Fatalf("checkout output = %q", output)
	}
	if got := strings.TrimSpace(runGit(t, worktree, "rev-parse", "HEAD")); got != localHead {
		t.Fatalf("local HEAD changed: %q, want %q", got, localHead)
	}
	if got := strings.TrimSpace(runGit(t, worktree, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")); got != "origin/main" {
		t.Fatalf("local upstream = %q", got)
	}
}

func TestCheckoutRejectsMissingAndDeletedRemoteBranches(t *testing.T) {
	source := makeSourceRepository(t, "main")
	runGit(t, source, "branch", "feature/42")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "branch", "-D", "feature/42")
	runGit(t, root, "branch", "-D", "feature/42")
	t.Chdir(root)
	for _, branch := range []string{"missing", "feature/42"} {
		if _, err := runCheckout(branch); err == nil || !strings.HasPrefix(err.Error(), "git worktree:") {
			t.Fatalf("checkout %q error = %v", branch, err)
		}
		folder, _ := worktreeName(branch)
		if _, err := os.Stat(filepath.Join(root, folder)); !os.IsNotExist(err) {
			t.Fatalf("worktree created for missing remote %q: %v", branch, err)
		}
	}
}

func TestCheckoutLocalOnlyBranchWithoutOrigin(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone("--bare-only", source, root); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "branch", "local/task")
	localHead := strings.TrimSpace(runGit(t, root, "rev-parse", "refs/heads/local/task"))
	runGit(t, root, "remote", "remove", "origin")
	t.Chdir(root)
	output, err := runCheckout("--print-path", "local/task")
	if err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "local-task")
	if output != worktree+"\n" {
		t.Fatalf("local checkout path = %q", output)
	}
	if got := strings.TrimSpace(runGit(t, worktree, "branch", "--show-current")); got != "local/task" {
		t.Fatalf("local branch = %q", got)
	}
	if got := strings.TrimSpace(runGit(t, worktree, "rev-parse", "HEAD")); got != localHead {
		t.Fatalf("local HEAD = %q, want %q", got, localHead)
	}
	if got := strings.TrimSpace(runGit(t, worktree, "for-each-ref", "--format=%(upstream)", "refs/heads/local/task")); got != "" {
		t.Fatalf("unexpected upstream for local-only branch: %q", got)
	}
}

func TestCheckoutRejectsInvalidBranches(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	for _, branch := range []string{"../escape", "task space", "-option", "feature/*", "@{-1}"} {
		if _, err := runCheckout("--", branch); err == nil {
			t.Fatalf("checkout invalid branch %q error = %v", branch, err)
		}
		if folder, err := worktreeName(branch); err == nil {
			if _, err := os.Stat(filepath.Join(root, folder)); !os.IsNotExist(err) {
				t.Fatalf("invalid branch %q created a worktree: %v", branch, err)
			}
		}
	}
}

func TestCheckoutPreservesChangesInExistingWorktree(t *testing.T) {
	source := makeSourceRepository(t, "main")
	runGit(t, source, "branch", "feature/42")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(root, "other-folder")
	runGit(t, root, "worktree", "add", existing, "feature/42")
	dirtyFile := filepath.Join(existing, "pending.txt")
	if err := os.WriteFile(dirtyFile, []byte("pending work"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	if _, err := runCheckout("feature/42"); err == nil || !strings.HasPrefix(err.Error(), "git worktree:") {
		t.Fatalf("branch already checked out error = %v", err)
	}
	if got, err := os.ReadFile(dirtyFile); err != nil || string(got) != "pending work" {
		t.Fatalf("pending changes = %q, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, "feature-42")); !os.IsNotExist(err) {
		t.Fatalf("duplicate worktree created: %v", err)
	}
}

func runCheckout(args ...string) (string, error) {
	var output, errors bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), append([]string{"vek", "bonsai", "checkout"}, args...))
	return output.String(), err
}
