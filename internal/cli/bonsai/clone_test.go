package bonsai

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestCloneBareAndMainWorktree(t *testing.T) {
	source := makeSourceRepository(t, "main")
	workspace := t.TempDir()
	t.Chdir(workspace)

	output, err := runClone(source)
	if err != nil {
		t.Fatal(err)
	}
	defaultRoot := filepath.Join(workspace, filepath.Base(source))
	if !strings.Contains(output, defaultRoot) {
		t.Fatalf("clone output %q does not contain %q", output, defaultRoot)
	}
	bare := filepath.Join(defaultRoot, ".bare")
	gitFile, err := os.ReadFile(filepath.Join(defaultRoot, ".git"))
	if err != nil || string(gitFile) != "gitdir: ./.bare\n" {
		t.Fatalf("root .git file = %q, %v", gitFile, err)
	}
	if got := strings.TrimSpace(runGit(t, defaultRoot, "rev-parse", "--path-format=absolute", "--git-common-dir")); got != bare {
		t.Fatalf("root common Git directory = %q", got)
	}
	if got := strings.TrimSpace(runGit(t, bare, "rev-parse", "--is-bare-repository")); got != "true" {
		t.Fatalf("bare repository = %q", got)
	}
	if got := strings.TrimSpace(runGit(t, bare, "config", "--get", "remote.origin.fetch")); got != "+refs/heads/*:refs/remotes/origin/*" {
		t.Fatalf("fetch refspec = %q", got)
	}
	mainTree := filepath.Join(defaultRoot, "main")
	if got := strings.TrimSpace(runGit(t, mainTree, "branch", "--show-current")); got != "main" {
		t.Fatalf("main worktree branch = %q", got)
	}
	if got := strings.TrimSpace(runGit(t, mainTree, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "main@{upstream}")); got != "origin/main" {
		t.Fatalf("main upstream = %q", got)
	}
	if got := strings.TrimSpace(runGit(t, mainTree, "rev-parse", "--path-format=absolute", "--git-common-dir")); got != bare {
		t.Fatalf("common Git directory = %q", got)
	}

	explicitRoot := filepath.Join(workspace, "nested", "custom")
	if _, err := runClone("--bare-only", source, explicitRoot); err != nil {
		t.Fatal(err)
	}
	refCheck := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/main")
	refCheck.Dir = filepath.Join(explicitRoot, ".bare")
	if err := refCheck.Run(); err == nil {
		t.Fatal("bare-only clone fetched origin/main")
	}
	if _, err := os.Stat(filepath.Join(explicitRoot, "main")); !os.IsNotExist(err) {
		t.Fatalf("unexpected main worktree: %v", err)
	}
	gitFile, err = os.ReadFile(filepath.Join(explicitRoot, ".git"))
	if err != nil || string(gitFile) != "gitdir: ./.bare\n" {
		t.Fatalf("bare-only root .git file = %q, %v", gitFile, err)
	}
}

func TestCloneUsesSourceDefaultBranch(t *testing.T) {
	source := makeSourceRepository(t, "trunk")
	destination := filepath.Join(t.TempDir(), "clone")
	if _, err := runClone(source, destination); err != nil {
		t.Fatal(err)
	}
	trunk := filepath.Join(destination, "trunk")
	if got := strings.TrimSpace(runGit(t, trunk, "branch", "--show-current")); got != "trunk" {
		t.Fatalf("default worktree branch = %q", got)
	}
	if got := strings.TrimSpace(runGit(t, trunk, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "trunk@{upstream}")); got != "origin/trunk" {
		t.Fatalf("default worktree upstream = %q", got)
	}
}

func TestCloneMapsDefaultBranchSlashToWorktreeFolder(t *testing.T) {
	source := makeSourceRepository(t, "release/main")
	root := filepath.Join(t.TempDir(), "clone")
	if _, err := runClone(source, root); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(root, "release-main")
	if got := strings.TrimSpace(runGit(t, worktree, "branch", "--show-current")); got != "release/main" {
		t.Fatalf("default worktree branch = %q", got)
	}
}

func TestCloneRejectsExistingDestinationAndUnbornDefaultBranch(t *testing.T) {
	source := makeSourceRepository(t, "main")
	existing := filepath.Join(t.TempDir(), "existing")
	if err := os.Mkdir(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runClone(source, existing); !os.IsExist(err) {
		t.Fatalf("existing destination error = %v", err)
	}

	empty := filepath.Join(t.TempDir(), "empty")
	runGit(t, "", "init", "-q", "-b", "main", empty)
	destination := filepath.Join(t.TempDir(), "clone")
	if _, err := runClone(empty, destination); err == nil || !strings.HasPrefix(err.Error(), "git ") {
		t.Fatalf("unborn default branch error = %v", err)
	}
	if got := strings.TrimSpace(runGit(t, filepath.Join(destination, ".bare"), "rev-parse", "--is-bare-repository")); got != "true" {
		t.Fatalf("retained bare clone = %q", got)
	}
}

func TestCloneFailureLeavesNoEmptyDestination(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "clone")
	if _, err := runClone(filepath.Join(t.TempDir(), "missing-repository"), destination); err == nil {
		t.Fatal("expected clone to fail")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("destination remains after failed clone: %v", err)
	}
}

func TestCloneAcceptsRelativeLocalSource(t *testing.T) {
	source := makeSourceRepository(t, "main")
	t.Chdir(filepath.Dir(source))
	root := filepath.Join(t.TempDir(), "clone")
	if _, err := runClone("sample", root); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, filepath.Join(root, "main"), "branch", "--show-current")); got != "main" {
		t.Fatalf("branch from relative source = %q", got)
	}
}

func TestCloneHelpHidesShellFlag(t *testing.T) {
	output, err := runClone("--help")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "--bare-only") || strings.Contains(output, "--print-path") {
		t.Fatalf("clone help flags = %q", output)
	}
}

func TestClonePrintPath(t *testing.T) {
	source := makeSourceRepository(t, "main")
	root := filepath.Join(t.TempDir(), "project")
	output, err := runClone("--print-path", source, root)
	if err != nil {
		t.Fatal(err)
	}
	if output != filepath.Join(root, "main")+"\n" {
		t.Fatalf("clone path output = %q", output)
	}
	bareRoot := filepath.Join(t.TempDir(), "bare-project")
	output, err = runClone("--bare-only", "--print-path", source, bareRoot)
	if err != nil {
		t.Fatal(err)
	}
	if output != bareRoot+"\n" {
		t.Fatalf("bare clone path output = %q", output)
	}
}

func TestRepoName(t *testing.T) {
	for _, tc := range []struct{ repository, want string }{
		{"git@example.com:group/sample.git", "sample"},
		{"example.com:sample.git", "sample"},
		{"ssh://git@example.com/group/sample.git", "sample"},
		{"https://example.com/group/sample.git", "sample"},
		{"https://example.com/group/sample.git/", "sample"},
	} {
		got, err := repoName(tc.repository)
		if err != nil || got != tc.want {
			t.Errorf("repoName(%q) = %q, %v; want %q", tc.repository, got, err, tc.want)
		}
	}
}

func runClone(args ...string) (string, error) {
	var output bytes.Buffer
	var errors bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), append([]string{"vek", "bonsai", "clone"}, args...))
	return output.String(), err
}

func makeSourceRepository(t *testing.T, branch string) string {
	t.Helper()
	source := filepath.Join(t.TempDir(), "sample")
	runGit(t, "", "init", "-q", "-b", branch, source)
	runGit(t, source, "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "-m", "initial")
	return source
}

func runGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return string(output)
}
