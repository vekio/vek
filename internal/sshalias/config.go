package sshalias

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	// ErrAliasExists indicates that an alias cannot be added without force.
	ErrAliasExists = errors.New("ssh alias already exists")
	// ErrAliasNotFound indicates that a requested managed alias does not exist.
	ErrAliasNotFound = errors.New("ssh alias not found")
	// ErrConfigChanged indicates that the config changed since it was loaded.
	ErrConfigChanged = errors.New("ssh config changed since it was loaded")
)

// Config is an in-memory view of an OpenSSH config and its managed aliases.
type Config struct {
	path            string
	content         string
	originalContent string
	aliases         map[string]Alias
}

// LoadConfig loads the user's SSH config without creating it.
func LoadConfig() (*Config, error) {
	configPath, err := sshConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadConfigPath(configPath)
}

// OpenOrCreateConfig prepares and loads the user's SSH config with secure permissions.
func OpenOrCreateConfig() (*Config, error) {
	configPath, err := sshConfigPath()
	if err != nil {
		return nil, err
	}
	if err := setupSSHConfig(configPath); err != nil {
		return nil, fmt.Errorf("setup ssh config: %w", err)
	}
	return LoadConfigPath(configPath)
}

// LoadConfigPath loads an SSH config from configPath without creating it.
func LoadConfigPath(configPath string) (*Config, error) {
	content, err := readConfigContent(configPath)
	if err != nil {
		return nil, err
	}
	aliases, err := parseAliases(content)
	if err != nil {
		return nil, err
	}
	return &Config{
		path:            configPath,
		content:         content,
		originalContent: content,
		aliases:         aliases,
	}, nil
}

// Get returns a managed alias by name.
func (c *Config) Get(alias string) (Alias, bool) {
	sshAlias, ok := c.aliases[strings.TrimSpace(alias)]
	return sshAlias, ok
}

// List returns all managed aliases sorted by name.
func (c *Config) List() []Alias {
	list := make([]Alias, 0, len(c.aliases))
	for _, sshAlias := range c.aliases {
		list = append(list, sshAlias)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].alias < list[j].alias })
	return list
}

// Add adds or replaces a managed alias and persists the config immediately.
func (c *Config) Add(sshAlias Alias, force bool) error {
	return c.update(func() error { return c.add(sshAlias, force) })
}

func (c *Config) add(sshAlias Alias, force bool) error {
	alias := strings.TrimSpace(sshAlias.alias)
	if alias == "" {
		return fmt.Errorf("alias is required")
	}
	if _, exists := c.aliases[alias]; exists && !force {
		return fmt.Errorf("%w: %q", ErrAliasExists, alias)
	}

	block := sshAlias.Block()
	if _, exists := c.aliases[alias]; exists {
		content, replaced, err := replaceAliasBlock(c.content, alias, block)
		if err != nil {
			return err
		}
		if !replaced {
			return fmt.Errorf("alias %q exists in index but block was not found in file content", alias)
		}
		c.content = content
	} else {
		c.content = appendAliasBlock(c.content, block)
	}
	c.aliases[alias] = sshAlias
	return nil
}

// Delete removes a managed alias and persists the config immediately.
func (c *Config) Delete(alias string) error {
	return c.update(func() error { return c.delete(alias) })
}

func (c *Config) delete(alias string) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return fmt.Errorf("alias is required")
	}
	if _, exists := c.aliases[alias]; !exists {
		return fmt.Errorf("%w: %q", ErrAliasNotFound, alias)
	}
	content, removed, err := removeAliasBlock(c.content, alias)
	if err != nil {
		return err
	}
	if !removed {
		return fmt.Errorf("alias %q exists in index but block was not found in file content", alias)
	}
	c.content = content
	delete(c.aliases, alias)
	return nil
}

func (c *Config) update(change func() error) error {
	content := c.content
	aliases := cloneAliases(c.aliases)
	if err := change(); err != nil {
		return err
	}
	if err := c.save(); err != nil {
		c.content = content
		c.aliases = aliases
		return err
	}
	return nil
}

func (c *Config) save() error {
	unlock, err := lockConfig(c.path)
	if err != nil {
		return err
	}
	defer unlock()

	currentContent, err := readConfigContent(c.path)
	if err != nil {
		return err
	}
	if currentContent != c.originalContent {
		return fmt.Errorf("%w: %s", ErrConfigChanged, c.path)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(c.path), ".vek-ssh-config-*")
	if err != nil {
		return fmt.Errorf("create temp config for %s: %w", c.path, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if err := tmpFile.Chmod(0o600); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("set permissions on temp config for %s: %w", c.path, err)
	}
	if _, err := tmpFile.WriteString(c.content); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp config for %s: %w", c.path, err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("sync temp config for %s: %w", c.path, err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp config for %s: %w", c.path, err)
	}
	if err := os.Rename(tmpPath, c.path); err != nil {
		return fmt.Errorf("replace config %s: %w", c.path, err)
	}
	c.originalContent = c.content
	return nil
}

func readConfigContent(configPath string) (string, error) {
	content, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read %s: %w", configPath, err)
	}
	return string(content), nil
}

func cloneAliases(aliases map[string]Alias) map[string]Alias {
	clone := make(map[string]Alias, len(aliases))
	maps.Copy(clone, aliases)
	return clone
}

func appendAliasBlock(content, block string) string {
	if content == "" {
		return block
	}
	if strings.HasSuffix(content, "\n") {
		return content + block
	}
	return content + "\n" + block
}

func replaceAliasBlock(content, alias, block string) (string, bool, error) {
	start, end, found, err := findAliasBlockBounds(content, alias)
	if err != nil || !found {
		return content, found, err
	}
	return content[:start] + block + content[end:], true, nil
}

func removeAliasBlock(content, alias string) (string, bool, error) {
	start, end, found, err := findAliasBlockBounds(content, alias)
	if err != nil || !found {
		return content, found, err
	}
	return content[:start] + content[end:], true, nil
}

func findAliasBlockBounds(content, alias string) (int, int, bool, error) {
	startMarker := fmt.Sprintf("%s%s%s", aliasStartPrefix, alias, aliasStartSuffix)
	endMarker := fmt.Sprintf("%s%s%s", aliasEndPrefix, alias, aliasEndSuffix)
	start := strings.Index(content, startMarker)
	if start == -1 {
		return 0, 0, false, nil
	}
	endMarkerRelative := strings.Index(content[start:], endMarker)
	if endMarkerRelative == -1 {
		return 0, 0, false, fmt.Errorf("alias %q has start marker but no end marker", alias)
	}
	end := start + endMarkerRelative + len(endMarker)
	if end < len(content) && content[end] == '\n' {
		end++
	}
	return start, end, true, nil
}
