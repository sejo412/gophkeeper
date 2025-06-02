package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func NewCert(name string) *Cert {
	return &Cert{
		Name:  name,
		Owner: "root",
		Type:  Unknown,
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
		if err := saveRegularFile(c.PrivateContent.File, c.PrivateContent.Content, 0600); err != nil {
			return fmt.Errorf("failed to save private certificate: %w", err)
		}
	}
	if c.PublicContent.File != "" {
		if err := saveRegularFile(c.PublicContent.File, c.PublicContent.Content, 0644); err != nil {
			return fmt.Errorf("failed to save public certificate: %w", err)
		}
	}
	if c.CAContent.File != "" {
		if err := saveRegularFile(c.CAContent.File, c.CAContent.Content, 0644); err != nil {
			return fmt.Errorf("failed to save CA certificate: %w", err)
		}
	}
	return nil
}

func (c *Cert) Setup(certType CertType, owner string, private, public Content) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Type = certType
	c.Owner = owner
	c.PrivateContent.File = private.File
	c.PrivateContent.Content = private.Content
	c.PublicContent.File = public.File
	c.PublicContent.Content = public.Content
}

func (c *Cert) SetupWithCA(certType CertType, owner string, private, public, ca Content) {
	c.Setup(certType, owner, private, public)
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

func LoadCert(name string, certType CertType, owner, private, public, ca string) (*Cert, error) {
	privateContent, err := os.ReadFile(private)
	if err != nil {
		return nil, fmt.Errorf("failed to read private certificate: %w", err)
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
			certType,
			owner,
			Content{
				File:    private,
				Content: privateContent,
			},
			Content{
				File:    public,
				Content: publicContent,
			},
			Content{
				File:    ca,
				Content: caContent,
			},
		)
		return cert, nil
	}
	cert.Setup(
		certType,
		owner,
		Content{
			File:    private,
			Content: privateContent,
		},
		Content{
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

func CopyCertFrom(c []*Cert, oldName, newName, newDir, newOwner string) *Cert {
	toClone := CertByName(c, oldName)
	if toClone == nil {
		return nil
	}
	newCert := NewCert(newName)
	if toClone.Type == Client {
		newCert.SetupWithCA(toClone.Type, newOwner, toClone.PrivateContent, toClone.PublicContent, toClone.CAContent)
		newCert.CAContent.File = filepath.Join(newDir, filepath.Base(toClone.CAContent.File))
	} else {
		newCert.Setup(toClone.Type, newOwner, toClone.PrivateContent, toClone.PublicContent)
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

// GenRsaCert returns generated certificate and key pair
func GenRsaCert(request CertRequest, signer CaSigner) (certOut, keyOut []byte, err error) {
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
	var signerKey *rsa.PrivateKey

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
	} else {
		block, _ := pem.Decode(signer.CACert)
		blockBytes := block.Bytes
		signerCert, err = x509.ParseCertificate(blockBytes)
		if err != nil {
			return nil, nil, err
		}
		signerKey = signer.CAKey
	}

	certOut, err = x509.CreateCertificate(rand.Reader, template, signerCert, &signerKey.PublicKey, signerKey)
	if err != nil {
		return nil, nil, err
	}

	return certOut, keyOut, nil
}

func genSerialNumber() (serialNumber *big.Int, err error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}
	return serialNumber, nil
}

// saveRegularFile writes file with creating parent dir.
func saveRegularFile(path string, content []byte, perms fs.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	if err := os.WriteFile(path, content, perms); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}
	return nil
}
