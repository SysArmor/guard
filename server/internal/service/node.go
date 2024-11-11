package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sysarmor/guard/server/internal/model"
	"github.com/sysarmor/guard/server/internal/service/errors"
	"github.com/sysarmor/guard/server/pkg/helper"
)

const (
	defaultIDLength     = 16
	defaultSecretLength = 32

	defaultAccount = "root"
)

// CreateNode create a node
func (g *guard) CreateNode(ctx context.Context, in *CreateNodeRequest) (*CreateNodeResponse, error) {
	if len(in.Accounts) == 0 {
		in.Accounts = append(in.Accounts, defaultAccount)
	}

	space, err := g.repo.Space().GetByID(ctx, in.SpaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space by id: %w", err)
	}

	if space == nil {
		slog.Error("space not found", "space_id", in.SpaceID)
		return nil, errors.ErrSpaceNotFound
	}

	node := &model.Node{
		SpaceID:     in.SpaceID,
		Name:        in.Name,
		Description: in.Description,
		UniqueID:    helper.RandString(defaultIDLength),
		Secret:      helper.RandString(defaultSecretLength),
		IP:          in.IP,
		Accounts:    in.Accounts,
	}

	if err := g.repo.Node().Create(ctx, node); err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	return &CreateNodeResponse{
		ID:       node.ID,
		UniqueID: node.UniqueID,
		Secret:   node.Secret,
	}, nil
}

// ListNode list nodes
func (g *guard) ListNode(ctx context.Context, in *ListNodeRequest) (*ListNodeResponse, error) {
	nodes, total, err := g.repo.Node().List(ctx, in.SpaceID,
		in.PageRequest.Offset(), in.PageRequest.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var resp []*ListNodeVO
	for _, node := range nodes {
		resp = append(resp, &ListNodeVO{
			ID:            node.ID,
			UniqueID:      node.UniqueID,
			Name:          node.Name,
			Description:   node.Description,
			IP:            node.IP,
			LastHeartbeat: node.LastHeartbeat,
			Accounts:      node.Accounts,
			CreatedAt:     node.CreatedAt,
		})
	}

	return &ListNodeResponse{
		Total: total,
		Nodes: resp,
	}, nil
}

// DeleteNode delete a node
func (g *guard) DeleteNode(ctx context.Context, id int64) error {
	// remove the node from all roles
	roles, err := g.repo.Role().ListRoleNodeByNodeID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to list role node by node id: %w", err)
	}

	for _, role := range roles {
		if err := g.repo.Role().RemoveNode(ctx, role.ID, id); err != nil {
			return fmt.Errorf("failed to remove node from role: %w", err)
		}
	}

	if err := g.repo.Node().Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return nil
}

func (g *guard) UpdateLastHeartbeat(ctx context.Context, uniqueID string) error {
	if err := g.repo.Node().UpdateLastHeartbeat(ctx, uniqueID); err != nil {
		return fmt.Errorf("failed to update last heartbeat: %w", err)
	}

	return nil
}
