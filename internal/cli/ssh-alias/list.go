package sshalias

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	sshalias "github.com/vekio/vek/internal/ssh-alias"
)

func newCmdList() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "print managed ssh-alias names",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "long",
				Aliases: []string{"l"},
				Usage:   "print alias, user, hostname, and port",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			config, err := sshalias.LoadConfig()
			if err != nil {
				return fmt.Errorf("load ssh config: %w", err)
			}

			for _, alias := range config.List() {
				if c.Bool("long") {
					fmt.Fprintf(c.Root().Writer, "%s\t%s\t%s\t%d\n", alias.Name(), alias.User(), alias.Hostname(), alias.Port())
					continue
				}

				fmt.Fprintln(c.Root().Writer, alias.Name())
			}

			return nil
		},
	}
}
