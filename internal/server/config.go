package server

import (
	"context"
	"os/user"
	"path/filepath"

	"github.com/sejo412/gophkeeper/internal/models"
)

type Config struct {
	PublicPort  int
	PrivatePort int
	CacheDir    string
	DNSNames    []string
	Storage     Storage
}

func NewConfig() *Config {
	return &Config{
		PublicPort:  -1,
		PrivatePort: -1,
		CacheDir:    "",
		Storage:     nil,
		DNSNames:    nil,
	}
}

func NewConfigWithOptions(opts Config) *Config {
	c := NewConfig()
	c.PublicPort = opts.PublicPort
	c.PrivatePort = opts.PrivatePort
	c.CacheDir = opts.CacheDir
	c.Storage = opts.Storage
	c.DNSNames = opts.DNSNames
	return c
}

func (c *Config) SetStorage(store Storage) {
	c.Storage = store
}

func DefaultCacheDir() string {
	u, err := user.Current()
	if err != nil {
		return "."
	}
	return filepath.Join(u.HomeDir, ".cache/gophkeeper/server")
}
