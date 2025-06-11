package server

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/internal/helpers"
	"github.com/sejo412/gophkeeper/internal/models"
	"github.com/sejo412/gophkeeper/internal/storage/sqlite"
	"github.com/sejo412/gophkeeper/pkg/certs"
)

// Storage interface supports by server.
type Storage interface {
	// Init creates and fills store with init data.
	Init(ctx context.Context) error
	// Close closes connection.
	Close() error
	// ListAll returns all id and meta by owner.
	ListAll(ctx context.Context, uid models.UserID) (models.RecordsEncrypted, error)
	// List returns all id and meta by owner and type of record.
	List(ctx context.Context, uid models.UserID, t models.RecordType) (models.RecordsEncrypted, error)
	// Get returns encrypted data of Record by owner and ID.
	Get(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) (models.RecordEncrypted, error)
	// Add creates new Record for User.
	Add(ctx context.Context, uid models.UserID, t models.RecordType, record models.RecordEncrypted) error
	// Update updates Record for User by Record ID.
	Update(
		ctx context.Context, uid models.UserID, t models.RecordType, id models.ID,
		record models.RecordEncrypted,
	) error
	// Delete deletes Record.
	Delete(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) error
	// IsExist returns true if Record exists in Storage.
	IsExist(ctx context.Context, user models.UserID, t models.RecordType, id models.ID) (bool, error)
	// Users returns all registered users.
	Users(ctx context.Context) ([]models.User, error)
	// NewUser creates new user in Storage.
	NewUser(ctx context.Context, cn string) (models.UserID, error)
	// IsUserExist return true if user registered in Storage.
	IsUserExist(ctx context.Context, uid models.UserID) (bool, error)
	// GetUserID returns User ID by CommonName.
	GetUserID(ctx context.Context, cn string) (models.UserID, error)
}

func createDatabase(ctx context.Context, dbFile string) error {
	if _, err := os.Create(dbFile); err != nil {
		return fmt.Errorf("could not create database: %w", err)
	}
	store, err := sqlite.New(dbFile)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer func() {
		_ = store.Close()
	}()
	if err = store.Init(ctx); err != nil {
		return fmt.Errorf("could not initialize database: %w", err)
	}
	return nil
}

func createCA(_ context.Context, cert, key string) error {
	req := certs.CertRequest{
		CommonName:  constants.CertCACommonName,
		DNSNames:    nil,
		IPAddresses: nil,
		Emails:      nil,
		IsCA:        true,
	}
	certBytes, keyBytes, err := certs.GenRsaCert(req, certs.CASigner{})
	if err != nil {
		return fmt.Errorf("could not generate CA certificate/key pair: %w", err)
	}
	if err = helpers.SaveRegularFile(key, keyBytes, 0600); err != nil {
		return fmt.Errorf("could not save CA key: %w", err)
	}
	if err = helpers.SaveRegularFile(cert, certBytes, 0644); err != nil {
		return fmt.Errorf("could not save CA certificate: %w", err)
	}
	return nil
}

func createServerCert(_ context.Context, dns []string, ip []net.IP, cert, key, caCert, caKey string) error {
	req := certs.NewCertRequest(
		constants.CertServerCommonName,
		dns,
		nil,
		ip,
		false,
	)
	rsaKey, err := certs.GenRsaKey(constants.KeyBits)
	if err != nil {
		return fmt.Errorf("could not generate server key: %w", err)
	}
	rsaBytes, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	if err != nil {
		return fmt.Errorf("could not marshal server key: %w", err)
	}
	if err = req.Sign(rsaBytes); err != nil {
		return fmt.Errorf("could not sign server certificate request: %w", err)
	}
	signer, err := certs.GetCASigner(caCert, caKey)
	if err != nil {
		return fmt.Errorf("could not get CA signer: %w", err)
	}
	certBytes, _, err := certs.GenRsaCert(*req, signer)
	if err != nil {
		return fmt.Errorf("could not generate server certificate: %w", err)
	}
	if err = helpers.SaveRegularFile(key, rsaBytes, 0600); err != nil {
		return fmt.Errorf("could not save server key: %w", err)
	}
	if err = helpers.SaveRegularFile(cert, certBytes, 0644); err != nil {
		return fmt.Errorf("could not save server certificate: %w", err)
	}
	return nil
}
