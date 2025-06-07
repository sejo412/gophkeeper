package server

import (
	"crypto/tls"
	"reflect"
	"testing"
	"time"
)

func TestServer_Init(t *testing.T) {
	type fields struct {
		config *Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				config: &Config{
					CacheDir: testCacheDir,
				},
			},
			wantErr: false,
		},
		{
			name: "error",
			fields: fields{
				config: &Config{
					CacheDir: "/zzzzzzzzzzzzzzzz",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				config: tt.fields.config,
			}
			if err := s.Init(); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Start(t *testing.T) {
	type fields struct {
		config *Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				config: &Config{
					CacheDir:    testCacheDir,
					PublicPort:  5200,
					PrivatePort: 5201,
				},
			},
			wantErr: false,
		},
		{
			name: "error cache dir",
			fields: fields{
				config: &Config{
					CacheDir:    "/zzzzzzzzzzzzzz",
					PublicPort:  5300,
					PrivatePort: 5301,
				},
			},
			wantErr: true,
		},
		{
			name: "error ports in use",
			fields: fields{
				config: &Config{
					CacheDir:    testCacheDir,
					PublicPort:  5200,
					PrivatePort: 5201,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				config: tt.fields.config,
			}
			time.Sleep(1 * time.Second) // wait for other goroutine binds ports
			go func() {
				if err := s.Start(); (err != nil) != tt.wantErr {
					t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
		})
	}
}

func Test_tlsConfig(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				dir: testCacheDir,
			},
			want: &tls.Config{
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				dir: "/zzzzzzzzzzzzzz",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tlsConfig(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("tlsConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if !reflect.DeepEqual(got.ClientAuth, tt.want.ClientAuth) {
					t.Errorf("tlsConfig() got = %v, want %v", got.ClientAuth, tt.want.ClientAuth)
				}
			}
		})
	}
}
