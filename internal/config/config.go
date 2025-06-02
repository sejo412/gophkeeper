package config

import (
	"context"
	"os/user"
	"path/filepath"

	"github.com/sejo412/gophkeeper/internal/models"
)

type Storage interface {
	Init(ctx context.Context) error
	Close() error
	List(ctx context.Context, uid models.UserID) (models.RecordsEncrypted, error)
	Get(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) (models.RecordEncrypted, error)
	Add(ctx context.Context, uid models.UserID, t models.RecordType, record models.RecordEncrypted) error
	Update(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID,
		record models.RecordEncrypted) error
	Delete(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) error
	IsExist(ctx context.Context, user models.UserID, t models.RecordType, id models.ID) (bool, error)
	Users(ctx context.Context) ([]*models.User, error)
	NewUser(ctx context.Context, uid string) (*models.UserID, error)
	IsUserExist(ctx context.Context, uid models.UserID) (bool, error)
}

type ServerConfig struct {
	PublicPort  int
	PrivatePort int
	CacheDir    string
	DNSNames    []string
	Storage     Storage
}

type ClientConfig struct{}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		PublicPort:  -1,
		PrivatePort: -1,
		CacheDir:    "",
		Storage:     nil,
		DNSNames:    nil,
	}
}

func NewServerConfigWithOptions(opts ServerConfig) *ServerConfig {
	c := NewServerConfig()
	c.PublicPort = opts.PublicPort
	c.PrivatePort = opts.PrivatePort
	c.CacheDir = opts.CacheDir
	c.Storage = opts.Storage
	c.DNSNames = opts.DNSNames
	return c
}

func (s *ServerConfig) SetStorage(store Storage) {
	s.Storage = store
}

func DefaultServerCacheDir() string {
	u, err := user.Current()
	if err != nil {
		return "."
	}
	return filepath.Join(u.HomeDir, ".cache/gophkeeper/server")
}

func DefaultClientCacheDir() string {
	u, err := user.Current()
	if err != nil {
		return "."
	}
	return filepath.Join(u.HomeDir, ".cache/gophkeeper/client")
}
