package sqlite

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sejo412/gophkeeper/internal/models"
)

var (
	testDB    *Storage
	testDBSql *sql.DB
	testUser1 = models.User{
		ID: models.UserID(1),
		Cn: "testUser1",
	}
	testUser2 = models.User{
		ID: models.UserID(2),
		Cn: "testUser2",
	}
	testUser42 = models.User{
		ID: models.UserID(42),
		Cn: "testUser42",
	}
	testPasswordEncrypted1 = models.PasswordEncrypted{
		ID:       1,
		Login:    []byte("preved"),
		Password: []byte("medved"),
		Meta:     []byte(`krevedko's site`),
	}
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "gophkeeper")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	testDBSql, err = sql.Open("sqlite3", filepath.Join(tmpDir, "sqlite.db"))
	if err != nil {
		panic(err)
	}
	testDB = &Storage{
		db: testDBSql,
	}
	defer func() {
		_ = testDB.Close()
	}()
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestNew(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		// want    *Storage
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: ":memory:",
			},
			wantErr: false,
		},
		{
			name: "failure",
			args: args{
				path: "/preved.db",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_, err := New(tt.args.path)
				if (err != nil) != tt.wantErr {
					t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}

func TestStorage_Init(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx context.Context
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
				db: testDBSql,
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				if err := s.Init(tt.args.ctx); (err != nil) != tt.wantErr {
					t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestStorage_NewUser(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx context.Context
		cn  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.UserID
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx: context.Background(),
				cn:  testUser1.Cn,
			},
			want:    testUser1.ID,
			wantErr: false,
		},
		{
			name: "already exists",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx: context.Background(),
				cn:  testUser1.Cn,
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				got, err := s.NewUser(tt.args.ctx, tt.args.cn)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewUser() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorage_userIDbyCn(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx context.Context
		cn  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.UserID
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx: context.Background(),
				cn:  testUser1.Cn,
			},
			want:    testUser1.ID,
			wantErr: false,
		},

		{
			name: "not found",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx: context.Background(),
				cn:  testUser42.Cn,
			},
			want:    -1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				got, err := s.GetUserID(tt.args.ctx, tt.args.cn)
				if (err != nil) != tt.wantErr {
					t.Errorf("userIDbyCn() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("userIDbyCn() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorage_Users(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.User
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx: context.Background(),
			},
			want: []models.User{
				testUser1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				got, err := s.Users(tt.args.ctx)
				if (err != nil) != tt.wantErr {
					t.Errorf("Users() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Users() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorage_IsUserExist(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx  context.Context
		user models.UserID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "exists",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser1.ID,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not found",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser42.ID,
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				got, err := s.IsUserExist(tt.args.ctx, tt.args.user)
				if (err != nil) != tt.wantErr {
					t.Errorf("IsUserExist() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("IsUserExist() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorage_Add(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx    context.Context
		user   models.UserID
		t      models.RecordType
		record models.RecordEncrypted
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "password success",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser1.ID,
				t:    models.RecordPassword,
				record: models.RecordEncrypted{
					Password: testPasswordEncrypted1,
				},
			},
			wantErr: false,
		},
		{
			name: "password failed",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser42.ID,
				t:    models.RecordPassword,
				record: models.RecordEncrypted{
					Password: testPasswordEncrypted1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				if err := s.Add(tt.args.ctx, tt.args.user, tt.args.t, tt.args.record); (err != nil) != tt.wantErr {
					t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestStorage_IsExist(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx  context.Context
		user models.UserID
		t    models.RecordType
		id   models.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "record exists",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser1.ID,
				t:    models.RecordPassword,
				id:   models.ID(1),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "record not found",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser42.ID,
				t:    models.RecordPassword,
				id:   models.ID(1),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				got, err := s.IsExist(tt.args.ctx, tt.args.user, tt.args.t, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("IsExist() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("IsExist() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorage_List(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx  context.Context
		user models.UserID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.RecordsEncrypted
		wantErr bool
	}{
		{
			name: "record list",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser1.ID,
			},
			want: models.RecordsEncrypted{
				Password: []models.PasswordEncrypted{testPasswordEncrypted1},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				got, err := s.ListAll(tt.args.ctx, tt.args.user)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListAll() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListAll() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorage_Get(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx  context.Context
		user models.UserID
		t    models.RecordType
		id   models.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.RecordEncrypted
		wantErr bool
	}{
		{
			name: "record found",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser1.ID,
				t:    models.RecordPassword,
				id:   models.ID(1),
			},
			want: models.RecordEncrypted{
				Password: testPasswordEncrypted1,
			},
			wantErr: false,
		},
		{
			name: "record not found",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser42.ID,
				t:    models.RecordPassword,
				id:   models.ID(1),
			},
			want:    models.RecordEncrypted{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				got, err := s.Get(tt.args.ctx, tt.args.user, tt.args.t, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Get() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestStorage_Update(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx    context.Context
		user   models.UserID
		t      models.RecordType
		id     models.ID
		record models.RecordEncrypted
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
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser1.ID,
				t:    models.RecordPassword,
				id:   models.ID(1),
				record: models.RecordEncrypted{
					Password: models.PasswordEncrypted{
						Login:    testPasswordEncrypted1.Password,
						Password: []byte("new password"),
						Meta:     testPasswordEncrypted1.Meta,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "record not found",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser42.ID,
				t:    models.RecordPassword,
				id:   models.ID(1),
				record: models.RecordEncrypted{
					Password: models.PasswordEncrypted{
						Login:    []byte("login"),
						Password: []byte("password"),
						Meta:     []byte("meta data"),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				if err := s.Update(
					tt.args.ctx, tt.args.user, tt.args.t, tt.args.id,
					tt.args.record,
				); (err != nil) != tt.wantErr {
					t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestStorage_Delete(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx  context.Context
		user models.UserID
		t    models.RecordType
		id   models.ID
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
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser1.ID,
				t:    models.RecordPassword,
				id:   models.ID(1),
			},
			wantErr: false,
		},
		{
			name: "record not found",
			fields: fields{
				db: testDBSql,
			},
			args: args{
				ctx:  context.Background(),
				user: testUser42.ID,
				t:    models.RecordPassword,
				id:   models.ID(1),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &Storage{
					db: tt.fields.db,
				}
				if err := s.Delete(tt.args.ctx, tt.args.user, tt.args.t, tt.args.id); (err != nil) != tt.wantErr {
					t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}
