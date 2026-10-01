package bonsai

import "github.com/urfave/cli/v3"

func repoURLArg() cli.Argument {
	return &cli.StringArg{
		Name:     "repoURL",
		Required: true,
	}
}

func pathArg() cli.Argument {
	return &cli.StringArg{
		Name: "path",
	}
}

func taskNameArg() cli.Argument {
	return &cli.StringArg{
		Name:     "task-name",
		Required: true,
	}
}

func shellArg() cli.Argument {
	return &cli.StringArg{
		Name:     "shell",
		Required: true,
	}
}
