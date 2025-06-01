package storage

import (
	"context"

	"github.com/sejo412/gophkeeper/internal/models"
)

type Storage interface {
	Init(ctx context.Context) error
	Close() error
	List(ctx context.Context, uid models.UserID) ([]models.RecordsEncrypted, error)
	Get(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) (*models.RecordEncrypted, error)
	Add(ctx context.Context, uid models.UserID, t models.RecordType, record models.RecordEncrypted) error
	Update(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID,
		record models.RecordEncrypted) error
	Delete(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) error
	IsExist(ctx context.Context, user models.UserID, t models.RecordType, id models.ID) (bool, error)
	Users(ctx context.Context) ([]*models.User, error)
	NewUser(ctx context.Context, uid string) (*models.UserID, error)
	IsUserExist(ctx context.Context, uid models.UserID) (bool, error)
}
