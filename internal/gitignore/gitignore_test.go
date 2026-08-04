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

func (c mapTemplateClient) getTemplate(_ context.Context, name string) (templateContent, error) {
	content, ok := c[name]
	if !ok {
		return templateContent{}, fmt.Errorf("template %q not found", name)
	}
	return templateContent{Name: name, Content: content}, nil
}

func TestParseTemplateNamesNormalizesAndDeduplicates(t *testing.T) {
	names, err := parseTemplateNames("go golang node ts python py terraform tofu")
	if err != nil {
		t.Fatalf("parseTemplateNames() error = %v", err)
	}
	want := []string{"Go", "Node", "Python", "Terraform"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("parseTemplateNames() = %#v, want %#v", names, want)
	}
}

func TestParseTemplateNamesRejectsInvalidNames(t *testing.T) {
	for _, input := range []string{"", "go,node", "../Go", "Global/macOS", "java"} {
		t.Run(input, func(t *testing.T) {
			if _, err := parseTemplateNames(input); err == nil {
				t.Fatal("parseTemplateNames() error = nil, want error")
			}
		})
	}
}

func TestTemplates(t *testing.T) {
	definitions := Templates()
	if len(definitions) != 4 {
		t.Fatalf("len(Templates()) = %d, want 4", len(definitions))
	}
	if got := definitions[3]; got.Name != "terraform" || !reflect.DeepEqual(got.Aliases, []string{"terraform", "tf", "opentofu", "tofu"}) {
		t.Fatalf("terraform definition = %#v", got)
	}
}

func TestHTTPTemplateClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Go.gitignore" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("bin/\n*.test\n"))
	}))
	defer server.Close()
	client := httpTemplateClient{baseURL: server.URL, httpClient: server.Client()}
	template, err := client.getTemplate(context.Background(), "go")
	if err != nil {
		t.Fatalf("getTemplate() error = %v", err)
	}
	if template.Name != "Go" || template.Content != "bin/\n*.test\n" {
		t.Fatalf("getTemplate() = %#v", template)
	}
}

func TestHTTPTemplateClientNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	client := httpTemplateClient{baseURL: server.URL, httpClient: server.Client()}
	if _, err := client.getTemplate(context.Background(), "Go"); !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("getTemplate() error = %v, want ErrTemplateNotFound", err)
	}
}

func TestGeneratorBuild(t *testing.T) {
	g := generator{client: mapTemplateClient{"Go": "bin/\n", "Node": "node_modules/\n"}}
	content, err := g.build(context.Background(), []string{"Go", "Node"})
	if err != nil {
		t.Fatalf("build() error = %v", err)
	}
	for _, expected := range []string{"# >>> vek-gitignore go <<<", "bin/", "# >>> vek-gitignore node <<<", "node_modules/"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("build() missing %q:\n%s", expected, content)
		}
	}
	if !strings.HasSuffix(content, "\n") {
		t.Fatalf("build() content does not end with newline: %q", content)
	}
}

func TestGeneratorGenerateHonorsForce(t *testing.T) {
	g := generator{client: mapTemplateClient{"Go": "bin/\n"}}
	output := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(output, []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := g.generate(context.Background(), "go", output, false); !errors.Is(err, ErrOutputExists) {
		t.Fatalf("generate() error = %v, want ErrOutputExists", err)
	}
	if err := g.generate(context.Background(), "go", output, true); err != nil {
		t.Fatalf("generate(force) error = %v", err)
	}
	content, err := os.ReadFile(output)
	if err != nil || !strings.Contains(string(content), "bin/") {
		t.Fatalf("generated content = %q, error = %v", content, err)
	}
}
