package server

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/pkg/certs"
	pb "github.com/sejo412/gophkeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const internalError = "internal error"

type publicConfig struct {
	port   int
	caCert *certs.Cert // Deprecated: use signer
	signer certs.CASigner
	store  Storage
}

func newPublicConfig(port int, caCertFile, caKeyFile string, store *Storage) (publicConfig, error) {
	signer, err := certs.GetSigner(caCertFile, caKeyFile)
	if err != nil {
		return publicConfig{}, fmt.Errorf("error loading CA signer: %w", err)
	}
	return publicConfig{
		port:   port,
		signer: signer,
		store:  *store,
	}, nil
}

type GRPCPublic struct {
	pb.UnimplementedPublicServer
	config publicConfig
}

func (p *GRPCPublic) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	ok := new(bool)
	msg := new(string)
	certRequest, err := certs.BinaryToRequest(in.GetCertRequest())
	if err != nil {
		*ok = false
		*msg = err.Error()
		slog.Error("parsing certificate request", "error", err)
		return &pb.RegisterResponse{
			Ok:    ok,
			Error: msg,
		}, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, err = p.config.store.NewUser(ctx, certRequest.CommonName); err != nil {
		*ok = false
		*msg = err.Error()
		slog.Error("creating new user", "error", err)
		return &pb.RegisterResponse{
			Ok:    ok,
			Error: msg,
		}, status.Error(codes.InvalidArgument, err.Error())
	}
	cert, _, err := certs.GenRsaCert(certRequest, p.config.signer)
	if err != nil {
		*ok = false
		*msg = err.Error()
		slog.Error("generating certificate", "error", err)
		return &pb.RegisterResponse{
			Ok:    ok,
			Error: msg,
		}, status.Error(codes.Internal, internalError)
	}
	*ok = true
	return &pb.RegisterResponse{
		Ok:                ok,
		CaCertificate:     p.config.signer.CACert,
		ClientCertificate: cert,
	}, nil
}

func NewGRPCPublic(config Config) (*GRPCPublic, error) {
	cfg, err := newPublicConfig(
		config.PublicPort,
		filepath.Join(config.CacheDir, constants.CertCAPublicFilename),
		filepath.Join(config.CacheDir, constants.CertCAPrivateFilename),
		&config.Storage,
	)
	if err != nil {
		return nil, fmt.Errorf("error create public config: %w", err)
	}
	return &GRPCPublic{
		config: cfg,
	}, nil
}

func grpcPublicServerOptions(server *GRPCPublic) []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)
	return opts
}

func registerGRPCPublicServer(grpc *grpc.Server, server *GRPCPublic) {
	pb.RegisterPublicServer(grpc, server)
}
