package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/proto"
)

type Server struct {
	config      *Config
	grpcPublic  proto.UnimplementedPublicServer
	grpcPrivate proto.UnimplementedPrivateServer
}

func NewServer(opts Config) *Server {
	return &Server{
		config: NewConfigWithOptions(opts),
	}
}

func (s *Server) Init() error {
	ctx := context.Background()
	if s.config == nil {
		return errors.New("server config not initialized")
	}
	if s.config.CacheDir == "" {
		return errors.New("cache dir not initialized")
	}
	if _, err := os.Stat(s.config.CacheDir); err != nil {
		if er := os.MkdirAll(s.config.CacheDir, 0755); er != nil {
			return fmt.Errorf("could not create cache dir: %w", er)
		}
	}
	dbFile := filepath.Join(s.config.CacheDir, constants.DBFilename)
	if err := createDatabase(ctx, dbFile); err != nil {
		return fmt.Errorf("could not create database: %w", err)
	}
	caCert := filepath.Join(s.config.CacheDir, constants.CertCAPublicFilename)
	caKey := filepath.Join(s.config.CacheDir, constants.CertCAPrivateFilename)
	if err := createCA(ctx, caCert, caKey); err != nil {
		return fmt.Errorf("could not create CA: %w", err)
	}
	return nil
}

func (s *Server) Start() error {
	return nil
}
