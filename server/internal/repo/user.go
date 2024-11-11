package repo

import (
	"context"

	"github.com/sysarmor/guard/server/internal/model"
)

type UserRepo interface {
	List(ctx context.Context, offset, limit int64) ([]*model.User, error)
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Ban(ctx context.Context, id int64) error
	UpdatePubKey(ctx context.Context, id int64, pubKey string) error
	GrantCert(ctx context.Context, cert *model.UserCert) error
	UpdateCert(ctx context.Context, id int64, cert string) error
	RevokeCert(ctx context.Context, id int64) error
	RevokeAllCerts(ctx context.Context, userID int64) error
	ListCerts(ctx context.Context, userID int64) ([]*model.UserCert, error)
}
