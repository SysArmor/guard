package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sysarmor/guard/server/internal/model"
	"github.com/sysarmor/guard/server/internal/service/errors"
)

// CreateUser is the request to create a user
func (g *guard) CreateUser(ctx context.Context, in *CreateUserRequest) (int64, error) {
	user, err := g.repo.User().GetByEmail(ctx, in.Email)
	if err != nil {
		return 0, fmt.Errorf("failed to get user by email: %w", err)
	}

	if user != nil {
		return 0, errors.ErrUserAlreadyExists
	}

	user = &model.User{
		Username: in.Username,
		Email:    in.Email,
		PubKey:   in.PublicKey,
	}

	if err := g.repo.User().Create(ctx, user); err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	slog.Info("create user", "username", in.Username, "email", in.Email)
	return user.ID, nil
}

// UpdateUserPublicKey updates the public key of a user
func (g *guard) UpdateUserPublicKey(ctx context.Context, in *UpdateUserPublicKeyRequest) error {
	user, err := g.repo.User().GetByID(ctx, in.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user by email: %w", err)
	}

	if user == nil {
		return errors.ErrUserNotFound
	}

	user.PubKey = in.PublicKey
	if err := g.repo.User().UpdatePubKey(ctx, in.UserID, in.PublicKey); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// revoke all certs, because the public key is changed
	if err := g.repo.User().RevokeAllCerts(ctx, in.UserID); err != nil {
		return fmt.Errorf("failed to revoke all certs: %w", err)
	}

	slog.Info("update user public key", "username", user.Username)
	return nil
}

// ListUser lists users
func (g *guard) ListUser(ctx context.Context, in *ListUserRequest) (ListUserResponse, error) {
	users, err := g.repo.User().List(ctx, in.Offset(), in.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var resp = make(ListUserResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, &UserListVO{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Ban:       user.Ban,
			CreatedAt: user.CreatedAt,
		})
	}

	return resp, nil
}

// GetUser returns a user by id
func (g *guard) GetUser(ctx context.Context, id int64) (*GetUserResponse, error) {
	user, err := g.repo.User().GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	return &GetUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		PubKey:    user.PubKey,
		Ban:       user.Ban,
		CreatedAt: user.CreatedAt,
		UpdateAt:  user.UpdatedAt,
	}, nil
}

func (g *guard) GetUserByEmail(ctx context.Context, email string) (*GetUserResponse, error) {
	user, err := g.repo.User().GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if user == nil {
		return nil, nil
	}

	return &GetUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		PubKey:    user.PubKey,
		Ban:       user.Ban,
		CreatedAt: user.CreatedAt,
		UpdateAt:  user.UpdatedAt,
	}, nil
}

// BanUser bans a user
func (g *guard) BanUser(ctx context.Context, id int64) error {
	user, err := g.repo.User().GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user by id: %w", err)
	}

	if user == nil {
		return errors.ErrUserNotFound
	}

	// remove user from all roles
	if err := g.repo.Role().RemoveUserByUserID(ctx, id); err != nil {
		return fmt.Errorf("failed to remove user from roles: %w", err)
	}

	if err := g.repo.User().Ban(ctx, id); err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	slog.Info("ban user", "username", user.Username)
	return nil
}

// GrantCert grants a cert to a user
func (g *guard) GrantCert(ctx context.Context, in *GrantCertRequest) (*GrantCertResponse, error) {
	user, err := g.repo.User().GetByID(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	if user.Ban {
		return nil, errors.ErrUserBanned
	}

	userCert := &model.UserCert{
		UserID:    user.ID,
		Cert:      "",
		ExpiresAt: in.Effect,
		IsRevoked: false,
	}

	if err := g.repo.User().GrantCert(ctx, userCert); err != nil {
		return nil, fmt.Errorf("failed to create user cert: %w", err)
	}

	stateDate := in.StartDate
	if stateDate == 0 {
		stateDate = time.Now().Unix()
	}
	endDate := stateDate + in.Effect

	cert, err := g.certificateSigner.SignCert(
		[]byte(g.getPassphrase(ctx)), []byte(user.PubKey),
		uint64(userCert.ID), user.Email, user.Email,
		uint64(stateDate), uint64(endDate))

	if err != nil {
		return nil, fmt.Errorf("failed to sign cert: %w", err)
	}

	if err := g.repo.User().UpdateCert(ctx, userCert.ID, string(cert)); err != nil {
		return nil, fmt.Errorf("failed to update user cert: %w", err)
	}

	return &GrantCertResponse{
		Cert: string(cert),
	}, nil
}
