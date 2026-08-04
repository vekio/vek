package sshalias

import (
	"errors"
	"fmt"
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

	contentBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(contentBytes)
	if !strings.Contains(content, initialContent) {
		t.Fatalf("content missing unmanaged block:\n%s", content)
	}

	if strings.Count(content, "# >>> vek-ssh-alias prod >>>") != 1 {
		t.Fatalf("start marker count = %d, want 1:\n%s", strings.Count(content, "# >>> vek-ssh-alias prod >>>"), content)
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

	contentBytes, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content = string(contentBytes)
	if strings.Contains(content, "vek-ssh-alias prod") {
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

	if err := config.Add(alias, false); !errors.Is(err, ErrAliasExists) {
		t.Fatalf("Add(existing) error = %v, want ErrAliasExists", err)
	}
}

func TestConfigDeleteMissingAliasReturnsTypedError(t *testing.T) {
	config, err := LoadConfigPath(filepath.Join(t.TempDir(), "config"))
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Delete("missing"); !errors.Is(err, ErrAliasNotFound) {
		t.Fatalf("Delete() error = %v, want ErrAliasNotFound", err)
	}
}

func TestConfigRejectsChangesSinceLoad(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(configPath, []byte("Host unmanaged\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	first, _ := LoadConfigPath(configPath)
	second, _ := LoadConfigPath(configPath)
	firstAlias, _ := NewAlias("first", "first.example.com", "ubuntu", 22)
	if err := first.Add(firstAlias, false); err != nil {
		t.Fatal(err)
	}
	secondAlias, _ := NewAlias("second", "second.example.com", "ubuntu", 22)
	if err := second.Add(secondAlias, false); !errors.Is(err, ErrConfigChanged) {
		t.Fatalf("Add() error = %v, want ErrConfigChanged", err)
	}
	if _, exists := second.Get("second"); exists {
		t.Fatal("failed Add() did not roll back in-memory alias")
	}
}

func TestConfigSerializesConcurrentWriters(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(configPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	configs := make([]*Config, 2)
	aliases := make([]Alias, 2)
	for i := range configs {
		configs[i], _ = LoadConfigPath(configPath)
		aliases[i], _ = NewAlias(fmt.Sprintf("host-%d", i), fmt.Sprintf("host-%d.example.com", i), "ubuntu", 22)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for i, config := range configs {
		go func(alias Alias) {
			<-start
			results <- config.Add(alias, false)
		}(aliases[i])
	}
	close(start)
	succeeded, conflicted := 0, 0
	for range configs {
		err := <-results
		if err == nil {
			succeeded++
		} else if errors.Is(err, ErrConfigChanged) {
			conflicted++
		} else {
			t.Fatalf("Add() unexpected error = %v", err)
		}
	}
	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("concurrent adds: succeeded = %d, conflicted = %d", succeeded, conflicted)
	}
}

func TestConfigDoesNotReusePredictableTempPath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")
	predictableTempPath := configPath + ".tmp"
	if err := os.WriteFile(predictableTempPath, []byte("unrelated\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, _ := LoadConfigPath(configPath)
	alias, _ := NewAlias("prod", "prod.example.com", "ubuntu", 22)
	if err := config.Add(alias, false); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(predictableTempPath)
	if err != nil || string(content) != "unrelated\n" {
		t.Fatalf("temporary neighbor content = %q, error = %v", content, err)
	}
}
