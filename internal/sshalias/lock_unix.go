//go:build darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd

package sshalias

import (
	"fmt"
	"os"
	"syscall"
)

func lockConfig(configPath string) (func() error, error) {
	lockFile, err := os.OpenFile(configPath+".vek.lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open config lock for %s: %w", configPath, err)
	}
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		_ = lockFile.Close()
		return nil, fmt.Errorf("lock config %s: %w", configPath, err)
	}
	return func() error {
		unlockErr := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		closeErr := lockFile.Close()
		if unlockErr != nil {
			return fmt.Errorf("unlock config %s: %w", configPath, unlockErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close config lock for %s: %w", configPath, closeErr)
		}
		return nil
	}, nil
}
