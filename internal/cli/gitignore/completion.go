package gitignore

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	gitignorecore "github.com/vekio/vek/internal/gitignore"
)

func completeTemplateArg(_ context.Context, c *cli.Command) {
	prefix := currentTemplatePrefix(c.Args().Slice())

	for _, name := range gitignorecore.ListTemplateCompletionNames() {
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}

		fmt.Fprintln(c.Writer, name)
	}
}

func currentTemplatePrefix(args []string) string {
	if len(args) == 0 {
		return ""
	}

	return strings.TrimSpace(args[len(args)-1])
}
