package server

import (
	"os/user"
	"path/filepath"
)

// Config main settings for Server.
type Config struct {
	// PublicPort port for listen No-TLS connections for register new Clients.
	PublicPort int
	// PrivatePort port for listen TLS connection for registered Clients.
	PrivatePort int
	// CacheDir os dir for save embedded storage files and Server's certificates.
	CacheDir string
	// DNSNames public server names for Server's certificate.
	DNSNames []string
	// Storage used store.
	Storage Storage
}

// NewConfig constructs new Config object.
func NewConfig() *Config {
	return &Config{
		PublicPort:  -1,
		PrivatePort: -1,
		CacheDir:    "",
		Storage:     nil,
		DNSNames:    nil,
	}
}

// NewConfigWithOptions constructs new Config object with predefined options.
func NewConfigWithOptions(opts Config) *Config {
	c := NewConfig()
	c.PublicPort = opts.PublicPort
	c.PrivatePort = opts.PrivatePort
	c.CacheDir = opts.CacheDir
	c.Storage = opts.Storage
	c.DNSNames = opts.DNSNames
	return c
}

// SetStorage sets implemented storage type for Server.
func (c *Config) SetStorage(store Storage) {
	c.Storage = store
}

// DefaultCacheDir returns default cacheDir for Server started by current user.
func DefaultCacheDir() string {
	u, err := user.Current()
	if err != nil {
		return "."
	}
	return filepath.Join(u.HomeDir, ".cache/gophkeeper/server")
}
