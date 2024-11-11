package repo

import (
	"context"

	"github.com/sysarmor/guard/server/internal/model"
)

type RoleRepo interface {
	List(ctx context.Context, spaceID int64) ([]*model.Role, error)
	GetByID(ctx context.Context, id int64) (*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id int64) error

	ListRoleNodeByNodeID(ctx context.Context, nodeID int64) ([]*model.RoleNode, error)
	ListNodeByRoleID(ctx context.Context, roleID int64) ([]*model.RoleNodeView, error)
	AddNode(ctx context.Context, roleID, nodeID int64, account string) error
	// RemoveNode remove node from role, if nodeID is empty, remove all nodes from role
	RemoveNode(ctx context.Context, roleID int64, nodeIDs ...int64) error
	GetRoleNodeByRoleIDAndNodeID(ctx context.Context, roleID, nodeID int64) (*model.RoleNode, error)

	ListUserByRoleID(ctx context.Context, roleID int64) ([]*model.User, error)
	AddUser(ctx context.Context, roleID, userID int64) error
	// RemoveUser remove user from role, if userID is empty, remove all users from role
	RemoveUser(ctx context.Context, roleID int64, userIDs ...int64) error
	RemoveUserByUserID(ctx context.Context, userID int64) error
	GetRoleUserByRoleIDAndUserID(ctx context.Context, roleID, userID int64) (*model.RoleUser, error)

	ListRevokedKeys(ctx context.Context, nodeID int64) ([]int64, error)
}
