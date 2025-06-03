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
	DBFilename                string = "database.db"
	CertCAPublicFilename      string = "ca.crt"
	CertCAPrivateFilename     string = "ca.key"
	CertCACommonName          string = "GophKeeper Root CA"
	CertCAName                string = "CA"
	CertClientPublicFilename  string = "client.crt"
	CertClientPrivateFilename string = "client.key"
	KeyBits                   int    = 2048
)

var (
	DefaultDNSNames = []string{DefaultServerHost}
	GracefulSignals = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}
)
