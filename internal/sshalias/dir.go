package sshalias

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/vekio/x/file"
)

func sshConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	return filepath.Join(home, ".ssh", "config"), nil
}

func setupSSHConfig(configPath string) error {
	sshDir := filepath.Dir(configPath)
	if err := file.EnsureParentDir(configPath, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", sshDir, err)
	}

	if err := os.Chmod(sshDir, 0o700); err != nil {
		return fmt.Errorf("set permissions on %s: %w", sshDir, err)
	}

	if err := file.WriteExclusive(configPath, nil, 0o600); err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("create %s: %w", configPath, err)
	}

	if err := os.Chmod(configPath, 0o600); err != nil {
		return fmt.Errorf("set permissions on %s: %w", configPath, err)
	}

	return nil
}
