package certs

import (
	"crypto/rsa"
	"net"
	"sync"
)

type CertType int

const (
	caYears    int = 10
	nonCaYears int = 1
	keyBits    int = 2048
)

const (
	Unknown CertType = iota
	CA
	Server
	Client
)

const (
	UnknownName string = "unknown"
	CAName      string = "ca"
	ServerName  string = "server"
	ClientName  string = "client"
)

type Cert struct {
	Name           string
	Owner          string
	PrivateContent Content
	PublicContent  Content
	CAContent      Content
	Type           CertType
	mutex          sync.Mutex
}

type Content struct {
	File    string
	Content []byte
}

type CertRequest struct {
	CommonName  string
	DNSNames    []string
	IPAddresses []net.IP
	Emails      []string
	IsCA        bool
}

type CaSigner struct {
	CACert []byte
	CAKey  *rsa.PrivateKey
}

func (c CertType) String() string {
	switch c {
	case CA:
		return CAName
	case Server:
		return ServerName
	case Client:
		return ClientName
	default:
		return UnknownName
	}
}
