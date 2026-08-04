package sshalias

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestCommandsLifecycle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	output := runCommand(t, "ssh-alias", "add", "--alias", "prod", "--hostname", "prod.example.com", "--user", "deploy", "--port", "2222")
	if !strings.Contains(output, `alias "prod" saved successfully`) {
		t.Fatalf("add output = %q", output)
	}
	if output = runCommand(t, "ssh-alias", "list", "--long"); output != "prod\tdeploy\tprod.example.com\t2222\n" {
		t.Fatalf("list output = %q", output)
	}
	output = runCommand(t, "ssh-alias", "show", "prod")
	if !strings.Contains(output, "Host prod\n") || !strings.Contains(output, "\tPort 2222\n") {
		t.Fatalf("show output = %q", output)
	}
	runCommand(t, "ssh-alias", "remove", "prod")
	content, err := os.ReadFile(filepath.Join(home, ".ssh", "config"))
	if err != nil || strings.Contains(string(content), "vek-ssh-alias prod") {
		t.Fatalf("config after remove = %q, error = %v", content, err)
	}
}

func runCommand(t *testing.T, args ...string) string {
	t.Helper()
	var output bytes.Buffer
	root := &cli.Command{Name: "vek", Writer: &output, Commands: []*cli.Command{NewCmd()}}
	if err := root.Run(context.Background(), append([]string{"vek"}, args...)); err != nil {
		t.Fatalf("Run(%q) error = %v", args, err)
	}
	return output.String()
}
