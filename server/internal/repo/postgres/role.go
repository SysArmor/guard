package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sysarmor/guard/server/internal/model"
	"github.com/sysarmor/guard/server/internal/repo"
)

type role struct {
	*baseRepo
}

// NewRole returns a new RoleRepo
func NewRole(br *baseRepo) repo.RoleRepo {
	return &role{
		baseRepo: br,
	}
}

// Create creates a new role
func (r *role) Create(ctx context.Context, role *model.Role) error {
	role.CreatedAt = time.Now().Unix()

	err := r.queryRowContext(ctx,
		`INSERT INTO role (space_id, name, description, created_at) VALUES ($1, $2, $3, $4) RETURNING id `,
		role.SpaceID, role.Name, role.Description, role.CreatedAt).
		Scan(&role.ID)

	if err != nil {
		return fmt.Errorf("failed to create role: %v", err)
	}

	return nil
}

// GetByID gets a role by id
func (r *role) GetByID(ctx context.Context, id int64) (*model.Role, error) {
	role := &model.Role{}

	var description sql.NullString
	err := r.queryRowContext(ctx, `SELECT id, space_id, name, description, created_at FROM role WHERE id = $1`, id).
		Scan(&role.ID, &role.SpaceID, &role.Name, &description, &role.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}

	role.Description = description.String

	return role, nil
}

// Delete deletes a role
func (r *role) Delete(ctx context.Context, id int64) error {
	_, err := r.execContext(ctx, `DELETE FROM role WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// List lists roles
func (r *role) List(ctx context.Context, spaceID int64) ([]*model.Role, error) {
	rows, err := r.queryContext(ctx, `SELECT id, space_id, name, description, created_at FROM role 
	WHERE space_id = $1`,
		spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var description sql.NullString

	roles := make([]*model.Role, 0)
	for rows.Next() {
		role := &model.Role{}
		err = rows.Scan(&role.ID, &role.SpaceID, &role.Name, &description, &role.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		role.Description = description.String
		roles = append(roles, role)
	}

	return roles, nil
}

// ListRoleNodeByNodeID lists role node by node id
func (r *role) ListRoleNodeByNodeID(ctx context.Context, nodeID int64) ([]*model.RoleNode, error) {
	rows, err := r.queryContext(ctx,
		`SELECT r.id, rn.account, rn.node_id, rn.role_id FROM role r 
		RIGHT JOIN role_node rn ON r.id = rn.role_id 
		WHERE rn.node_id = $1`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]*model.RoleNode, 0)
	for rows.Next() {
		role := &model.RoleNode{}
		err = rows.Scan(&role.ID, &role.Account, &role.NodeID, &role.RoleID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// ListNodeByRoleID lists role node by role id
func (r *role) ListNodeByRoleID(ctx context.Context, roleID int64) ([]*model.RoleNodeView, error) {
	rows, err := r.queryContext(ctx,
		`SELECT n.id, n.name, n.description, n.unique_id, n.secret, n.ip, n.last_heartbeat, n.accounts, n.created_at, n.updated_at, 
		rn.account, rn.created_at
		FROM node n 
		RIGHT JOIN role_node rn ON n.id = rn.node_id 
		WHERE rn.role_id = $1`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]*model.RoleNodeView, 0)
	for rows.Next() {
		node := &model.RoleNodeView{}
		var description sql.NullString
		var lastHeartbeat sql.NullInt64
		var updatedAt sql.NullInt64

		if err := rows.Scan(&node.ID, &node.Name, &description, &node.UniqueID, &node.Secret,
			&node.IP, &lastHeartbeat, pq.Array(&node.Accounts), &node.Node.CreatedAt, &updatedAt,
			&node.Account, &node.CreatedAt,
		); err != nil {
			return nil, err
		}

		if description.Valid {
			node.Description = description.String
		}

		if lastHeartbeat.Valid {
			node.LastHeartbeat = lastHeartbeat.Int64
		}

		if updatedAt.Valid {
			node.UpdatedAt = updatedAt.Int64
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetRoleNodeByRoleIDAndNodeID gets role node by role id and node id
func (r *role) GetRoleNodeByRoleIDAndNodeID(ctx context.Context, roleID, nodeID int64) (*model.RoleNode, error) {
	roleNode := &model.RoleNode{}

	err := r.queryRowContext(ctx,
		`SELECT id, role_id, node_id, account, created_at FROM role_node WHERE role_id = $1 AND node_id = $2`,
		roleID, nodeID).
		Scan(&roleNode.ID, &roleNode.RoleID, &roleNode.NodeID, &roleNode.Account, &roleNode.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get role node by role id and node id: %w", err)
	}

	return roleNode, nil
}

func (r *role) AddNode(ctx context.Context, roleID, nodeID int64, account string) error {
	_, err := r.execContext(ctx, `INSERT INTO role_node (role_id, node_id, account, created_at) 
	VALUES ($1, $2, $3, $4)`,
		roleID, nodeID, account, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to add node to role: %w", err)
	}

	return nil
}

func (r *role) RemoveNode(ctx context.Context, roleID int64, nodeIDs ...int64) error {
	sql := "DELETE FROM role_node WHERE role_id = $1"
	values := []interface{}{roleID}

	if len(nodeIDs) != 0 {
		sql += " AND node_id = ANY($2)"
		values = append(values, pq.Array(nodeIDs))
	}

	_, err := r.execContext(ctx, sql,
		values...)
	if err != nil {
		return fmt.Errorf("failed to remove node from role: %w", err)
	}

	return nil
}

// ListUserByRoleID lists users by role id
func (r *role) ListUserByRoleID(ctx context.Context, roleID int64) ([]*model.User, error) {
	rows, err := r.queryContext(ctx,
		`SELECT u.id, u.username, u.email FROM "user" u 
	RIGHT JOIN role_user ur ON u.id = ur.user_id 
	WHERE ur.role_id = $1`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*model.User, 0)
	for rows.Next() {
		user := &model.User{}
		err = rows.Scan(&user.ID, &user.Username, &user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// AddUser adds user to role
func (r *role) AddUser(ctx context.Context, roleID, userID int64) error {
	_, err := r.execContext(ctx, `INSERT INTO role_user (role_id, user_id, created_at) 
	VALUES ($1, $2, $3)`,
		roleID, userID, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to add user to role: %w", err)
	}

	return nil
}

// RemoveUser removes user from role
func (r *role) RemoveUser(ctx context.Context, roleID int64, userIDs ...int64) error {
	sql := "DELETE FROM role_user WHERE role_id = $1"
	values := []interface{}{roleID}

	if len(userIDs) != 0 {
		sql += " AND user_id = ANY($2)"
		values = append(values, pq.Array(userIDs))
	}
	_, err := r.execContext(ctx, sql,
		values...)
	if err != nil {
		return fmt.Errorf("failed to remove user from role: %w", err)
	}

	return nil
}

// RemoveUserByUserID removes user from role by user id
func (r *role) RemoveUserByUserID(ctx context.Context, userID int64) error {
	_, err := r.execContext(ctx, `DELETE FROM role_user WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from role by user id: %w", err)
	}

	return nil
}

