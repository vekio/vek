package gitignore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/vekio/x/file"
)

type generator struct {
	client templateClient
}

func defaultGenerator() generator {
	return generator{client: defaultTemplateClient()}
}

// Generate builds a gitignore from templates and writes it to output.
func Generate(ctx context.Context, templates, output string, force bool) error {
	return defaultGenerator().generate(ctx, templates, output, force)
}

// Build returns gitignore content assembled from the requested templates.
func Build(ctx context.Context, templates string) (string, error) {
	names, err := parseTemplateNames(templates)
	if err != nil {
		return "", err
	}
	return defaultGenerator().build(ctx, names)
}

func (g generator) generate(ctx context.Context, templates, output string, force bool) error {
	names, err := parseTemplateNames(templates)
	if err != nil {
		return err
	}
	output = strings.TrimSpace(output)
	if output == "" {
		output = defaultOutput
	}
	if !force {
		exists, err := file.Exists(output)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: %s", ErrOutputExists, output)
		}
	}
	content, err := g.build(ctx, names)
	if err != nil {
		return err
	}
	if force {
		return file.WriteAtomic(output, []byte(content), 0o644)
	}
	if err := file.WriteExclusive(output, []byte(content), 0o644); errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("%w: %s", ErrOutputExists, output)
	} else if err != nil {
		return err
	}
	return nil
}

func (g generator) build(ctx context.Context, names []string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("at least one template is required")
	}
	client := g.client
	if client == nil {
		client = defaultTemplateClient()
	}
	var builder strings.Builder
	for i, name := range names {
		template, err := client.getTemplate(ctx, name)
		if err != nil {
			return "", err
		}
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString("# >>> vek-gitignore ")
		builder.WriteString(displayTemplateName(template.Name))
		builder.WriteString(" <<<\n")
		builder.WriteString(strings.TrimRight(template.Content, "\r\n"))
	}
	builder.WriteByte('\n')
	return builder.String(), nil
}
