package sshalias

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	sshalias "github.com/vekio/vek/internal/ssh-alias"
)

func newCmdAdd() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "add a ssh-alias block",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "alias",
				Aliases:  []string{"a"},
				Usage:    "host alias in ~/.ssh/config",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "hostname",
				Aliases:  []string{"n"},
				Usage:    "remote host or IP address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"u"},
				Usage:    "SSH username",
				Required: true,
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   22,
				Usage:   "SSH port",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "overwrite existing ssh-alias block if present",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			alias, err := sshalias.NewAlias(
				c.String("alias"),
				c.String("hostname"),
				c.String("user"),
				c.Int("port"),
			)
			if err != nil {
				return fmt.Errorf("build ssh alias: %w", err)
			}

			config, err := sshalias.OpenOrCreateConfig()
			if err != nil {
				return fmt.Errorf("load ssh config: %w", err)
			}

			if err := config.Add(alias, c.Bool("force")); err != nil {
				return fmt.Errorf("add ssh alias: %w", err)
			}

			if err := config.Save(); err != nil {
				return fmt.Errorf("save ssh config: %w", err)
			}

			fmt.Fprintf(c.Root().Writer, "alias %q saved successfully\n", alias.Name())
			return nil
		},
	}
}
