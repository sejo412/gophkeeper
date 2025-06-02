package client

import (
	"os/user"
	"path/filepath"
)

func DefaultCacheDir() string {
	u, err := user.Current()
	if err != nil {
		return "."
	}
	return filepath.Join(u.HomeDir, ".cache/gophkeeper/client")
}
