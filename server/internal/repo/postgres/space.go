package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sysarmor/guard/server/internal/model"
	"github.com/sysarmor/guard/server/internal/repo"
)

type space struct {
	*baseRepo
}

func NewSpace(br *baseRepo) repo.SpaceRepo {
	return &space{
		baseRepo: br,
	}
}

func (s *space) GetByName(ctx context.Context, name string) (*model.Space, error) {
	var space model.Space
	if err := s.queryRowContext(ctx, `SELECT id, name, description, created_at FROM space WHERE name = $1`,
		name).Scan(&space.ID, &space.Name, &space.Description, &space.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get space by name: %v", err)
	}

	return &space, nil
}

func (s *space) GetByID(ctx context.Context, spaceID int64) (*model.Space, error) {
	var space model.Space
	if err := s.queryRowContext(ctx, `SELECT id, name, description, created_at FROM space WHERE id = $1`,
		spaceID).Scan(&space.ID, &space.Name, &space.Description, &space.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get space by id: %v", err)
	}

	return &space, nil
}

// Create creates a new space.
func (s *space) Create(ctx context.Context, space *model.Space) error {
	space.CreatedAt = time.Now().Unix()

	err := s.queryRowContext(ctx,
		`INSERT INTO space (name, description, created_at) VALUES ($1, $2, $3) RETURNING id `,
		space.Name, space.Description, space.CreatedAt).
		Scan(&space.ID)

	if err != nil {
		return fmt.Errorf("failed to create space: %v", err)
	}

	return nil
}

// List returns all spaces.
func (s *space) List(ctx context.Context) ([]*model.Space, error) {
	rows, err := s.queryContext(ctx, `SELECT id, name, description, created_at FROM space`)
	if err != nil {
		return nil, fmt.Errorf("failed to list spaces: %v", err)
	}
	defer rows.Close()

	var spaces []*model.Space
	for rows.Next() {
		var space model.Space
		if err := rows.Scan(&space.ID, &space.Name, &space.Description, &space.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan space: %v", err)
		}
		spaces = append(spaces, &space)
	}

	return spaces, nil
}

// AddUser adds a user to the space.
func (s *space) AddUser(ctx context.Context, spaceID, userID int64) error {
	_, err := s.execContext(ctx, `INSERT INTO space_user (space_id, user_id, created_at) 
	VALUES ($1, $2, $3)`,
		spaceID, userID, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to add user to space: %v", err)
	}

	return nil
}

// RemoveUser removes a user from the space.
func (s *space) RemoveUser(ctx context.Context, spaceID, userID int64) error {
	_, err := s.execContext(ctx, `DELETE FROM space_user 
	WHERE space_id = $1 AND user_id = $2`, spaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from space: %v", err)
	}

	return nil
}

// ListUser returns all users of the space.
func (s *space) ListUser(ctx context.Context, spaceID int64, offset, limit int64) ([]*model.User, int64, error) {
	rows, err := s.queryContext(ctx, `SELECT u.id, u.username, u.email, u.created_at 
	FROM space_user su J
	OIN user u ON su.user_id = u.id 
	WHERE su.space_id = $1 LIMIT $2 OFFSET $3`,
		spaceID, limit, offset)
	if err != nil {
		return nil, 9, fmt.Errorf("failed to list users of space: %v", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, &user)
	}

	var total int64
	err = s.queryRowContext(ctx, `SELECT COUNT(id) FROM space_user WHERE space_id = $1`, spaceID).
		Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users of space: %v", err)
	}

	return users, total, nil
}

// ExistUser checks if the user exists in the space.
func (s *space) ExistUser(ctx context.Context, spaceID, userID int64) (bool, error) {
	var id int64
	err := s.queryRowContext(ctx, `SELECT id FROM space_user WHERE space_id = $1 AND user_id = $2`,
		spaceID, userID).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check if user exists in space: %v", err)
	}

	return true, nil
}
