package sshalias

import (
	"fmt"
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", sshDir, err)
	}

	if err := os.Chmod(sshDir, 0o700); err != nil {
		return fmt.Errorf("set permissions on %s: %w", sshDir, err)
	}

	configFile, err := os.OpenFile(configPath, os.O_RDONLY|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("create %s: %w", configPath, err)
	}
	if err := configFile.Close(); err != nil {
		return fmt.Errorf("close %s: %w", configPath, err)
	}

	if err := os.Chmod(configPath, 0o600); err != nil {
		return fmt.Errorf("set permissions on %s: %w", configPath, err)
	}

	return nil
}
