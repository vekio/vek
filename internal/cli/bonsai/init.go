package bonsai

import (
	"context"
	_ "embed"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

func newInitCmd() *cli.Command {
	return &cli.Command{
		Name:        "init",
		Usage:       "print shell integration for Bonsai",
		Description: "Load with 'vek bonsai init fish | source' in Fish or 'source <(vek bonsai init bash)' in Bash.",
		ArgsUsage:   "<shell>",
		Arguments: []cli.Argument{
			shellArg(),
		},
		Action: func(_ context.Context, c *cli.Command) error {
			var script string
			switch shell := c.StringArg("shell"); shell {
			case "bash":
				script = bashInit
			case "fish":
				script = fishInit
			default:
				return fmt.Errorf("unsupported shell %q (supported: bash, fish)", shell)
			}
			_, err := io.WriteString(c.Root().Writer, script)
			return err
		},
	}
}

//go:embed shell/bonsai.bash
var bashInit string

//go:embed shell/bonsai.fish
var fishInit string
