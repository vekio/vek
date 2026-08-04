package sshalias

import (
	"strings"
	"testing"
)

func TestNewAliasDefaultsPort(t *testing.T) {
	alias, err := NewAlias("prod", "example.com", "ubuntu", 0)
	if err != nil {
		t.Fatalf("NewAlias() error = %v", err)
	}

	if alias.Name() != "prod" {
		t.Fatalf("Name() = %q, want %q", alias.Name(), "prod")
	}

	if alias.Port() != defaultSSHPort {
		t.Fatalf("Port() = %d, want %d", alias.Port(), defaultSSHPort)
	}
}

func TestNewAliasRejectsInvalidTokens(t *testing.T) {
	tests := []struct {
		name     string
		alias    string
		hostname string
		user     string
	}{
		{name: "alias whitespace", alias: "prod db", hostname: "example.com", user: "ubuntu"},
		{name: "hostname newline", alias: "prod", hostname: "example.com\nUser root", user: "ubuntu"},
		{name: "user tab", alias: "prod", hostname: "example.com", user: "ubuntu\troot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewAlias(tt.alias, tt.hostname, tt.user, 22); err == nil {
				t.Fatal("NewAlias() error = nil, want error")
			}
		})
	}
}

func TestAliasBlockUsesSSHAliasMarkers(t *testing.T) {
	alias, err := NewAlias("prod", "example.com", "ubuntu", 2222)
	if err != nil {
		t.Fatalf("NewAlias() error = %v", err)
	}

	block := alias.Block()
	if !strings.Contains(block, "# >>> vek-ssh-alias prod >>>") {
		t.Fatalf("Block() missing ssh-alias start marker:\n%s", block)
	}

	if !strings.Contains(block, "# <<< vek-ssh-alias prod <<<") {
		t.Fatalf("Block() missing ssh-alias end marker:\n%s", block)
	}

	if strings.Contains(block, "ssh-entry") {
		t.Fatalf("Block() contains old ssh-entry marker:\n%s", block)
	}
}
