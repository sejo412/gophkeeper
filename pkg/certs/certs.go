package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sejo412/gophkeeper/internal/helpers"
)

func NewCert(name string) *Cert {
	return &Cert{
		Name: name,
		Type: Unknown,
		PrivateContent: Content{
			File:    "",
			Content: []byte{},
		},
		PublicContent: Content{
			File:    "",
			Content: []byte{},
		},
		CAContent: Content{
			File:    "",
			Content: []byte{},
		},
		mutex: sync.Mutex{},
	}
}

func (c *Cert) Save() error {
	if c.PrivateContent.File != "" {
		if err := saveCert(c.PrivateContent.File, c.PrivateContent.Content, 0600); err != nil {
			return fmt.Errorf("failed to save private certificate: %w", err)
		}
	}
	if c.PublicContent.File != "" {
		if err := saveCert(c.PublicContent.File, c.PublicContent.Content, 0644); err != nil {
			return fmt.Errorf("failed to save public certificate: %w", err)
		}
	}
	if c.CAContent.File != "" {
		if err := saveCert(c.CAContent.File, c.CAContent.Content, 0644); err != nil {
			return fmt.Errorf("failed to save CA certificate: %w", err)
		}
	}
	return nil
}

func (c *Cert) Setup(certType CertType, private, public Content) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Type = certType
	c.PrivateContent.File = private.File
	c.PrivateContent.Content = private.Content
	c.PublicContent.File = public.File
	c.PublicContent.Content = public.Content
}

func (c *Cert) SetupWithCA(certType CertType, private, public, ca Content) {
	c.Setup(certType, private, public)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.CAContent.File = ca.File
	c.CAContent.Content = ca.Content
}

func (c *Cert) Get() *Cert {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c
}

func NewCertRequest(cn string, dnsNames, emails []string, ip []net.IP, isCA bool) *CertRequest {
	return &CertRequest{
		CommonName:  cn,
		DNSNames:    dnsNames,
		Emails:      emails,
		IPAddresses: ip,
		IsCA:        isCA,
	}
}

func (csr *CertRequest) Sign(key []byte) error {
	var err error
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: csr.CommonName,
		},
		DNSNames:       csr.DNSNames,
		EmailAddresses: csr.Emails,
		IPAddresses:    csr.IPAddresses,
	}
	parsed, err := x509.ParsePKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	csr.Signed, err = x509.CreateCertificateRequest(rand.Reader, &csrTemplate, parsed)
	if err != nil {
		return fmt.Errorf("failed to sign certificate request: %w", err)
	}
	return nil
}

func LoadCert(name string, certType CertType, private, public, ca string) (*Cert, error) {
	var privateContent []byte
	if private != "" {
		var err error
		privateContent, err = os.ReadFile(private)
		if err != nil {
			return nil, fmt.Errorf("failed to read private certificate: %w", err)
		}
	}
	publicContent, err := os.ReadFile(public)
	if err != nil {
		return nil, fmt.Errorf("failed to read public certificate: %w", err)
	}
	cert := NewCert(name)
	if certType == Client && ca != "" {
		caContent, err := os.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		cert.SetupWithCA(
			certType, Content{
				File:    private,
				Content: privateContent,
			}, Content{
				File:    public,
				Content: publicContent,
			}, Content{
				File:    ca,
				Content: caContent,
			},
		)
		return cert, nil
	}
	cert.Setup(
		certType, Content{
			File:    private,
			Content: privateContent,
		}, Content{
			File:    public,
			Content: publicContent,
		},
	)
	return cert, nil
}

func CertByName(c []*Cert, name string) *Cert {
	for _, cert := range c {
		if cert.Name == name {
			return cert.Get()
		}
	}
	return nil
}

func CopyCertFrom(c []*Cert, oldName, newName, newDir string) *Cert {
	toClone := CertByName(c, oldName)
	if toClone == nil {
		return nil
	}
	newCert := NewCert(newName)
	if toClone.Type == Client {
		newCert.SetupWithCA(toClone.Type, toClone.PrivateContent, toClone.PublicContent, toClone.CAContent)
		newCert.CAContent.File = filepath.Join(newDir, filepath.Base(toClone.CAContent.File))
	} else {
		newCert.Setup(toClone.Type, toClone.PrivateContent, toClone.PublicContent)
	}
	newCert.PrivateContent.File = filepath.Join(newDir, filepath.Base(toClone.PrivateContent.File))
	newCert.PublicContent.File = filepath.Join(newDir, filepath.Base(toClone.PublicContent.File))
	return newCert
}

// GenRsaKey returns generated rsa private key
func GenRsaKey(bits int) (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rsa key: %v", err)
	}
	return key, nil
}

// GenRsaCert returns generated certificate and key pair.
func GenRsaCert(request CertRequest, signer CASigner) (certOut, keyOut []byte, err error) {
	serial, err := genSerialNumber()
	if err != nil {
		return
	}
	template := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: request.CommonName,
		},
		SerialNumber:          serial,
		NotBefore:             time.Now(),
		IsCA:                  request.IsCA,
		BasicConstraintsValid: true,
		DNSNames:              request.DNSNames,
		IPAddresses:           request.IPAddresses,
		EmailAddresses:        request.Emails,
	}
	if request.IsCA {
		template.NotAfter = time.Now().AddDate(caYears, 0, 0)
		template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment
	} else {
		template.NotAfter = time.Now().AddDate(nonCaYears, 0, 0)
		template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageContentCommitment
	}

	var signerCert *x509.Certificate
	var signerKey any
	var publicKey any
	if request.IsCA {
		key, err := GenRsaKey(keyBits)
		if err != nil {
			return nil, nil, err
		}
		keyOut, err = x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
		}
		signerCert = template
		signerKey = key
		publicKey = &key.PublicKey
	} else {
		signerCert, err = x509.ParseCertificate(signer.CACert)
		if err != nil {
			return nil, nil, err
		}
		if request.Signed == nil {
			return nil, nil, errors.New("could not find signed csr")
		}
		csr, err := x509.ParseCertificateRequest(request.Signed)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse signed csr: %w", err)
		}
		publicKey = csr.PublicKey

		signerKey = signer.CAKey
	}

	certOut, err = x509.CreateCertificate(rand.Reader, template, signerCert, publicKey, signerKey)
	if err != nil {
		return nil, nil, err
	}

	return certOut, keyOut, nil
}

func GetCASigner(certFile, certKey string) (CASigner, error) {
	res := CASigner{}
	keyBytes, err := os.ReadFile(certKey)
	if err != nil {
		return res, fmt.Errorf("failed to read key file: %w", err)
	}
	key, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return res, fmt.Errorf("failed to parse key file: %w", err)
	}
	res.CAKey = key.(*rsa.PrivateKey)
	res.CACert, err = os.ReadFile(certFile)
	if err != nil {
		return res, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return res, nil
}

func genSerialNumber() (serialNumber *big.Int, err error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}
	return serialNumber, nil
}

// saveCert writes file with creating parent dir.
func saveCert(path string, content []byte, perms fs.FileMode) error {
	return helpers.SaveRegularFile(path, content, perms)
}

func RequestToBinary(req CertRequest) ([]byte, error) {
	res, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed convert request to binary: %w", err)
	}
	return res, nil
}

func BinaryToRequest(req []byte) (CertRequest, error) {
	var request CertRequest
	if err := json.Unmarshal(req, &request); err != nil {
		return CertRequest{}, fmt.Errorf("failed convert request to binary: %w", err)
	}
	return request, nil
}
