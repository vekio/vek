package gitignore

import (
	"context"
	"errors"
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
		Commands:  []*cli.Command{newCmdTemplates()},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Value: ".gitignore", Usage: "path to write the generated gitignore file"},
			&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "overwrite .gitignore if it already exists"},
			&cli.BoolFlag{Name: "print", Aliases: []string{"p"}, Usage: "print generated .gitignore content instead of writing it"},
		},
		ShellComplete: func(ctx context.Context, c *cli.Command) { completeTemplateArg(ctx, c) },
		Action: func(ctx context.Context, c *cli.Command) error {
			templates := strings.TrimSpace(strings.Join(c.Args().Slice(), " "))
			if templates == "" {
				return fmt.Errorf("no templates provided; try %q or run %q to see the available templates", "vek gitignore go", "vek gitignore list")
			}
			if c.Bool("print") {
				content, err := gitignorecore.Build(ctx, templates)
				if err != nil {
					return err
				}
				fmt.Fprint(c.Root().Writer, content)
				return nil
			}
			output := strings.TrimSpace(c.String("output"))
			if output == "" {
				output = ".gitignore"
			}
			if err := gitignorecore.Generate(ctx, templates, output, c.Bool("force")); errors.Is(err, gitignorecore.ErrOutputExists) {
				return fmt.Errorf("output %q already exists; use --force to overwrite it", output)
			} else if err != nil {
				return err
			}
			fmt.Fprintf(c.Root().Writer, "%s generated successfully\n", output)
			return nil
		},
	}
}
