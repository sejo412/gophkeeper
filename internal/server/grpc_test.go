package server

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/internal/models"
	pb "github.com/sejo412/gophkeeper/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	testRecordTypePassword = pb.RecordType_PASSWORD
)

func Test_protoRecordTypeToModel(t *testing.T) {
	type args struct {
		r pb.RecordType
	}
	tests := []struct {
		name string
		args args
		want models.RecordType
	}{
		{
			name: "password",
			args: args{
				r: pb.RecordType_PASSWORD,
			},
			want: models.RecordPassword,
		},
		{
			name: "text",
			args: args{
				r: pb.RecordType_TEXT,
			},
			want: models.RecordText,
		},
		{
			name: "bin",
			args: args{
				r: pb.RecordType_BIN,
			},
			want: models.RecordBin,
		},
		{
			name: "bank",
			args: args{
				r: pb.RecordType_BANK,
			},
			want: models.RecordBank,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := protoRecordTypeToModel(tt.args.r); got != tt.want {
				t.Errorf("protoRecordTypeToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newPublicConfig(t *testing.T) {
	type args struct {
		port       int
		caCertFile string
		caKeyFile  string
		store      *Storage
	}
	tests := []struct {
		name    string
		args    args
		want    publicConfig
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				port:       testServer.config.PublicPort,
				caCertFile: filepath.Join(testCacheDir, constants.CertCAPublicFilename),
				caKeyFile:  filepath.Join(testCacheDir, constants.CertCAPrivateFilename),
				store:      &testServer.config.Storage,
			},
			want: publicConfig{
				port: testServer.config.PublicPort,
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				port:       testServer.config.PublicPort,
				caCertFile: filepath.Join("/zzzzzzzz", constants.CertCAPublicFilename),
				caKeyFile:  filepath.Join("/zzzzzzzz", constants.CertCAPrivateFilename),
				store:      &testServer.config.Storage,
			},
			want:    publicConfig{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newPublicConfig(tt.args.port, tt.args.caCertFile, tt.args.caKeyFile, tt.args.store)
			if (err != nil) != tt.wantErr {
				t.Errorf("newPublicConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.port, tt.want.port) {
				t.Errorf("newPublicConfig() got = %v, want %v", got.port, tt.want.port)
			}
		})
	}
}

func Test_newPrivateConfig(t *testing.T) {
	type args struct {
		port  int
		store *Storage
	}
	tests := []struct {
		name string
		args args
		want privateConfig
	}{
		{
			name: "success",
			args: args{
				port:  testServer.config.PrivatePort,
				store: &testServer.config.Storage,
			},
			want: privateConfig{
				port: testServer.config.PrivatePort,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newPrivateConfig(tt.args.port, tt.args.store); !reflect.DeepEqual(got.port, tt.want.port) {
				t.Errorf("newPrivateConfig() = %v, want %v", got.port, tt.want.port)
			}
		})
	}
}

func TestGRPCPublic_Register(t *testing.T) {
	type fields struct {
		UnimplementedPublicServer pb.UnimplementedPublicServer
		config                    publicConfig
	}
	type args struct {
		ctx context.Context
		in  *pb.RegisterRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.RegisterResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				config: testServer.grpcPublic.config,
			},
			args: args{
				ctx: context.Background(),
				in: &pb.RegisterRequest{
					CertRequest: testUser1CertRequest,
				},
			},
			want: &pb.RegisterResponse{
				CaCertificate: []byte("some"),
			},
			wantErr: false,
		},
		{
			name: "error already registered",
			fields: fields{
				config: testServer.grpcPublic.config,
			},
			args: args{
				ctx: context.Background(),
				in: &pb.RegisterRequest{
					CertRequest: testUser1CertRequest,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error request",
			fields: fields{
				config: testServer.grpcPublic.config,
			},
			args: args{
				ctx: context.Background(),
				in: &pb.RegisterRequest{
					CertRequest: []byte("some shit"),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &GRPCPublic{
				UnimplementedPublicServer: tt.fields.UnimplementedPublicServer,
				config:                    tt.fields.config,
			}
			got, err := sp.Register(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if len(got.ClientCertificate) == 0 || len(got.CaCertificate) == 0 {
					t.Errorf("Register() got empty cert")
				}
			}
		})
	}
}

func TestGRPCPrivate_Create(t *testing.T) {
	validCtx := context.Background()
	validCtx = context.WithValue(validCtx, ctxUIDKey, 1)

	type fields struct {
		UnimplementedPrivateServer pb.UnimplementedPrivateServer
		config                     privateConfig
	}
	type args struct {
		ctx context.Context
		in  *pb.AddRecordRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
				in: &pb.AddRecordRequest{
					Type:   &testRecordTypePassword,
					Record: testRecordPasswordEncrypted,
				},
			},
			wantErr: false,
		},
		{
			name: "error unauthorized",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: context.Background(),
				in: &pb.AddRecordRequest{
					Type:   &testRecordTypePassword,
					Record: testRecordPasswordEncrypted,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCPrivate{
				UnimplementedPrivateServer: tt.fields.UnimplementedPrivateServer,
				config:                     tt.fields.config,
			}
			_, err := s.Create(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGRPCPrivate_Update(t *testing.T) {
	validCtx := context.Background()
	validCtx = context.WithValue(validCtx, ctxUIDKey, 1)
	invalidCtx := context.Background()
	invalidCtx = context.WithValue(invalidCtx, ctxUIDKey, 42)
	validRecordID := new(int64)
	*validRecordID = 1
	invalidRecordID := new(int64)
	*invalidRecordID = 42

	type fields struct {
		UnimplementedPrivateServer pb.UnimplementedPrivateServer
		config                     privateConfig
	}
	type args struct {
		ctx context.Context
		in  *pb.UpdateRecordRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
				in: &pb.UpdateRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: validRecordID,
					Record:       testNewRecordPasswordEncrypted,
				},
			},
			wantErr: false,
		},
		{
			name: "error unauthorized",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: context.Background(),
				in: &pb.UpdateRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: validRecordID,
					Record:       testNewRecordPasswordEncrypted,
				},
			},
			wantErr: true,
		},
		{
			name: "error request id",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
				in: &pb.UpdateRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: invalidRecordID,
					Record:       testNewRecordPasswordEncrypted,
				},
			},
			wantErr: true,
		},
		{
			name: "error invalid user for record",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: invalidCtx,
				in: &pb.UpdateRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: validRecordID,
					Record:       testNewRecordPasswordEncrypted,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCPrivate{
				UnimplementedPrivateServer: tt.fields.UnimplementedPrivateServer,
				config:                     tt.fields.config,
			}
			_, err := s.Update(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGRPCPrivate_Read(t *testing.T) {
	validCtx := context.Background()
	validCtx = context.WithValue(validCtx, ctxUIDKey, 1)
	invalidCtx := context.Background()
	invalidCtx = context.WithValue(invalidCtx, ctxUIDKey, 42)
	validRecordID := new(int64)
	*validRecordID = 1
	invalidRecordID := new(int64)
	*invalidRecordID = 42

	type fields struct {
		UnimplementedPrivateServer pb.UnimplementedPrivateServer
		config                     privateConfig
	}
	type args struct {
		ctx context.Context
		in  *pb.GetRecordRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
				in: &pb.GetRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: validRecordID,
				},
			},
			wantErr: false,
		},
		{
			name: "error unauthorized",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: invalidCtx,
				in: &pb.GetRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: validRecordID,
				},
			},
			wantErr: true,
		},
		{
			name: "error record not found",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
				in: &pb.GetRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: invalidRecordID,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCPrivate{
				UnimplementedPrivateServer: tt.fields.UnimplementedPrivateServer,
				config:                     tt.fields.config,
			}
			_, err := s.Read(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGRPCPrivate_List(t *testing.T) {
	validCtx := context.Background()
	validCtx = context.WithValue(validCtx, ctxUIDKey, 1)
	invalidCtx := context.Background()
	invalidCtx = context.WithValue(invalidCtx, ctxUIDKey, 42)
	validRecordID := new(int64)
	*validRecordID = 1
	invalidRecordID := new(int64)
	*invalidRecordID = 42

	type fields struct {
		UnimplementedPrivateServer pb.UnimplementedPrivateServer
		config                     privateConfig
	}
	type args struct {
		ctx context.Context
		in  *pb.ListRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
				in: &pb.ListRequest{
					Type: &testRecordTypePassword,
				},
			},
			wantErr: false,
		},
		{
			name: "error unauthorized",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: invalidCtx,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCPrivate{
				UnimplementedPrivateServer: tt.fields.UnimplementedPrivateServer,
				config:                     tt.fields.config,
			}
			_, err := s.List(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGRPCPrivate_ListAll(t *testing.T) {
	validCtx := context.Background()
	validCtx = context.WithValue(validCtx, ctxUIDKey, 1)

	type fields struct {
		UnimplementedPrivateServer pb.UnimplementedPrivateServer
		config                     privateConfig
	}
	type args struct {
		ctx context.Context
		in  *emptypb.Empty
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCPrivate{
				UnimplementedPrivateServer: tt.fields.UnimplementedPrivateServer,
				config:                     tt.fields.config,
			}
			_, err := s.ListAll(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGRPCPrivate_Delete(t *testing.T) {
	validCtx := context.Background()
	validCtx = context.WithValue(validCtx, ctxUIDKey, 1)
	invalidCtx := context.Background()
	invalidCtx = context.WithValue(invalidCtx, ctxUIDKey, 42)
	validRecordID := new(int64)
	*validRecordID = 1
	invalidRecordID := new(int64)
	*invalidRecordID = 42

	type fields struct {
		UnimplementedPrivateServer pb.UnimplementedPrivateServer
		config                     privateConfig
	}
	type args struct {
		ctx context.Context
		in  *pb.DeleteRecordRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error unauthorized",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: invalidCtx,
				in: &pb.DeleteRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: validRecordID,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid record",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
				in: &pb.DeleteRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: invalidRecordID,
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				config: testServer.grpcPrivate.config,
			},
			args: args{
				ctx: validCtx,
				in: &pb.DeleteRecordRequest{
					Type:         &testRecordTypePassword,
					RecordNumber: validRecordID,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCPrivate{
				UnimplementedPrivateServer: tt.fields.UnimplementedPrivateServer,
				config:                     tt.fields.config,
			}
			_, err := s.Delete(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
