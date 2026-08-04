package gitignore

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type TemplateDefinition struct {
	Name      string
	Canonical string
	Aliases   []string
}

func ParseTemplateNames(input string) ([]string, error) {
	parts := strings.Fields(input)
	names := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		name := NormalizeTemplateName(part)
		if name == "" {
			continue
		}

		if err := validateTemplateName(name); err != nil {
			return nil, err
		}

		if !isKnownTemplateName(name) {
			return nil, fmt.Errorf("template name %q is not available", part)
		}

		if _, exists := seen[name]; exists {
			continue
		}

		seen[name] = struct{}{}
		names = append(names, name)
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("at least one template is required")
	}

	return names, nil
}

func NormalizeTemplateName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimSuffix(name, ".gitignore")
	if name == "" {
		return ""
	}

	if normalized, ok := templateAliases[strings.ToLower(name)]; ok {
		return normalized
	}

	return uppercaseFirst(name)
}

func DisplayTemplateName(name string) string {
	name = NormalizeTemplateName(name)
	for _, definition := range templateDefinitions {
		if definition.Canonical == name {
			return definition.Name
		}
	}

	return strings.ToLower(name)
}

func ListTemplateDefinitions() []TemplateDefinition {
	definitions := make([]TemplateDefinition, len(templateDefinitions))
	copy(definitions, templateDefinitions)

	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].Name < definitions[j].Name
	})

	return definitions
}

func ListTemplateCompletionNames() []string {
	seen := make(map[string]struct{})
	names := make([]string, 0, len(templateDefinitions))

	for _, definition := range templateDefinitions {
		for _, alias := range definition.Aliases {
			if _, exists := seen[alias]; exists {
				continue
			}

			seen[alias] = struct{}{}
			names = append(names, alias)
		}
	}

	sort.Strings(names)
	return names
}

func validateTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("template name is required")
	}

	for _, r := range name {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return fmt.Errorf("template name %q must not contain whitespace or control characters", name)
		}
	}

	if strings.Contains(name, ",") {
		return fmt.Errorf("template name %q must be separated with spaces, not commas", name)
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return fmt.Errorf("template name %q must be a root github/gitignore template", name)
	}

	return nil
}

func uppercaseFirst(value string) string {
	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

var templateDefinitions = []TemplateDefinition{
	{Name: "go", Canonical: "Go", Aliases: []string{"go", "golang"}},
	{Name: "node", Canonical: "Node", Aliases: []string{"node", "nodejs", "javascript", "js", "typescript", "ts", "vue", "react", "next", "nuxt"}},
}

var templateAliases = buildTemplateAliases()

func isKnownTemplateName(name string) bool {
	for _, definition := range templateDefinitions {
		if definition.Canonical == name {
			return true
		}
	}

	return false
}

func buildTemplateAliases() map[string]string {
	aliases := make(map[string]string)
	for _, definition := range templateDefinitions {
		for _, alias := range definition.Aliases {
			aliases[alias] = definition.Canonical
		}
	}

	return aliases
}
