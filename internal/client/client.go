package client

import (
	"context"
	"fmt"

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
	certRequest := certs.CertRequest{
		CommonName:  name,
		DNSNames:    nil,
		IPAddresses: nil,
		Emails:      nil,
		IsCA:        false,
	}
	req, err := certs.RequestToBinary(certRequest)
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
	return nil
}
