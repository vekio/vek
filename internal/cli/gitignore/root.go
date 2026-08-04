package gitignore

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	gitignorecore "github.com/vekio/vek/internal/gitignore"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:      "gitignore",
		Aliases:   []string{"gi"},
		Usage:     "create a .gitignore from github/gitignore templates",
		ArgsUsage: "<templates...>",
		Commands: []*cli.Command{
			newCmdTemplates(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   ".gitignore",
				Usage:   "path to write the generated gitignore file",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "overwrite .gitignore if it already exists",
			},
			&cli.BoolFlag{
				Name:    "print",
				Aliases: []string{"p"},
				Usage:   "print generated .gitignore content instead of writing it",
			},
		},
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			completeTemplateArg(ctx, c)
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			templates := strings.TrimSpace(strings.Join(c.Args().Slice(), " "))
			if templates == "" {
				return fmt.Errorf("templates argument is required")
			}

			if c.Bool("print") {
				content, err := gitignorecore.Build(ctx, templates)
				if err != nil {
					return fmt.Errorf("build gitignore: %w", err)
				}

				fmt.Fprint(c.Root().Writer, content)
				return nil
			}

			if err := gitignorecore.Generate(ctx, gitignorecore.GenerateOptions{
				Templates: templates,
				Output:    c.String("output"),
				Force:     c.Bool("force"),
			}); err != nil {
				return fmt.Errorf("generate gitignore: %w", err)
			}

			fmt.Fprintln(c.Root().Writer, ".gitignore generated successfully")
			return nil
		},
	}
}
