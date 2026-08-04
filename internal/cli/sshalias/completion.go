package sshalias

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/vekio/vek/internal/sshalias"
)

func completeAliasArg(_ context.Context, c *cli.Command) {
	config, err := sshalias.LoadConfig()
	if err != nil {
		return
	}

	prefix := strings.TrimSpace(c.Args().Get(0))

	for _, alias := range config.List() {
		name := alias.Name()
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		fmt.Fprintln(c.Writer, name)
	}
}
