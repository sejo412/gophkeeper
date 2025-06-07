package client

import (
	"os/user"
	"path/filepath"
)

// Config main configuration for Client.
type Config struct {
	// PublicAddress to connect for register.
	PublicAddress string
	// PrivateAddress to connect for request private data.
	PrivateAddress string
	// CacheDir for save certificates.
	CacheDir string
}

// NewConfig constructs Config object.
func NewConfig() *Config {
	return &Config{
		PublicAddress:  "",
		PrivateAddress: "",
		CacheDir:       "",
	}
}

// NewConfigWithOptions constructs Config object with predefined options.
func NewConfigWithOptions(config Config) *Config {
	c := NewConfig()
	c.PublicAddress = config.PublicAddress
	c.PrivateAddress = config.PrivateAddress
	c.CacheDir = config.CacheDir
	return c
}

// DefaultCacheDir returns default cacheDir for current user.
func DefaultCacheDir() string {
	u, err := user.Current()
	if err != nil {
		return "."
	}
	return filepath.Join(u.HomeDir, ".cache/gophkeeper/client")
}
