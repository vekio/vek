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

func TestListWorktreesAndChanges(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, "main"))
	if _, err := runStart("feature/42"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "feature-42", "new.txt"), []byte("change"), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := runList()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"WORKTREE", "BRANCH", "STATUS", "main", "feature/42", "feature-42", "clean", "modified"} {
		if !strings.Contains(output, want) {
			t.Errorf("list output %q is missing %q", output, want)
		}
	}
	if strings.Contains(output, filepath.Join(root, ".bare")) {
		t.Fatalf("bare entry appears in list: %q", output)
	}
	t.Chdir(root)
	if _, err := runList(); err != nil {
		t.Fatalf("list from project root: %v", err)
	}
}

func TestListBareCloneAndDetachedWorktree(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone("--bare-only", source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, ".bare"))
	output, err := runList()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Fields(output); len(got) != 3 || got[0] != "WORKTREE" || got[1] != "BRANCH" || got[2] != "STATUS" {
		t.Fatalf("bare-only list = %q", output)
	}
	if _, err := runStart("feature/42"); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "feature-42")
	runGit(t, worktree, "switch", "--detach", "--quiet")
	output, err = runList()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "feature-42") || !strings.Contains(output, "(detached HEAD)") {
		t.Fatalf("detached worktree missing: %q", output)
	}
}

func TestListRejectsNonBonsaiRepository(t *testing.T) {
	t.Chdir(makeSourceRepository(t, "main"))
	if _, err := runList(); err == nil || !strings.Contains(err.Error(), "not a Bonsai clone") {
		t.Fatalf("list ordinary repository error = %v", err)
	}
}

func TestListRejectsLegacyBareClone(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, "", "clone", "--bare", "--", source, filepath.Join(root, ".git"))
	t.Chdir(root)
	if _, err := runList(); err == nil || !strings.Contains(err.Error(), "not a Bonsai clone") {
		t.Fatalf("list legacy clone error = %v", err)
	}
}

func TestListPrunableWorktree(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, ".bare"))
	if _, err := runStart("feature/42"); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(root, "feature-42")); err != nil {
		t.Fatal(err)
	}
	output, err := runList()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "feature-42") || !strings.Contains(output, "prunable") {
		t.Fatalf("prunable worktree missing: %q", output)
	}
}

func runList(args ...string) (string, error) {
	var output, errors bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), append([]string{"vek", "bonsai", "list"}, args...))
	return output.String(), err
}
