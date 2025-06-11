package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/internal/helpers"
	"github.com/sejo412/gophkeeper/internal/storage/sqlite"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// Server main application object.
type Server struct {
	config      *Config
	grpcPublic  *GRPCPublic
	grpcPrivate *GRPCPrivate
}

// NewServer constructs object Server with predefined Config.
func NewServer(opts Config) *Server {
	return &Server{
		config: NewConfigWithOptions(opts),
	}
}

// Init destroys all presents data and creates new data (Storage and certificates).
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
	if err := createServerCert(
		ctx, s.config.DNSNames, helpers.MyIPAddresses(),
		filepath.Join(s.config.CacheDir, constants.CertServerPublicFilename),
		filepath.Join(s.config.CacheDir, constants.CertServerPrivateFilename), caCert, caKey,
	); err != nil {
		return fmt.Errorf("could not create server cert: %w", err)
	}
	return nil
}

// Start starts main application.
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
	registerGRPCPublicServer(publicGRPCServer, s.grpcPublic)
	// for debug
	reflection.Register(publicGRPCServer)

	tlsCfg, err := tlsConfig(s.config.CacheDir)
	if err != nil {
		return fmt.Errorf("could not create TLS configuration: %w", err)
	}
	creds := credentials.NewTLS(tlsCfg)

	s.grpcPrivate = NewGRPCPrivate(*s.config)
	privateGRPCServer := grpc.NewServer(grpcPrivateServerOptions(s.grpcPrivate, creds)...)
	registerGRPCPrivateServer(privateGRPCServer, s.grpcPrivate)

	publicListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.PublicPort))
	if err != nil {
		return fmt.Errorf("could not listen on port %d: %w", s.config.PublicPort, err)
	}
	privateListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.PrivatePort))
	if err != nil {
		return fmt.Errorf("could not listen on port %d: %w", s.config.PrivatePort, err)
	}

	idleConnsClosed := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, constants.GracefulSignals...)
	go func() {
		<-sigs
		fmt.Println()
		slog.Info("shutting down public server...")
		publicGRPCServer.GracefulStop()
		slog.Info("shutting down private server...")
		privateGRPCServer.GracefulStop()
		close(idleConnsClosed)
	}()

	var errGrp errgroup.Group
	errGrp.Go(
		func() error {
			slog.Info("starting public server")
			return publicGRPCServer.Serve(publicListener)
		},
	)
	errGrp.Go(
		func() error {
			slog.Info("starting private server")
			return privateGRPCServer.Serve(privateListener)
		},
	)

	if err = errGrp.Wait(); err != nil {
		return fmt.Errorf("could not start server: %w", err)
	}
	<-idleConnsClosed
	slog.Info("server stopped")
	return nil
}

func tlsConfig(dir string) (*tls.Config, error) {
	derCert, err := os.ReadFile(filepath.Join(dir, constants.CertServerPublicFilename))
	if err != nil {
		return nil, fmt.Errorf("could not read server certificate: %w", err)
	}
	derKey, err := os.ReadFile(filepath.Join(dir, constants.CertServerPrivateFilename))
	if err != nil {
		return nil, fmt.Errorf("could not read private key: %w", err)
	}
	keyPair, err := tls.X509KeyPair(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  constants.PemCertType,
				Bytes: derCert,
			},
		), pem.EncodeToMemory(
			&pem.Block{
				Type:  constants.PemKeyType,
				Bytes: derKey,
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not load server key pair: %w", err)
	}
	derCaCert, err := os.ReadFile(filepath.Join(dir, constants.CertCAPublicFilename))
	if err != nil {
		return nil, fmt.Errorf("could not load CA certificate: %w", err)
	}
	caCert, err := x509.ParseCertificate(derCaCert)
	if err != nil {
		return nil, fmt.Errorf("could not parse CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)
	return &tls.Config{
		Certificates: []tls.Certificate{keyPair},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}
