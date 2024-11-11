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

type node struct {
	*baseRepo
}

func NewNode(br *baseRepo) repo.NodeRepo {
	return &node{
		baseRepo: br,
	}
}

func (n *node) GetByUniqueID(ctx context.Context, uniqueID string) (*model.Node, error) {
	return n.scan(n.queryRowContext(ctx, `SELECT id, space_id, name, description, unique_id, secret, ip, 
	last_heartbeat, accounts, created_at, updated_at 
	FROM node WHERE unique_id = $1`, uniqueID))
}

func (n *node) scan(row *sql.Row) (*model.Node, error) {
	node := &model.Node{}

	var description sql.NullString
	var lastHeartbeat sql.NullInt64
	var updatedAt sql.NullInt64

	err := row.Scan(&node.ID, &node.SpaceID, &node.Name, &description, &node.UniqueID, &node.Secret,
		&node.IP, &lastHeartbeat, pq.Array(&node.Accounts), &node.CreatedAt, &updatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan node: %w", err)
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

	return node, nil
}

// GetByID returns a node by id.
func (n *node) GetByID(ctx context.Context, id int64) (*model.Node, error) {
	return n.scan(n.queryRowContext(ctx, `SELECT id, space_id, name, description, unique_id, secret, ip,
	last_heartbeat, accounts, created_at, updated_at FROM node WHERE id = $1`, id))
}

// Create creates a new node.
func (n *node) Create(ctx context.Context, node *model.Node) error {
	node.CreatedAt = time.Now().Unix()

	err := n.queryRowContext(ctx,
		`INSERT INTO node (space_id, name, description, unique_id, secret, ip, accounts, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		node.SpaceID, node.Name, node.Description, node.UniqueID, node.Secret, node.IP, pq.Array(node.Accounts), node.CreatedAt).
		Scan(&node.ID)

	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	return nil
}

// Delete deletes a node.
func (n *node) Delete(ctx context.Context, id int64) error {
	_, err := n.execContext(ctx, `DELETE FROM node WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return nil
}

// List returns all nodes.
func (n *node) List(ctx context.Context, spaceID int64, offset, limit int64) ([]*model.Node, int64, error) {
	rows, err := n.queryContext(ctx, `SELECT id, space_id, name, description, unique_id, secret, ip, 
	last_heartbeat, accounts, created_at, updated_at FROM node WHERE space_id = $1 LIMIT $2 OFFSET $3`, spaceID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*model.Node
	for rows.Next() {
		node := &model.Node{}
		var description sql.NullString
		var lastHeartbeat sql.NullInt64
		var updatedAt sql.NullInt64

		if err := rows.Scan(&node.ID, &node.SpaceID, &node.Name, &description, &node.UniqueID, &node.Secret,
			&node.IP, &lastHeartbeat, pq.Array(&node.Accounts), &node.CreatedAt, &updatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan node: %w", err)
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

	var total int64
	if err := n.queryRowContext(ctx, `SELECT COUNT(id) FROM node WHERE space_id = $1`, spaceID).
		Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count nodes: %w", err)
	}

	return nodes, total, nil
}

func (n *node) UpdateLastHeartbeat(ctx context.Context, uniqueID string) error {
	_, err := n.execContext(ctx, `UPDATE node SET last_heartbeat = $1 WHERE unique_id = $2`, time.Now().Unix(), uniqueID)
	if err != nil {
		return fmt.Errorf("failed to update last heartbeat: %w", err)
	}

	return nil
}
