package client

import (
	"os/user"
	"path/filepath"
)

type Config struct {
	PublicAddress  string
	PrivateAddress string
	CacheDir       string
}

func NewConfig() *Config {
	return &Config{
		PublicAddress:  "",
		PrivateAddress: "",
		CacheDir:       "",
	}
}

func NewConfigWithOptions(config Config) *Config {
	c := NewConfig()
	c.PublicAddress = config.PublicAddress
	c.PrivateAddress = config.PrivateAddress
	c.CacheDir = config.CacheDir
	return c
}

func DefaultCacheDir() string {
	u, err := user.Current()
	if err != nil {
		return "."
	}
	return filepath.Join(u.HomeDir, ".cache/gophkeeper/client")
}
