package constants

import (
	"os"
	"syscall"
)

const (
	DefaultPublicPort  int    = 3200
	DefaultPrivatePort int    = 3201
	DefaultServerHost  string = "localhost"
)

const (
	DBFilename string = "database.db"
)

const (
	CertCAPublicFilename      string = "ca.crt"
	CertCAPrivateFilename     string = "ca.key"
	CertCACommonName          string = "GophKeeper Root CA"
	CertCAName                string = "CA"
	CertClientPublicFilename  string = "client.crt"
	CertClientPrivateFilename string = "client.key"
	CertServerPublicFilename  string = "server.crt"
	CertServerPrivateFilename string = "server.key"
	CertServerCommonName      string = "GophKeeper Server"
	KeyBits                   int    = 2048
	PemCertType               string = "CERTIFICATE"
	PemKeyType                string = "RSA PRIVATE KEY"
)

var (
	DefaultDNSNames = []string{DefaultServerHost}
	GracefulSignals = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}
)
