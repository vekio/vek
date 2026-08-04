package sshalias

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/vekio/vek/internal/sshalias"
)

func newCmdShow() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Aliases:   []string{},
		Usage:     "show a managed ssh-alias block by alias",
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

			sshAlias, ok := config.Get(alias)
			if !ok {
				return fmt.Errorf("alias %q not found", alias)
			}

			fmt.Fprint(c.Root().Writer, sshAlias.Block())
			return nil
		},
	}
}
