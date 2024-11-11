package repo

import (
	"context"

	"github.com/sysarmor/guard/server/internal/model"
)

type NodeRepo interface {
	GetByUniqueID(ctx context.Context, uniqueID string) (*model.Node, error)
	GetByID(ctx context.Context, id int64) (*model.Node, error)
	Create(ctx context.Context, node *model.Node) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, spaceID int64, offset, limit int64) ([]*model.Node, int64, error)
	UpdateLastHeartbeat(ctx context.Context, uniqueID string) error
}
