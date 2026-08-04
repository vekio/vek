package sshalias

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Config struct {
	path    string
	content string
	aliases map[string]Alias
}

func LoadConfig() (*Config, error) {
	configPath, err := sshConfigPath()
	if err != nil {
		return nil, err
	}

	return LoadConfigPath(configPath)
}

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

func LoadConfigPath(configPath string) (*Config, error) {
	rawContent, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{
				path:    configPath,
				aliases: map[string]Alias{},
			}, nil
		}
		return nil, fmt.Errorf("read %s: %w", configPath, err)
	}

	content := string(rawContent)
	aliases, err := parseAliases(content)
	if err != nil {
		return nil, err
	}

	return &Config{
		path:    configPath,
		content: content,
		aliases: aliases,
	}, nil
}

func (c *Config) Get(alias string) (Alias, bool) {
	normalizedAlias := strings.TrimSpace(alias)
	sshAlias, ok := c.aliases[normalizedAlias]
	return sshAlias, ok
}

func (c *Config) List() []Alias {
	list := make([]Alias, 0, len(c.aliases))
	for _, sshAlias := range c.aliases {
		list = append(list, sshAlias)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].alias < list[j].alias
	})

	return list
}

func (c *Config) Add(sshAlias Alias, force bool) error {
	alias := strings.TrimSpace(sshAlias.alias)
	if alias == "" {
		return fmt.Errorf("alias is required")
	}

	if _, exists := c.aliases[alias]; exists && !force {
		return fmt.Errorf("alias %q already exists", alias)
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

func (c *Config) Delete(alias string) error {
	normalizedAlias := strings.TrimSpace(alias)
	if normalizedAlias == "" {
		return fmt.Errorf("alias is required")
	}

	if _, exists := c.aliases[normalizedAlias]; !exists {
		return fmt.Errorf("alias %q not found", normalizedAlias)
	}

	content, removed, err := removeAliasBlock(c.content, normalizedAlias)
	if err != nil {
		return err
	}
	if !removed {
		return fmt.Errorf("alias %q exists in index but block was not found in file content", normalizedAlias)
	}

	c.content = content
	delete(c.aliases, normalizedAlias)
	return nil
}

func (c *Config) Save() error {
	tmpPath := c.path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(c.content), 0o600); err != nil {
		return fmt.Errorf("write temp config %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, c.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace config %s: %w", c.path, err)
	}

	if err := os.Chmod(c.path, 0o600); err != nil {
		return fmt.Errorf("set permissions on %s: %w", c.path, err)
	}

	return nil
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
	if err != nil {
		return "", false, err
	}
	if !found {
		return content, false, nil
	}

	return content[:start] + block + content[end:], true, nil
}

func removeAliasBlock(content, alias string) (string, bool, error) {
	start, end, found, err := findAliasBlockBounds(content, alias)
	if err != nil {
		return "", false, err
	}
	if !found {
		return content, false, nil
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
