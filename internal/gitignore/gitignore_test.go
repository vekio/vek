package gitignore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type mapTemplateClient map[string]string

func (c mapTemplateClient) GetTemplate(_ context.Context, name string) (Template, error) {
	content, ok := c[name]
	if !ok {
		return Template{}, fmt.Errorf("template %q not found", name)
	}

	return Template{
		Name:    name,
		Content: content,
	}, nil
}

func TestParseTemplateNamesNormalizesAndDeduplicates(t *testing.T) {
	names, err := ParseTemplateNames(" go golang Go node ts vue react ")
	if err != nil {
		t.Fatalf("ParseTemplateNames() error = %v", err)
	}

	want := []string{"Go", "Node"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("ParseTemplateNames() = %#v, want %#v", names, want)
	}
}

func TestParseTemplateNamesRejectsInvalidNames(t *testing.T) {
	tests := []string{
		"",
		"go,node",
		"../Go",
		"Global/macOS",
		"python",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseTemplateNames(input); err == nil {
				t.Fatal("ParseTemplateNames() error = nil, want error")
			}
		})
	}
}

func TestDisplayTemplateName(t *testing.T) {
	tests := map[string]string{
		"Go":         "go",
		"go":         "go",
		"golang":     "go",
		"Node":       "node",
		"node":       "node",
		"typescript": "node",
		"ts":         "node",
		"vue":        "node",
		"react":      "node",
		"next":       "node",
		"nuxt":       "node",
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			if got := DisplayTemplateName(input); got != want {
				t.Fatalf("DisplayTemplateName() = %q, want %q", got, want)
			}
		})
	}
}

func TestListTemplateDefinitionsIncludesAliases(t *testing.T) {
	definitions := ListTemplateDefinitions()
	byName := make(map[string]TemplateDefinition, len(definitions))
	for _, definition := range definitions {
		byName[definition.Name] = definition
	}

	if len(definitions) != 2 {
		t.Fatalf("len(definitions) = %d, want 2", len(definitions))
	}

	goTemplate := byName["go"]
	if !reflect.DeepEqual(goTemplate.Aliases, []string{"go", "golang"}) {
		t.Fatalf("go aliases = %#v", goTemplate.Aliases)
	}

	nodeTemplate := byName["node"]
	wantNodeAliases := []string{"node", "nodejs", "javascript", "js", "typescript", "ts", "vue", "react", "next", "nuxt"}
	if !reflect.DeepEqual(nodeTemplate.Aliases, wantNodeAliases) {
		t.Fatalf("node aliases = %#v, want %#v", nodeTemplate.Aliases, wantNodeAliases)
	}
}

func TestListTemplateCompletionNamesIncludesAliases(t *testing.T) {
	names := ListTemplateCompletionNames()
	want := []string{"go", "golang", "javascript", "js", "next", "node", "nodejs", "nuxt", "react", "ts", "typescript", "vue"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("ListTemplateCompletionNames() = %#v, want %#v", names, want)
	}
}

func TestHTTPClientGetTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Go.gitignore" {
			http.NotFound(w, r)
			return
		}

		_, _ = w.Write([]byte("bin/\n*.test\n"))
	}))
	defer server.Close()

	client := HTTPClient{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	template, err := client.GetTemplate(context.Background(), "go")
	if err != nil {
		t.Fatalf("GetTemplate() error = %v", err)
	}

	if template.Name != "Go" {
		t.Fatalf("Name = %q, want %q", template.Name, "Go")
	}

	if template.Content != "bin/\n*.test\n" {
		t.Fatalf("Content = %q", template.Content)
	}
}

func TestHTTPClientGetTemplateNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	client := HTTPClient{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	if _, err := client.GetTemplate(context.Background(), "Go"); err == nil {
		t.Fatal("GetTemplate() error = nil, want error")
	}
}

func TestGeneratorBuildCombinesTemplatesWithMarkers(t *testing.T) {
	generator := Generator{
		Client: mapTemplateClient{
			"Go":   "bin/\n",
			"Node": "node_modules/\n",
		},
	}

	content, err := generator.Build(context.Background(), []string{"Go", "Node"})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	for _, expected := range []string{
		"# >>> go <<<",
		"bin/",
		"# >>> node <<<",
		"node_modules/",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("Build() missing %q:\n%s", expected, content)
		}
	}

	if strings.Contains(content, "vek gitignore") {
		t.Fatalf("Build() contains the vek gitignore label:\n%s", content)
	}

	if strings.Contains(content, "# <<<") {
		t.Fatalf("Build() contains a closing marker:\n%s", content)
	}
}

func TestGeneratorBuildFromInput(t *testing.T) {
	generator := Generator{
		Client: mapTemplateClient{
			"Go":   "bin/\n",
			"Node": "node_modules/\n",
		},
	}

	content, err := generator.BuildFromInput(context.Background(), "go vue")
	if err != nil {
		t.Fatalf("BuildFromInput() error = %v", err)
	}

	for _, expected := range []string{"# >>> go <<<", "bin/", "# >>> node <<<", "node_modules/"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("BuildFromInput() missing %q:\n%s", expected, content)
		}
	}
}

func TestGeneratorGenerateWritesOutputAndHonorsForce(t *testing.T) {
	generator := Generator{
		Client: mapTemplateClient{
			"Go": "bin/\n",
		},
	}

	output := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(output, []byte("existing\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := generator.Generate(context.Background(), GenerateOptions{
		Templates: "go",
		Output:    output,
	})
	if !errors.Is(err, ErrOutputExists) {
		t.Fatalf("Generate() error = %v, want ErrOutputExists", err)
	}

	if err := generator.Generate(context.Background(), GenerateOptions{
		Templates: "go",
		Output:    output,
		Force:     true,
	}); err != nil {
		t.Fatalf("Generate(force) error = %v", err)
	}

	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(content), "bin/") {
		t.Fatalf("generated file missing template content:\n%s", string(content))
	}
}

func TestGeneratorGenerateWritesCustomOutput(t *testing.T) {
	generator := Generator{
		Client: mapTemplateClient{
			"Go": "bin/\n",
		},
	}

	output := filepath.Join(t.TempDir(), ".gitignore.local")
	if err := generator.Generate(context.Background(), GenerateOptions{
		Templates: "go",
		Output:    output,
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(content), "bin/") {
		t.Fatalf("custom output missing template content:\n%s", string(content))
	}
}
