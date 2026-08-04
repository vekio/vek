package gitignore

import (
	"context"
	"errors"
)

const defaultOutput = ".gitignore"

var (
	// ErrOutputExists indicates that Generate would overwrite a file without force.
	ErrOutputExists = errors.New("output file already exists")
	// ErrTemplateNotFound indicates that GitHub does not have the requested template.
	ErrTemplateNotFound = errors.New("gitignore template not found")
)

type templateClient interface {
	getTemplate(ctx context.Context, name string) (templateContent, error)
}

type templateContent struct {
	Name    string
	Content string
}
