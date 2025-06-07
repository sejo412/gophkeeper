package server

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sejo412/gophkeeper/internal/models"
	"github.com/sejo412/gophkeeper/internal/storage/sqlite"
	"github.com/sejo412/gophkeeper/pkg/certs"
)

var (
	testCacheDir     string
	testUserCacheDir string
	testServer       *Server
)

var (
	testUser1CertRequest []byte
	testUser1            = models.User{
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
	testRecordPasswordEncrypted    []byte
	testNewRecordPasswordEncrypted []byte
)

func TestMain(m *testing.M) {
	var err error
	testCacheDir, err = os.MkdirTemp(os.TempDir(), "cache-dir")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(testCacheDir)
	}()
	testUserCacheDir, err = os.MkdirTemp(os.TempDir(), "user-cache-dir")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(testUserCacheDir)
	}()
	store, err := sqlite.New(filepath.Join(testCacheDir, "test.db"))
	if err != nil {
		panic(err)
	}
	testServer = NewServer(Config{
		CacheDir:    testCacheDir,
		PublicPort:  6200,
		PrivatePort: 6201,
	})
	if err = store.Init(context.Background()); err != nil {
		panic(err)
	}
	testServer.config.SetStorage(store)
	if err = testServer.Init(); err != nil {
		panic(err)
	}
	testServer.grpcPublic, err = NewGRPCPublic(*testServer.config)
	if err != nil {
		panic(err)
	}
	testServer.grpcPrivate = NewGRPCPrivate(*testServer.config)
	if err = testServer.Init(); err != nil {
		panic(err)
	}
	userPrivKey, err := certs.GenRsaKey(2048)
	if err != nil {
		panic(err)
	}
	userPrivKeyBytes, err := x509.MarshalPKCS8PrivateKey(userPrivKey)
	if err != nil {
		panic(err)
	}
	csr := certs.NewCertRequest(testUser1.Cn, nil, nil, nil, false)
	if err = csr.Sign(userPrivKeyBytes); err != nil {
		panic(err)
	}
	testUser1CertRequest, err = certs.RequestToBinary(*csr)
	if err != nil {
		panic(err)
	}
	testRecordPasswordEncryptedModel := models.RecordEncrypted{
		Password: models.PasswordEncrypted{
			ID:       1,
			Login:    []byte("testLogin"),
			Password: []byte("testPassword"),
			Meta:     []byte("testMeta"),
		},
	}
	testRecordPasswordEncrypted, _ = json.Marshal(testRecordPasswordEncryptedModel)
	newTestRecordPasswordEncryptedModel := models.RecordEncrypted{
		Password: models.PasswordEncrypted{
			ID:       1,
			Login:    []byte("testLogin"),
			Password: []byte("NEWTestPassword"),
			Meta:     []byte("testMeta"),
		},
	}
	testNewRecordPasswordEncrypted, _ = json.Marshal(newTestRecordPasswordEncryptedModel)
	exitVal := m.Run()
	os.Exit(exitVal)
}
