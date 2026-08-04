package gitignore

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	gitignorecore "github.com/vekio/vek/internal/gitignore"
)

func newCmdTemplates() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "list available gitignore templates and aliases",
		Action: func(ctx context.Context, c *cli.Command) error {
			for _, template := range gitignorecore.ListTemplateDefinitions() {
				fmt.Fprintf(c.Root().Writer, "%s (%s)\n", template.Name, strings.Join(template.Aliases, ", "))
			}

			return nil
		},
	}
}
