package gitignore

import (
	"reflect"
	"testing"
)

func TestTemplateCompletionNames(t *testing.T) {
	want := []string{"go", "golang", "javascript", "js", "next", "node", "nodejs", "nuxt", "opentofu", "py", "python", "react", "terraform", "tf", "tofu", "ts", "typescript", "vue"}
	if got := templateCompletionNames(); !reflect.DeepEqual(got, want) {
		t.Fatalf("templateCompletionNames() = %#v, want %#v", got, want)
	}
}

func TestCurrentTemplatePrefix(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "empty", args: nil, want: ""},
		{name: "single", args: []string{"go"}, want: "go"},
		{name: "next template", args: []string{"go", "golang"}, want: "golang"},
		{name: "partial", args: []string{"go", "t"}, want: "t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := currentTemplatePrefix(tt.args); got != tt.want {
				t.Fatalf("currentTemplatePrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}
