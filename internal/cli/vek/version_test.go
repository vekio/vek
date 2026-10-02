package vek

import (
	"bytes"
	"context"
	"testing"
)

func TestVersionFlags(t *testing.T) {
	for _, argument := range []string{"--version", "-v"} {
		t.Run(argument, func(t *testing.T) {
			var output bytes.Buffer
			command := NewCmd()
			command.Version = "v1.2.3"
			command.Writer = &output
			if err := command.Run(context.Background(), []string{"vek", argument}); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != "vek version v1.2.3\n" {
				t.Fatalf("version output = %q", got)
			}
		})
	}
}
