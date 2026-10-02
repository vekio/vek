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

func TestInitShellChangesDirectory(t *testing.T) {
	for _, shell := range []string{"fish", "bash"} {
		t.Run(shell, func(t *testing.T) {
			if _, err := exec.LookPath(shell); err != nil {
				t.Skip(shell + " is not installed")
			}
			script, err := runInit(shell)
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			initPath := filepath.Join(directory, "bonsai."+shell)
			if err := os.WriteFile(initPath, []byte(script), 0o644); err != nil {
				t.Fatal(err)
			}
			fakeVek := `#!/bin/sh
if [ "$1" != bonsai ]; then
    printf '%s\n' "$@"
    exit 0
fi
if [ "$3" != --print-path ]; then
    echo 'missing --print-path' >&2
    exit 2
fi
if [ "$4" = fail ]; then exit 7; fi
printf '%s\n' "$BONSAI_TEST_DEST"
`
			if err := os.WriteFile(filepath.Join(directory, "vek"), []byte(fakeVek), 0o755); err != nil {
				t.Fatal(err)
			}
			destination := filepath.Join(directory, "worktree with spaces")
			if err := os.Mkdir(destination, 0o755); err != nil {
				t.Fatal(err)
			}
			run := func(arguments ...string) (string, error) {
				var args []string
				if shell == "fish" {
					args = []string{"--no-config", "-c", `source $argv[1]; source $argv[1]; $argv[2] $argv[3..-1]; set -l result $status; pwd; exit $result`, initPath}
				} else {
					args = []string{"--noprofile", "--norc", "-c", `source "$1"; source "$1"; shift; "$@"; result=$?; pwd; exit "$result"`, "bash", initPath}
				}
				cmd := exec.Command(shell, append(args, arguments...)...)
				cmd.Dir = directory
				cmd.Env = append(os.Environ(), "PATH="+directory+string(os.PathListSeparator)+os.Getenv("PATH"), "BONSAI_TEST_DEST="+destination)
				output, err := cmd.CombinedOutput()
				return strings.TrimSpace(string(output)), err
			}
			for _, prefix := range [][]string{{"vek", "bonsai"}, {"bonsai"}} {
				for _, action := range []string{"clone", "start", "checkout", "clean"} {
					arguments := append(append([]string{}, prefix...), action, "feature/42")
					if output, err := run(arguments...); err != nil || output != destination {
						t.Fatalf("%v directory = %q, %v; want %q", arguments, output, err, destination)
					}
				}
				arguments := append(append([]string{}, prefix...), "clone", "fail")
				output, err := run(arguments...)
				if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 7 || output != directory {
					t.Fatalf("failed %v = %q, %v; want exit 7 and %q", arguments, output, err, directory)
				}
			}
			if output, err := run("vek", "--version"); err != nil || output != "--version\n"+directory {
				t.Fatalf("vek --version = %q, %v", output, err)
			}
			if output, err := run("vek"); err != nil || output != directory {
				t.Fatalf("vek without arguments = %q, %v", output, err)
			}
		})
	}
}

func TestInitRejectsUnsupportedShell(t *testing.T) {
	output, err := runInit("zsh")
	if err == nil || !strings.Contains(err.Error(), "unsupported shell") || output != "" {
		t.Fatalf("init zsh = %q, %v", output, err)
	}
}

func runInit(args ...string) (string, error) {
	var output, errors bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, ErrWriter: &errors, Commands: []*cli.Command{NewCmd()}}
	err := root.Run(context.Background(), append([]string{"vek", "bonsai", "init"}, args...))
	return output.String(), err
}
