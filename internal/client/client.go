package client

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/internal/helpers"
	"github.com/sejo412/gophkeeper/pkg/certs"
	pb "github.com/sejo412/gophkeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	config *Config
}

func NewClient(config Config) *Client {
	return &Client{config: NewConfigWithOptions(config)}
}

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
	if resp.Error != nil {
		return fmt.Errorf("failed to register user: %s", resp.GetError())
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
