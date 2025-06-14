package client

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/internal/helpers"
	"github.com/sejo412/gophkeeper/pkg/certs"
	pb "github.com/sejo412/gophkeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a main client object.
type Client struct {
	config     *Config
	client     pb.PrivateClient
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewClient constructs Client object.
func NewClient(config Config) *Client {
	return &Client{config: NewConfigWithOptions(config)}
}

// Register registers new client on the server with creating RSA-keys and certificates.
func (c *Client) Register(name string) error {
	if err := os.MkdirAll(c.config.CacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}
	privKey, err := certs.GenRsaKey(constants.KeyBits)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}
	keyOut, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}
	if err = helpers.SaveRegularFile(
		filepath.Join(c.config.CacheDir, constants.CertClientPrivateFilename), keyOut, 0600,
	); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}
	csr := certs.NewCertRequest(name, nil, nil, nil, false)
	if err = csr.Sign(keyOut); err != nil {
		return fmt.Errorf("failed to sign certificate request: %w", err)
	}
	req, err := certs.RequestToBinary(*csr)
	if err != nil {
		return fmt.Errorf("failed to create certificate request: %w", err)
	}
	grpcClient, err := grpc.NewClient(c.config.PublicAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create public client: %w", err)
	}
	defer func() {
		_ = grpcClient.Close()
	}()
	publicClient := pb.NewPublicClient(grpcClient)
	resp, err := publicClient.Register(
		context.Background(), &pb.RegisterRequest{
			CertRequest: req,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	if err = helpers.SaveRegularFile(
		filepath.Join(c.config.CacheDir, constants.CertCAPublicFilename),
		resp.GetCaCertificate(), 0644,
	); err != nil {
		return fmt.Errorf("failed to save CA certificate: %w", err)
	}
	if err = helpers.SaveRegularFile(
		filepath.Join(c.config.CacheDir, constants.CertClientPublicFilename), resp.GetClientCertificate(), 0644,
	); err != nil {
		return fmt.Errorf("failed to save client certificate: %w", err)
	}
	return nil
}

// Run runs application's interactive mode.
func (c *Client) Run() error {
	tlsCfg, err := tlsConfig(c.config.CacheDir)
	if err != nil {
		return fmt.Errorf("failed to create tls config: %w", err)
	}
	if err = c.SetRSAKeys(); err != nil {
		return fmt.Errorf("failed to set RSA keys: %w", err)
	}
	grpcClient, err := grpc.NewClient(c.config.PrivateAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	if err != nil {
		return fmt.Errorf("failed to create private client: %w", err)
	}
	defer func() {
		_ = grpcClient.Close()
	}()
	c.client = pb.NewPrivateClient(grpcClient)
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sig:
			fmt.Println("\nShutting down...")
			cancel()
			os.Exit(0)
		case <-ctx.Done():
			return
		}
	}()
	mainMenu(ctx, c)
	return nil
}

// SetRSAKeys sets readable keys to Client object.
func (c *Client) SetRSAKeys() error {
	derKey, err := os.ReadFile(filepath.Join(c.config.CacheDir, constants.CertClientPrivateFilename))
	if err != nil {
		return fmt.Errorf("could not read private key: %w", err)
	}
	key, err := x509.ParsePKCS8PrivateKey(derKey)
	if err != nil {
		return fmt.Errorf("could not parse private key: %w", err)
	}
	c.privateKey = key.(*rsa.PrivateKey)
	c.publicKey = &c.privateKey.PublicKey
	return nil
}

func tlsConfig(dir string) (*tls.Config, error) {
	derCert, err := os.ReadFile(filepath.Join(dir, constants.CertClientPublicFilename))
	if err != nil {
		return nil, fmt.Errorf("could not read client certificate: %w", err)
	}
	derKey, err := os.ReadFile(filepath.Join(dir, constants.CertClientPrivateFilename))
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
		return nil, fmt.Errorf("could not load client key pair: %w", err)
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
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
