package gitignore

import "testing"

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
