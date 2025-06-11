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

const (
	errorInternal  = "internal error"
	errorDelete    = "error deleting record"
	errorList      = "error listing records"
	errorMarshal   = "error marshalling record"
	errorUnmarshal = "error unmarshalling record"
	errorAdd       = "error adding record"
	errorGet       = "error getting record"
	errorUpdate    = "error updating record"
)

type ctxKey string

const ctxCNKey ctxKey = "CommonName"
const ctxUIDKey ctxKey = "UserID"

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

// NewGRPCPublic constructs new GRPCPublic object for Server.
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

// NewGRPCPrivate constructs new GRPCPrivate object for Server.
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

// GRPCPublic implements proto public server.
type GRPCPublic struct {
	pb.UnimplementedPublicServer
	config publicConfig
}

// GRPCPrivate implements proto private server.
type GRPCPrivate struct {
	pb.UnimplementedPrivateServer
	config privateConfig
}

// ListAll returns ID and Meta for all records by User ID.
func (s *GRPCPrivate) ListAll(ctx context.Context, in *emptypb.Empty) (
	*pb.ListResponse,
	error,
) {
	ctxUID, _ := ctx.Value(ctxUIDKey).(int)
	uid := models.UserID(ctxUID)
	r, err := s.config.store.ListAll(ctx, uid)
	if err != nil {
		slog.Info(errorList, "error", err)
		return nil, status.Error(codes.Internal, errorList)
	}
	data, err := json.Marshal(r)
	if err != nil {
		slog.Info(errorMarshal, "error", err)
		return nil, status.Error(codes.Internal, errorMarshal)
	}
	return &pb.ListResponse{Records: data}, nil
}

// List returns ID and Meta for all records by User ID and models.RecordType.
func (s *GRPCPrivate) List(ctx context.Context, in *pb.ListRequest) (*pb.ListResponse, error) {
	ctxUID, _ := ctx.Value(ctxUIDKey).(int)
	uid := models.UserID(ctxUID)
	r, err := s.config.store.List(ctx, uid, protoRecordTypeToModel(in.GetType()))
	if err != nil {
		slog.Info(errorList, "error", err)
		return nil, status.Error(codes.Internal, errorList)
	}
	data, err := json.Marshal(r)
	if err != nil {
		slog.Info(errorMarshal, "error", err)
		return nil, status.Error(codes.Internal, errorMarshal)
	}
	return &pb.ListResponse{Records: data}, nil
}

// Create creates new models.RecordEncrypted for User by models.RecordType.
func (s *GRPCPrivate) Create(ctx context.Context, in *pb.AddRecordRequest) (*emptypb.Empty, error) {
	ctxUID, _ := ctx.Value(ctxUIDKey).(int)
	uid := models.UserID(ctxUID)
	var record models.RecordEncrypted
	if err := json.Unmarshal(in.GetRecord(), &record); err != nil {
		slog.Info(errorUnmarshal, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, errorUnmarshal)
	}
	if err := s.config.store.Add(ctx, uid, protoRecordTypeToModel(in.GetType()), record); err != nil {
		slog.Info(errorAdd, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, errorAdd)
	}
	return &emptypb.Empty{}, nil
}

// Read returns models.Record for User by models.RecordType and models.ID.
func (s *GRPCPrivate) Read(ctx context.Context, in *pb.GetRecordRequest) (*pb.GetRecordResponse, error) {
	ctxUID, _ := ctx.Value(ctxUIDKey).(int)
	uid := models.UserID(ctxUID)
	r, err := s.config.store.Get(ctx, uid, protoRecordTypeToModel(in.GetType()), models.ID(in.GetRecordNumber()))
	if err != nil {
		slog.Error(errorGet, "error", err)
		return nil, status.Error(codes.Internal, errorGet)
	}
	data, err := json.Marshal(r)
	if err != nil {
		slog.Error(errorMarshal, "error", err)
		return nil, status.Error(codes.Internal, errorMarshal)
	}
	return &pb.GetRecordResponse{
		Type:   in.Type,
		Record: data,
		Error:  nil,
	}, nil
}

// Update updates models.Record for User by models.RecordType and models.ID.
func (s *GRPCPrivate) Update(ctx context.Context, in *pb.UpdateRecordRequest) (*emptypb.Empty, error) {
	ctxUID, _ := ctx.Value(ctxUIDKey).(int)
	uid := models.UserID(ctxUID)
	var record models.RecordEncrypted
	if err := json.Unmarshal(in.GetRecord(), &record); err != nil {
		slog.Info(errorUnmarshal, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, errorUnmarshal)
	}
	if err := s.config.store.Update(
		ctx, uid, protoRecordTypeToModel(in.GetType()), models.ID(in.GetRecordNumber()),
		record,
	); err != nil {
		slog.Info(errorUpdate, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, errorUpdate)
	}
	return &emptypb.Empty{}, nil
}

// Delete deletes models.Record for User by models.RecordType and models.ID.
func (s *GRPCPrivate) Delete(ctx context.Context, in *pb.DeleteRecordRequest) (*emptypb.Empty, error) {
	ctxUID, _ := ctx.Value(ctxUIDKey).(int)
	uid := models.UserID(ctxUID)
	if err := s.config.store.Delete(
		ctx, uid, protoRecordTypeToModel(in.GetType()),
		models.ID(in.GetRecordNumber()),
	); err != nil {
		slog.Info(errorDelete, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, errorDelete)
	}
	return &emptypb.Empty{}, nil
}

// Register creates new models.User by certificate request.
func (sp *GRPCPublic) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	msg := new(string)
	certRequest, err := certs.BinaryToRequest(in.GetCertRequest())
	if err != nil {
		*msg = err.Error()
		slog.Error("parsing certificate request", "error", err)
		return &pb.RegisterResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, err = sp.config.store.NewUser(ctx, certRequest.CommonName); err != nil {
		*msg = err.Error()
		slog.Error("creating new user", "error", err)
		return &pb.RegisterResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}
	cert, _, err := certs.GenRsaCert(certRequest, sp.config.signer)
	if err != nil {
		*msg = err.Error()
		slog.Error("generating certificate", "error", err)
		return &pb.RegisterResponse{}, status.Error(codes.Internal, errorInternal)
	}
	return &pb.RegisterResponse{
		CaCertificate:     sp.config.signer.CACert,
		ClientCertificate: cert,
	}, nil
}

func loggerInterceptor() logging.Logger {
	return logging.LoggerFunc(
		func(ctx context.Context, lvl logging.Level, msg string, keyvals ...any) {
			uidCtx := ctx.Value(ctxUIDKey)
			uid := -1
			if uidCtx != nil {
				uid = uidCtx.(int)
			}
			cnCtx := ctx.Value(ctxCNKey)
			cn := ""
			if cnCtx != nil {
				cn = cnCtx.(string)
			}
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

func protoRecordTypeToModel(r pb.RecordType) models.RecordType {
	switch r {
	case pb.RecordType_PASSWORD:
		return models.RecordPassword
	case pb.RecordType_TEXT:
		return models.RecordText
	case pb.RecordType_BIN:
		return models.RecordBin
	case pb.RecordType_BANK:
		return models.RecordBank
	default:
		return models.RecordUnknown
	}
}
