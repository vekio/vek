package gitignore

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"
	gitignorecore "github.com/vekio/vek/internal/gitignore"
)

func completeTemplateArg(_ context.Context, c *cli.Command) {
	prefix := currentTemplatePrefix(c.Args().Slice())
	for _, name := range templateCompletionNames() {
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		fmt.Fprintln(c.Writer, name)
	}
}

func templateCompletionNames() []string {
	seen := make(map[string]struct{})
	var names []string
	for _, template := range gitignorecore.Templates() {
		for _, alias := range template.Aliases {
			if _, exists := seen[alias]; exists {
				continue
			}
			seen[alias] = struct{}{}
			names = append(names, alias)
		}
	}
	sort.Strings(names)
	return names
}

func currentTemplatePrefix(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.TrimSpace(args[len(args)-1])
}
