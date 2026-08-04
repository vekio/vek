package main

import (
	"context"
	"fmt"
	"os"

	"github.com/vekio/vek/internal/cli/vek"
)

func main() {
	ctx := context.Background()

	vek := vek.NewCmd()

	if err := vek.Run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
