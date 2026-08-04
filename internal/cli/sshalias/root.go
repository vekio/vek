package sshalias

import "github.com/urfave/cli/v3"

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:    "ssh-alias",
		Aliases: []string{"sa"},
		Usage:   "manage ssh-alias blocks in ~/.ssh/config",
		Commands: []*cli.Command{
			newCmdAdd(),
			newCmdRemove(),
			newCmdList(),
			newCmdShow(),
		},
	}
}
