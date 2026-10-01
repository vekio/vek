package bonsai

import "github.com/urfave/cli/v3"

func bareOnlyFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:  "bare-only",
		Usage: "clone without creating a worktree for the default branch",
	}
}

func printPathFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:   "print-path",
		Usage:  "print only the destination path for shell integration",
		Hidden: true,
	}
}
