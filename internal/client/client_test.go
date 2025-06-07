package client

import (
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	type args struct {
		config Config
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "success",
			args: args{
				config: Config{
					PublicAddress:  testPublicAddress,
					PrivateAddress: testPrivateAddress,
					CacheDir:       testCacheDir,
				},
			},
			want: &Client{
				config: &Config{
					PublicAddress:  testPublicAddress,
					PrivateAddress: testPrivateAddress,
					CacheDir:       testCacheDir,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClient(tt.args.config)
			if !reflect.DeepEqual(got.config.PublicAddress, tt.want.config.PublicAddress) {
				t.Errorf("NewClient() = %v, want %v", got.config.PublicAddress, tt.want.config.PublicAddress)
			} else {
				testClient = got
			}
		})
	}
}

func TestClient_Register(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				name: testUser1.Cn,
			},
			wantErr: false,
		},
		{
			name: "already exists",
			args: args{
				name: testUser1.Cn,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testClient.Register(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_SetRSAKeys(t *testing.T) {
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
				config: testClient.config,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				config: tt.fields.config,
			}
			if err := c.SetRSAKeys(); (err != nil) != tt.wantErr {
				t.Errorf("SetRSAKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
