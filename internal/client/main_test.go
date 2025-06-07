package client

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/sejo412/gophkeeper/internal/models"
	"github.com/sejo412/gophkeeper/internal/server"
	"github.com/sejo412/gophkeeper/internal/storage/sqlite"
)

const (
	testPublicPort  = 7200
	testPrivatePort = 7201
)

var (
	testPublicAddress  = net.JoinHostPort("127.0.0.1", strconv.Itoa(testPublicPort))
	testPrivateAddress = net.JoinHostPort("127.0.0.1", strconv.Itoa(testPrivatePort))
)

var testClient *Client

var (
	testCacheDir         string
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
	serverCacheDir, err := os.MkdirTemp(os.TempDir(), "server-cache")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(serverCacheDir)
	}()
	store, err := sqlite.New(filepath.Join(serverCacheDir, "test.db"))
	if err != nil {
		panic(err)
	}
	if err = store.Init(context.Background()); err != nil {
		panic(err)
	}
	testServer := server.NewServer(
		server.Config{
			PublicPort:  testPublicPort,
			PrivatePort: testPrivatePort,
			CacheDir:    serverCacheDir,
			DNSNames:    []string{"localhost"},
		},
	)
	if err = testServer.Init(); err != nil {
		panic(err)
	}
	go func() {
		if er := testServer.Start(); er != nil {
			panic(er)
		}
	}()
	testCacheDir, err = os.MkdirTemp(os.TempDir(), "cache-dir")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(testCacheDir)
	}()

	exitVal := m.Run()
	os.Exit(exitVal)
}
