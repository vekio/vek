//go:build !darwin && !dragonfly && !freebsd && !illumos && !linux && !netbsd && !openbsd

package sshalias

import "sync"

var fallbackConfigLock sync.Mutex

func lockConfig(string) (func() error, error) {
	fallbackConfigLock.Lock()
	return func() error {
		fallbackConfigLock.Unlock()
		return nil
	}, nil
}
