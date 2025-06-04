package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/internal/models"
	"github.com/sejo412/gophkeeper/pkg/certs"
	pb "github.com/sejo412/gophkeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

const internalError = "internal error"
const ctxCNKey = "CommonName"
const ctxUIDKey = "UserID"

type publicConfig struct {
	port   int
	caCert *certs.Cert // Deprecated: use signer
	signer certs.CASigner
	store  Storage
}

type privateConfig struct {
	port  int
	store Storage
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

func NewGRPCPrivate(config Config) *GRPCPrivate {
	cfg := newPrivateConfig(
		config.PrivatePort,
		&config.Storage,
	)
	return &GRPCPrivate{
		config: cfg,
	}
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

func newPrivateConfig(port int, store *Storage) privateConfig {
	return privateConfig{
		port:  port,
		store: *store,
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

func (s *GRPCPrivate) List(ctx context.Context, in *emptypb.Empty) (
	*pb.ListResponse,
	error,
) {
	// TODO implement me
	panic("implement me")
}

func (s *GRPCPrivate) Create(ctx context.Context, in *pb.AddRecordRequest) (*emptypb.Empty, error) {
	ctxUID, _ := ctx.Value(ctxUIDKey).(int)
	uid := models.UserID(ctxUID)
	switch in.GetType() {
	case pb.RecordType_PASSWORD:
		var record models.PasswordEncrypted
		if err := json.Unmarshal(in.GetRecord(), &record); err != nil {
			slog.Error("error unmarshalling password encrypted record")
			return nil, status.Error(codes.Internal, "error unmarshalling password encrypted record")
		}
		if err := s.config.store.Add(
			ctx, uid, models.RecordPassword, models.RecordEncrypted{Password: record},
		); err != nil {
			slog.Error("error adding record", "error", err)
			return nil, status.Error(codes.Internal, "error adding record")
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid type")
	}
	return &emptypb.Empty{}, nil
}

func (s *GRPCPrivate) Read(ctx context.Context, in *pb.GetRecordRequest) (*pb.GetRecordResponse, error) {
	ctxUID, _ := ctx.Value(ctxUIDKey).(int)
	uid := models.UserID(ctxUID)
	switch in.GetType() {
	case pb.RecordType_PASSWORD:
		r, err := s.config.store.Get(ctx, uid, models.RecordPassword, models.ID(in.GetRecordNumber()))
		if err != nil {
			slog.Error("error getting record", "error", err)
			return nil, status.Error(codes.Internal, "error getting record")
		}
		data, err := json.Marshal(r)
		if err != nil {
			slog.Error("error marshalling record", "error", err)
			return nil, status.Error(codes.Internal, "error marshalling record")
		}
		return &pb.GetRecordResponse{
			Type:   protoRecordType(pb.RecordType_PASSWORD),
			Record: data,
			Error:  nil,
		}, nil
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid type")
	}
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

func loggerInterceptor() logging.Logger {
	return logging.LoggerFunc(
		func(ctx context.Context, lvl logging.Level, msg string, keyvals ...any) {
			uid := ctx.Value(ctxUIDKey).(int)
			cn := ctx.Value(ctxCNKey).(string)
			keyvals = append(keyvals, slog.String("User", cn), slog.Int("UserID", uid))
			slog.Log(ctx, slog.Level(lvl), msg, keyvals...)
		},
	)
}

func (s *GRPCPrivate) authInterceptor(
	ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	message := func(msg string, args ...any) string {
		return "[auth] " + fmt.Sprintf(msg, args...)
	}
	p, ok := peer.FromContext(ctx)
	if !ok {
		msg := message("client information not found")
		slog.Info(msg)
		return nil, status.Error(codes.Unauthenticated, msg)
	}
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		msg := message("client credentials not found")
		slog.Info(msg)
		return nil, status.Error(codes.Unauthenticated, message(msg))
	}
	if len(tlsInfo.State.VerifiedChains) == 0 || len(tlsInfo.State.VerifiedChains[0]) == 0 {
		msg := message("client verified certificate not found")
		slog.Info(msg)
		return nil, status.Error(codes.Unauthenticated, msg)
	}

	cert := tlsInfo.State.VerifiedChains[0][0]
	commonName := cert.Subject.CommonName
	if commonName == "" {
		msg := message("client common name not found")
		slog.Info(msg)
		return nil, status.Error(codes.Unauthenticated, msg)
	}
	uid, err := s.config.store.GetUserID(ctx, commonName)
	if err != nil {
		msg := message("user id for %q not found", commonName)
		slog.Info(msg)
		return nil, status.Error(codes.Unauthenticated, msg)
	}
	ctx = context.WithValue(ctx, ctxUIDKey, int(uid))
	ctx = context.WithValue(ctx, ctxCNKey, commonName)
	return handler(ctx, req)
}

func grpcPublicServerOptions(_ *GRPCPublic) []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0)
	unaryInterceptors = append(unaryInterceptors, logging.UnaryServerInterceptor(loggerInterceptor()))
	opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	return opts
}

func registerGRPCPublicServer(grpc *grpc.Server, server *GRPCPublic) {
	pb.RegisterPublicServer(grpc, server)
}

func grpcPrivateServerOptions(server *GRPCPrivate, creds credentials.TransportCredentials) []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0)
	unaryInterceptors = append(unaryInterceptors, server.authInterceptor)
	unaryInterceptors = append(unaryInterceptors, logging.UnaryServerInterceptor(loggerInterceptor()))
	opts = append(opts, grpc.Creds(creds))
	opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	return opts
}

func registerGRPCPrivateServer(grpc *grpc.Server, server *GRPCPrivate) {
	pb.RegisterPrivateServer(grpc, server)
}

func stringToPtr(s string) *string {
	return &s
}

func protoRecordType(r pb.RecordType) *pb.RecordType {
	return &r
}
