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

func TestInitFishChangesDirectoryAfterSuccessfulCommand(t *testing.T) {
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish is not installed")
	}
	script, err := runInit("fish")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	initPath := filepath.Join(directory, "bonsai.fish")
	if err := os.WriteFile(initPath, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	vekPath := filepath.Join(directory, "vek")
	fakeVek := "#!/bin/sh\nif [ \"$4\" = fail ]; then exit 7; fi\nprintf '%s\\n' \"$BONSAI_TEST_DEST\"\n"
	if err := os.WriteFile(vekPath, []byte(fakeVek), 0o755); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(directory, "worktree with spaces")
	if err := os.Mkdir(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("fish", "--no-config", "-c", `source $argv[1]; bonsai start feature/42; pwd`, initPath)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), "PATH="+directory+string(os.PathListSeparator)+os.Getenv("PATH"), "BONSAI_TEST_DEST="+destination)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run fish integration: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != destination {
		t.Fatalf("Fish directory = %q, want %q", got, destination)
	}

	cmd = exec.Command("fish", "--no-config", "-c", `source $argv[1]; bonsai start fail; set -l result $status; pwd; exit $result`, initPath)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), "PATH="+directory+string(os.PathListSeparator)+os.Getenv("PATH"), "BONSAI_TEST_DEST="+destination)
	output, err = cmd.CombinedOutput()
	if err == nil {
		t.Fatal("failed Bonsai command succeeded")
	}
	if got := strings.TrimSpace(string(output)); got != directory {
		t.Fatalf("Fish changed directory after failure: %q", got)
	}
}

func TestInitRejectsUnsupportedShell(t *testing.T) {
	output, err := runInit("zsh")
	if err == nil || !strings.Contains(err.Error(), "unsupported shell") || output != "" {
		t.Fatalf("init zsh = %q, %v", output, err)
	}
}

func TestInitBashChangesDirectoryAfterSuccessfulCommand(t *testing.T) {
	script, err := runInit("bash")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	initPath := filepath.Join(directory, "bonsai.bash")
	if err := os.WriteFile(initPath, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	vekPath := filepath.Join(directory, "vek")
	fakeVek := "#!/bin/sh\nif [ \"$4\" = fail ]; then exit 7; fi\nprintf '%s\\n' \"$BONSAI_TEST_DEST\"\n"
	if err := os.WriteFile(vekPath, []byte(fakeVek), 0o755); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(directory, "worktree with spaces")
	if err := os.Mkdir(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "--noprofile", "--norc", "-c", `source "$1"; bonsai start feature/42; pwd`, "bash", initPath)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), "PATH="+directory+string(os.PathListSeparator)+os.Getenv("PATH"), "BONSAI_TEST_DEST="+destination)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run Bash integration: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != destination {
		t.Fatalf("Bash directory = %q, want %q", got, destination)
	}
}

func runInit(args ...string) (string, error) {
	var output, errors bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), append([]string{"vek", "bonsai", "init"}, args...))
	return output.String(), err
}
