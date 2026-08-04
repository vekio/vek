package sshalias

import "testing"

func TestParseAliasesReadsManagedBlocksOnly(t *testing.T) {
	content := `Host unmanaged
	HostName unmanaged.example.com
	User root

# >>> ssh-alias prod >>>
Host prod
	HostName prod.example.com
	User ubuntu
	Port 2222
# <<< ssh-alias prod <<<
`

	aliases, err := parseAliases(content)
	if err != nil {
		t.Fatalf("parseAliases() error = %v", err)
	}

	if len(aliases) != 1 {
		t.Fatalf("len(aliases) = %d, want 1", len(aliases))
	}

	alias := aliases["prod"]
	if alias.Hostname() != "prod.example.com" {
		t.Fatalf("Hostname() = %q, want %q", alias.Hostname(), "prod.example.com")
	}

	if alias.User() != "ubuntu" {
		t.Fatalf("User() = %q, want %q", alias.User(), "ubuntu")
	}

	if alias.Port() != 2222 {
		t.Fatalf("Port() = %d, want 2222", alias.Port())
	}
}

func TestParseAliasesIgnoresOldSSHEntryBlocks(t *testing.T) {
	content := `# >>> ssh-entry old >>>
Host old
	HostName old.example.com
	User ubuntu
	Port 22
# <<< ssh-entry old <<<
`

	aliases, err := parseAliases(content)
	if err != nil {
		t.Fatalf("parseAliases() error = %v", err)
	}

	if len(aliases) != 0 {
		t.Fatalf("len(aliases) = %d, want 0", len(aliases))
	}
}

func TestParseAliasesRejectsCorruptBlocks(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "unterminated",
			content: `# >>> ssh-alias prod >>>
Host prod
	HostName prod.example.com
	User ubuntu
	Port 22
`,
		},
		{
			name: "mismatched markers",
			content: `# >>> ssh-alias prod >>>
Host prod
	HostName prod.example.com
	User ubuntu
	Port 22
# <<< ssh-alias other <<<
`,
		},
		{
			name: "host differs from marker",
			content: `# >>> ssh-alias prod >>>
Host other
	HostName prod.example.com
	User ubuntu
	Port 22
# <<< ssh-alias prod <<<
`,
		},
		{
			name: "duplicate alias",
			content: `# >>> ssh-alias prod >>>
Host prod
	HostName prod.example.com
	User ubuntu
	Port 22
# <<< ssh-alias prod <<<
# >>> ssh-alias prod >>>
Host prod
	HostName other.example.com
	User ubuntu
	Port 22
# <<< ssh-alias prod <<<
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseAliases(tt.content); err == nil {
				t.Fatal("parseAliases() error = nil, want error")
			}
		})
	}
}
