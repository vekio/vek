package gitignore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		if _, err := os.Stat(output); err == nil {
			return fmt.Errorf("%w: %s", ErrOutputExists, output)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", output, err)
		}
	}
	content, err := g.build(ctx, names)
	if err != nil {
		return err
	}
	if force {
		return replaceFile(output, content)
	}
	return createFile(output, content)
}

func createFile(output, content string) error {
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("%w: %s", ErrOutputExists, output)
	}
	if err != nil {
		return fmt.Errorf("write %s: %w", output, err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.Remove(output)
		}
	}()
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", output, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", output, err)
	}
	complete = true
	return nil
}

func replaceFile(output, content string) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(output), ".vek-gitignore-*")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", output, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if err := tmpFile.Chmod(0o644); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("set permissions on temp file for %s: %w", output, err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp file for %s: %w", output, err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("sync temp file for %s: %w", output, err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file for %s: %w", output, err)
	}
	if err := os.Rename(tmpPath, output); err != nil {
		return fmt.Errorf("replace %s: %w", output, err)
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
		builder.WriteString("# >>> ")
		builder.WriteString(displayTemplateName(template.Name))
		builder.WriteString(" <<<\n")
		builder.WriteString(strings.TrimRight(template.Content, "\r\n"))
	}
	builder.WriteByte('\n')
	return builder.String(), nil
}
