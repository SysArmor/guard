package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sysarmor/guard/server/internal/model"
	"github.com/sysarmor/guard/server/internal/service/errors"
)

// CreateRole create a role
func (g *guard) CreateRole(ctx context.Context, in *CreateRoleRequest) (int64, error) {
	space, err := g.repo.Space().GetByID(ctx, in.SpaceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get space by id: %w", err)
	}

	if space == nil {
		slog.Error("space not found", "space_id", in.SpaceID)
		return 0, errors.ErrSpaceNotFound
	}

	role := &model.Role{
		SpaceID:     in.SpaceID,
		Name:        in.Name,
		Description: in.Description,
	}

	if err := g.repo.Role().Create(ctx, role); err != nil {
		return 0, fmt.Errorf("failed to create role: %w", err)
	}

	return role.ID, nil
}

// ListRole list roles
func (g *guard) ListRole(ctx context.Context, in *ListRoleRequest) (ListRoleResponse, error) {
	roles, err := g.repo.Role().List(ctx, in.SpaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	var resp = make(ListRoleResponse, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, &ListRoleVO{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			CreatedAt:   role.CreatedAt,
		})
	}

	return resp, nil
}

// DeleteRole delete a role
func (g *guard) DeleteRole(ctx context.Context, roleID int64) error {
	// delete users from role
	if err := g.repo.Role().RemoveUser(ctx, roleID); err != nil {
		return fmt.Errorf("failed to remove users from role: %w", err)
	}

	// delete nodes from role
	if err := g.repo.Role().RemoveNode(ctx, roleID); err != nil {
		return fmt.Errorf("failed to remove nodes from role: %w", err)
	}

	// delete role
	if err := g.repo.Role().Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	slog.Info("role deleted", "role_id", roleID)
	return nil
}

// AddNodeToRole add a node to a role
func (g *guard) AddNodeToRole(ctx context.Context, in *AddNodeToRoleRequest) (err error) {
	tx, err := g.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			txErr := tx.RollbackTx(ctx)
			if txErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w", txErr)
			}
		} else {
			txErr := tx.CommitTx(ctx)
			if txErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", txErr)
			}
		}
	}()

	role, err := tx.Role().GetByID(ctx, in.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get role by id: %w", err)
	}

	if role == nil {
		return errors.ErrRoleNotFound
	}

	for _, src := range in.Nodes {
		var node *model.Node
		if src.NodeID != 0 {
			node, err = tx.Node().GetByID(ctx, src.NodeID)
			if err != nil {
				return fmt.Errorf("failed to get node by id: %w", err)
			}
		} else {
			node, err = tx.Node().GetByUniqueID(ctx, src.UniqueID)
			if err != nil {
				return fmt.Errorf("failed to get node by unique id: %w", err)
			}
		}

		if node == nil {
			return errors.ErrNodeNotFound
		}

		roleNode, err := tx.Role().GetRoleNodeByRoleIDAndNodeID(ctx, role.ID, node.ID)
		if err != nil {
			return fmt.Errorf("failed to get role node by role id and node id: %w", err)
		}

		if roleNode != nil {
			// already in the role, skip
			continue
		}

		var account string = src.Account
		if account == "" {
			// use the default account, if not specified
			// accounts length must be greater than 0
			account = node.Accounts[0]
		}

		if err := tx.Role().AddNode(ctx, role.ID, node.ID, account); err != nil {
			return fmt.Errorf("failed to add node to role: %w", err)
		}
	}

	return nil
}

// ListRoleNode list role nodes
func (g *guard) ListRoleNode(ctx context.Context, in *ListRoleNodeRequest) (ListRoleNodeResponse, error) {
	nodes, err := g.repo.Role().ListNodeByRoleID(ctx, in.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list role nodes: %w", err)
	}

	var resp = make(ListRoleNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		resp = append(resp, &RoleNodeListVO{
			ID:            node.ID,
			Name:          node.Name,
			Description:   node.Description,
			UniqueID:      node.UniqueID,
			IP:            node.IP,
			LastHeartbeat: node.LastHeartbeat,
			Account:       node.Account,
		})
	}

	return resp, nil
}

// RemoveNodeFromRole remove a node from a role
func (g *guard) RemoveNodeFromRole(ctx context.Context, in *RemoveNodeFromRoleRequest) error {
	if err := g.repo.Role().RemoveNode(ctx, in.RoleID, in.NodeIDs...); err != nil {
		return fmt.Errorf("failed to remove node from role: %w", err)
	}

	return nil
}

// AddUserToRole add a user to a role
func (g *guard) AddUserToRole(ctx context.Context, in *AddUserToRoleRequest) error {
	tx, err := g.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			txErr := tx.RollbackTx(ctx)
			if txErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w", txErr)
			}
		} else {
			txErr := tx.CommitTx(ctx)
			if txErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", txErr)
			}
		}
	}()

	role, err := tx.Role().GetByID(ctx, in.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get role by id: %w", err)
	}

	if role == nil {
		return fmt.Errorf("role not found")
	}

	for _, userID := range in.UserIDs {
		roleUser, err := tx.Role().GetRoleUserByRoleIDAndUserID(ctx, role.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to get role user by role id and user id: %w", err)
		}

		if roleUser != nil {
			// already in the role, skip
			continue
		}

		user, err := tx.User().GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user by id: %w", err)
		}

		if user == nil {
			slog.Error("user not found", "user_id", userID)
			return errors.ErrUserNotFound
		}

		if err := tx.Role().AddUser(ctx, role.ID, user.ID); err != nil {
			return fmt.Errorf("failed to add user to role: %w", err)
		}
	}

	return nil
}

// ListRoleUser list role users
func (g *guard) ListRoleUser(ctx context.Context, in *ListRoleUserRequest) (ListRoleUserResponse, error) {
	users, err := g.repo.Role().ListUserByRoleID(ctx, in.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list role users: %w", err)
	}

	var resp = make(ListRoleUserResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, &RoleUserListVO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		})
	}

	return resp, nil
}

// RemoveUserFromRole remove a user from a role
func (g *guard) RemoveUserFromRole(ctx context.Context, in *RemoveUserFromRoleRequest) error {
	role, err := g.repo.Role().GetByID(ctx, in.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get role by id: %w", err)
	}

	if role == nil {
		return fmt.Errorf("role not found")
	}

	if err := g.repo.Role().RemoveUser(ctx, in.RoleID, in.UserIDs...); err != nil {
		return fmt.Errorf("failed to remove user from role: %w", err)
	}

	return nil
}
