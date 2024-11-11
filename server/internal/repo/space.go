package repo

import (
	"context"

	"github.com/sysarmor/guard/server/internal/model"
)

// SpaceRepo is the interface that provides space methods.
type SpaceRepo interface {
	GetByName(ctx context.Context, name string) (*model.Space, error)
	GetByID(ctx context.Context, spaceID int64) (*model.Space, error)
	Create(ctx context.Context, space *model.Space) error
	List(ctx context.Context) ([]*model.Space, error)
}