// GetRoleUserByRoleIDAndUserID gets role user by role id and user id
func (r *role) GetRoleUserByRoleIDAndUserID(ctx context.Context, roleID, userID int64) (*model.RoleUser, error) {
	roleUser := &model.RoleUser{}

	err := r.queryRowContext(ctx,
		`SELECT id, role_id, user_id, created_at FROM role_user WHERE role_id = $1 AND user_id = $2`,
		roleID, userID).
		Scan(&roleUser.ID, &roleUser.RoleID, &roleUser.UserID, &roleUser.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get role user by role id and user id: %w", err)
	}

	return roleUser, nil
}

// ListRevokedKeys lists revoked keys
func (r *role) ListRevokedKeys(ctx context.Context, nodeID int64) ([]int64, error) {
	rows, err := r.queryContext(ctx,
		`SELECT uc.id
		FROM user_cert uc
		JOIN role_user ru ON uc.user_id = ru.user_id
		JOIN role_node rn ON ru.role_id = rn.role_id
		WHERE uc.is_revoked = TRUE
		AND rn.node_id = $1 and uc.expires_at < $2`, nodeID, time.Now().Unix())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revokedKeys := make([]int64, 0)
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan revoked key: %w", err)
		}
		revokedKeys = append(revokedKeys, id)
	}

	return revokedKeys, nil
}

// ListUserPublicKeyByRoleID lists user public key by role id
func (r *role) ListUserPublicKeyByRoleID(ctx context.Context, roleID int64) ([]string, error) {
	rows, err := r.queryContext(ctx,
		`SELECT pub_key FROM "user" as u 
		LEFT JOIN role_user ru ON ru.user_id = u.id 
		LEFT JOIN role ON ru.role_id = role.id 
		WHERE role.id = $1`, roleID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]string, 0)
	for rows.Next() {
		var key sql.NullString
		err = rows.Scan(&key)
		if err != nil {
			return nil, fmt.Errorf("failed to scan public key: %w", err)
		}
		keys = append(keys, key.String)
	}

	return keys, nil
}
