package service

import (
	"context"
	"fmt"

	"github.com/sysarmor/guard/server/internal/model"
	"github.com/sysarmor/guard/server/internal/service/errors"
)

// CreateSpace is the request to create a space
func (g *guard) CreateSpace(ctx context.Context, in *CreateSpaceRequest) (int64, error) {
	space, err := g.repo.Space().GetByName(ctx, in.Name)
	if err != nil {
		return 0, fmt.Errorf("failed to get space by name: %w", err)
	}

	if space != nil {
		return 0, errors.ErrSpaceNameAlreadyExists
	}

	space = &model.Space{
		Name:        in.Name,
		Description: in.Description,
	}

	if err := g.repo.Space().Create(ctx, space); err != nil {
		return 0, fmt.Errorf("failed to create space: %w", err)
	}

	return space.ID, nil
}

// ListSpace is the request to list spaces
func (g *guard) ListSpace(ctx context.Context) (ListSpaceResponse, error) {
	spaces, err := g.repo.Space().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list spaces: %w", err)
	}

	response := ListSpaceResponse{}
	for _, space := range spaces {
		response = append(response, &ListSpaceVO{
			ID:          space.ID,
			Name:        space.Name,
			Description: space.Description,
			CreatedAt:   space.CreatedAt,
		})
	}

	return response, nil
}
