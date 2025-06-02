package constants

const (
	DefaultPublicPort  int = 3200
	DefaultPrivatePort int = 3201
)

var (
	DefaultDNSNames = []string{"localhost"}
)

const (
	DBFilename            string = "database.db"
	CertCAPublicFilename  string = "ca.crt"
	CertCAPrivateFilename string = "ca.key"
	CertBits              int    = 2048
	CertCACommonName      string = "GophKeeper Root CA"
)
