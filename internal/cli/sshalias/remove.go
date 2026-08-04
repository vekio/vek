package sshalias

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/vekio/vek/internal/sshalias"
)

func newCmdRemove() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "remove a managed ssh-alias block",
		ArgsUsage: "<alias>",
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			completeAliasArg(ctx, c)
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			alias := strings.TrimSpace(c.Args().Get(0))
			if alias == "" {
				return fmt.Errorf("alias argument is required")
			}

			config, err := sshalias.LoadConfig()
			if err != nil {
				return fmt.Errorf("load ssh config: %w", err)
			}

			if err := config.Delete(alias); err != nil {
				return fmt.Errorf("delete ssh alias: %w", err)
			}

			fmt.Fprintf(c.Root().Writer, "alias %q deleted successfully\n", alias)
			return nil
		},
	}
}
