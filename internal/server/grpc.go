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
	"google.golang.org/protobuf/types/known/emptypb"
)

const internalError = "internal error"
const ctxCNKey = "CommonName"

type publicConfig struct {
	port   int
	caCert *certs.Cert // Deprecated: use signer
	signer certs.CASigner
	store  Storage
}

type privateConfig struct {
	port   int
	caCert string
	cert   string
	key    string
	store  Storage
}

func newPublicConfig(port int, caCertFile, caKeyFile string, store *Storage) (publicConfig, error) {
	signer, err := certs.GetCASigner(caCertFile, caKeyFile)
	if err != nil {
		return publicConfig{}, fmt.Errorf("error loading CA signer: %w", err)
	}
	return publicConfig{
		port:   port,
		signer: signer,
		store:  *store,
	}, nil
}

func newPrivateConfig(port int, caCertFile, certFile, keyFile string, store *Storage) privateConfig {
	return privateConfig{
		port:   port,
		caCert: caCertFile,
		cert:   certFile,
		key:    keyFile,
		store:  *store,
	}
}

type GRPCPublic struct {
	pb.UnimplementedPublicServer
	config publicConfig
}

type GRPCPrivate struct {
	pb.UnimplementedPrivateServer
	config privateConfig
}

func (s *GRPCPrivate) List(ctx context.Context, in *emptypb.Empty) (*pb.ListResponse,
	error) {
	// TODO implement me
	panic("implement me")
}

func (s *GRPCPrivate) Create(ctx context.Context, in *pb.AddRecordRequest) (*pb.AddRecordResponse, error) {
	_, ok := ctx.Value(ctxCNKey).(string)
	if !ok {
		return nil, status.Error(codes.Internal, internalError)
	}
	switch in.GetType() {
	case pb.RecordType_PASSWORD:
		return nil, status.Error(codes.Unknown, internalError)
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid type")
	}
}

func (s *GRPCPrivate) Read(ctx context.Context, in *pb.GetRecordRequest) (*pb.GetRecordResponse,
	error) {
	// TODO implement me
	panic("implement me")
}

func (s *GRPCPrivate) Update(ctx context.Context, in *pb.UpdateRecordRequest) (*pb.UpdateRecordResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (s *GRPCPrivate) Delete(ctx context.Context, in *pb.DeleteRecordRequest) (*pb.DeleteRecordResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (sp *GRPCPublic) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	msg := new(string)
	certRequest, err := certs.BinaryToRequest(in.GetCertRequest())
	if err != nil {
		*msg = err.Error()
		slog.Error("parsing certificate request", "error", err)
		return &pb.RegisterResponse{
			Error: msg,
		}, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, err = sp.config.store.NewUser(ctx, certRequest.CommonName); err != nil {
		*msg = err.Error()
		slog.Error("creating new user", "error", err)
		return &pb.RegisterResponse{
			Error: msg,
		}, status.Error(codes.InvalidArgument, err.Error())
	}
	cert, _, err := certs.GenRsaCert(certRequest, sp.config.signer)
	if err != nil {
		*msg = err.Error()
		slog.Error("generating certificate", "error", err)
		return &pb.RegisterResponse{
			Error: msg,
		}, status.Error(codes.Internal, internalError)
	}
	return &pb.RegisterResponse{
		CaCertificate:     sp.config.signer.CACert,
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

func grpcPublicServerOptions(_ *GRPCPublic) []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)
	return opts
}

func registerGRPCPublicServer(grpc *grpc.Server, server *GRPCPublic) {
	pb.RegisterPublicServer(grpc, server)
}

func NewGRPCPrivate(config Config) *GRPCPrivate {
	cfg := newPrivateConfig(
		config.PrivatePort,
		filepath.Join(config.CacheDir, constants.CertCAPublicFilename),
		filepath.Join(config.CacheDir, constants.CertServerPublicFilename),
		filepath.Join(config.CacheDir, constants.CertServerPrivateFilename),
		&config.Storage,
	)
	return &GRPCPrivate{
		config: cfg,
	}
}

func grpcPrivateServerOptions(_ *GRPCPrivate) []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)
	return opts
}

func registerGRPCPrivateServer(grpc *grpc.Server, server *GRPCPrivate) {
	pb.RegisterPrivateServer(grpc, server)
}
