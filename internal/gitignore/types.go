package gitignore

import (
	"context"
	"errors"
)

const defaultOutput = ".gitignore"

var ErrOutputExists = errors.New("output file already exists")

type TemplateClient interface {
	GetTemplate(ctx context.Context, name string) (Template, error)
}

type Template struct {
	Name    string
	Content string
}

type GenerateOptions struct {
	Templates string
	Output    string
	Force     bool
}
