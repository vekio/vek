package gitignore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Generator struct {
	Client TemplateClient
}

func DefaultGenerator() Generator {
	return Generator{
		Client: DefaultTemplateClient(),
	}
}

func Generate(ctx context.Context, options GenerateOptions) error {
	generator := DefaultGenerator()
	return generator.Generate(ctx, options)
}

func Build(ctx context.Context, templates string) (string, error) {
	generator := DefaultGenerator()
	return generator.BuildFromInput(ctx, templates)
}

func (g Generator) Generate(ctx context.Context, options GenerateOptions) error {
	names, err := ParseTemplateNames(options.Templates)
	if err != nil {
		return err
	}

	output := strings.TrimSpace(options.Output)
	if output == "" {
		output = defaultOutput
	}

	if !options.Force {
		if _, err := os.Stat(output); err == nil {
			return fmt.Errorf("%w: %s", ErrOutputExists, output)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", output, err)
		}
	}

	content, err := g.Build(ctx, names)
	if err != nil {
		return err
	}

	if err := os.WriteFile(output, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", output, err)
	}

	return nil
}

func (g Generator) BuildFromInput(ctx context.Context, templates string) (string, error) {
	names, err := ParseTemplateNames(templates)
	if err != nil {
		return "", err
	}

	return g.Build(ctx, names)
}

func (g Generator) Build(ctx context.Context, names []string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("at least one template is required")
	}

	client := g.Client
	if client == nil {
		client = DefaultTemplateClient()
	}

	var builder strings.Builder
	for i, name := range names {
		template, err := client.GetTemplate(ctx, name)
		if err != nil {
			return "", err
		}

		if i > 0 {
			builder.WriteString("\n\n")
		}

		builder.WriteString("# >>> ")
		builder.WriteString(DisplayTemplateName(template.Name))
		builder.WriteString(" <<<\n")
		builder.WriteString(strings.TrimRight(template.Content, "\r\n"))
	}

	return builder.String(), nil
}
