package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/internal/storage/sqlite"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	config     *Config
	grpcPublic *GRPCPublic
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
	// open storage
	store, err := sqlite.New(filepath.Join(s.config.CacheDir, constants.DBFilename))
	if err != nil {
		return fmt.Errorf("could not open storage: %w", err)
	}
	defer func() {
		_ = store.Close()
	}()
	s.config.SetStorage(store)

	// start grpc servers
	s.grpcPublic, err = NewGRPCPublic(*s.config)
	if err != nil {
		return fmt.Errorf("could not create public server: %w", err)
	}
	publicGRPCServer := grpc.NewServer(grpcPublicServerOptions(s.grpcPublic)...)
	slog.Info("starting public server")
	registerGRPCPublicServer(publicGRPCServer, s.grpcPublic)

	// for debug
	reflection.Register(publicGRPCServer)

	publicListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.PublicPort))
	if err != nil {
		return fmt.Errorf("could not listen on port %d: %w", s.config.PublicPort, err)
	}

	idleConnsClosed := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, constants.GracefulSignals...)
	go func() {
		<-sigs
		slog.Info("shutting down server...")
		publicGRPCServer.GracefulStop()
		close(idleConnsClosed)
	}()

	var errGrp errgroup.Group
	errGrp.Go(
		func() error {
			return publicGRPCServer.Serve(publicListener)
		},
	)

	if err = errGrp.Wait(); err != nil {
		return fmt.Errorf("could not start server: %w", err)
	}
	<-idleConnsClosed
	slog.Info("server stopped")
	return nil
}
