package gitignore

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"
)

// TemplateDefinition describes an available template and its accepted names.
type TemplateDefinition struct {
	Name    string
	Aliases []string
}

type templateDefinition struct {
	name      string
	canonical string
	aliases   []string
}

func parseTemplateNames(input string) ([]string, error) {
	parts := strings.Fields(input)
	names := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		name := normalizeTemplateName(part)
		if err := validateTemplateName(name); err != nil {
			return nil, err
		}
		if !isKnownTemplateName(name) {
			return nil, fmt.Errorf("template %q is not supported; choose from go, node, python, or terraform", part)
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

func normalizeTemplateName(name string) string {
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

func displayTemplateName(name string) string {
	name = normalizeTemplateName(name)
	for _, definition := range templateDefinitions {
		if definition.canonical == name {
			return definition.name
		}
	}
	return strings.ToLower(name)
}

// Templates returns the available templates and their accepted aliases.
func Templates() []TemplateDefinition {
	definitions := make([]TemplateDefinition, len(templateDefinitions))
	for i, definition := range templateDefinitions {
		definitions[i] = TemplateDefinition{Name: definition.name, Aliases: slices.Clone(definition.aliases)}
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	return definitions
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

var templateDefinitions = []templateDefinition{
	{name: "go", canonical: "Go", aliases: []string{"go", "golang"}},
	{name: "node", canonical: "Node", aliases: []string{"node", "nodejs", "javascript", "js", "typescript", "ts", "vue", "react", "next", "nuxt"}},
	{name: "python", canonical: "Python", aliases: []string{"python", "py"}},
	{name: "terraform", canonical: "Terraform", aliases: []string{"terraform", "tf", "opentofu", "tofu"}},
}

var templateAliases = buildTemplateAliases()

func isKnownTemplateName(name string) bool {
	for _, definition := range templateDefinitions {
		if definition.canonical == name {
			return true
		}
	}
	return false
}

func buildTemplateAliases() map[string]string {
	aliases := make(map[string]string)
	for _, definition := range templateDefinitions {
		for _, alias := range definition.aliases {
			aliases[alias] = definition.canonical
		}
	}
	return aliases
}
