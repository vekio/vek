package sshalias

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigPathDoesNotCreateMissingFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ".ssh", "config")

	config, err := LoadConfigPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigPath() error = %v", err)
	}

	if got := config.List(); len(got) != 0 {
		t.Fatalf("len(List()) = %d, want 0", len(got))
	}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("os.Stat(%q) error = %v, want not exist", configPath, err)
	}
}

func TestConfigAddReplaceDeleteAlias(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	initialContent := `Host unmanaged
	HostName unmanaged.example.com
	User root
`
	if err := os.WriteFile(configPath, []byte(initialContent), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := LoadConfigPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigPath() error = %v", err)
	}

	alias, err := NewAlias("prod", "prod.example.com", "ubuntu", 22)
	if err != nil {
		t.Fatalf("NewAlias() error = %v", err)
	}

	if err := config.Add(alias, false); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if err := config.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	updatedAlias, err := NewAlias("prod", "prod.internal", "deploy", 2222)
	if err != nil {
		t.Fatalf("NewAlias() error = %v", err)
	}

	config, err = LoadConfigPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigPath() error = %v", err)
	}

	if err := config.Add(updatedAlias, true); err != nil {
		t.Fatalf("Add(force) error = %v", err)
	}

	if err := config.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	contentBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(contentBytes)
	if !strings.Contains(content, initialContent) {
		t.Fatalf("content missing unmanaged block:\n%s", content)
	}

	if strings.Count(content, "# >>> ssh-alias prod >>>") != 1 {
		t.Fatalf("start marker count = %d, want 1:\n%s", strings.Count(content, "# >>> ssh-alias prod >>>"), content)
	}

	if !strings.Contains(content, "HostName prod.internal") {
		t.Fatalf("content missing updated hostname:\n%s", content)
	}

	if !strings.Contains(content, "User deploy") {
		t.Fatalf("content missing updated user:\n%s", content)
	}

	config, err = LoadConfigPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigPath() error = %v", err)
	}

	if err := config.Delete("prod"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if err := config.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	contentBytes, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content = string(contentBytes)
	if strings.Contains(content, "ssh-alias prod") {
		t.Fatalf("content still contains deleted alias:\n%s", content)
	}

	if !strings.Contains(content, initialContent) {
		t.Fatalf("content missing unmanaged block after delete:\n%s", content)
	}
}

func TestConfigAddExistingAliasRequiresForce(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	config, err := LoadConfigPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigPath() error = %v", err)
	}

	alias, err := NewAlias("prod", "prod.example.com", "ubuntu", 22)
	if err != nil {
		t.Fatalf("NewAlias() error = %v", err)
	}

	if err := config.Add(alias, false); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if err := config.Add(alias, false); err == nil {
		t.Fatal("Add(existing) error = nil, want error")
	}
}
