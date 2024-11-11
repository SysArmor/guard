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

type user struct {
	*baseRepo
}

func NewUser(br *baseRepo) repo.UserRepo {
	return &user{baseRepo: br}
}

// List list users
func (u *user) List(ctx context.Context, offset, limit int64) ([]*model.User, error) {
	rows, err := u.queryContext(ctx, `
		SELECT id, username, email, pub_key, ban, created_at, updated_at
		FROM "user"
		ORDER BY id DESC
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := []*model.User{}
	for rows.Next() {
		user := &model.User{}

		var ban sql.NullBool

		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PubKey, &ban, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.Ban = ban.Bool
		users = append(users, user)
	}

	return users, nil
}

// Create a new user
func (u *user) Create(ctx context.Context, user *model.User) error {
	user.CreatedAt = time.Now().Unix()

	err := u.queryRowContext(ctx, `
		INSERT INTO "user" (username, email, pub_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, user.Username, user.Email, user.PubKey, user.CreatedAt, user.UpdatedAt).
		Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (u *user) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user := &model.User{}

	var ban sql.NullBool

	err := u.queryRowContext(ctx, `
		SELECT id, username, email, pub_key, ban, created_at, updated_at
		FROM "user"
		WHERE id = $1
	`, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.PubKey, &ban, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	user.Ban = ban.Bool

	return user, nil
}

func (u *user) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}

	var ban sql.NullBool
	var updated sql.NullInt64

	err := u.queryRowContext(ctx, `
		SELECT id, username, email, pub_key, ban, created_at, updated_at
		FROM "user"
		WHERE email = $1
	`, email).
		Scan(&user.ID, &user.Username, &user.Email, &user.PubKey, &ban, &user.CreatedAt, &updated)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Ban = ban.Bool
	user.UpdatedAt = updated.Int64

	return user, nil
}

// Get user by username
func (u *user) Ban(ctx context.Context, id int64) error {
	_, err := u.execContext(ctx, `
		UPDATE "user"
		SET ban = true
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	return nil
}

// Update PubKey of the user
func (u *user) UpdatePubKey(ctx context.Context, id int64, pubKey string) error {
	_, err := u.execContext(ctx, `
		UPDATE "user"
		SET pub_key = $1,
			updated_at = $2
		WHERE id = $3
	`, pubKey, time.Now().Unix(), id)

	if err != nil {
		return fmt.Errorf("failed to update pub key: %w", err)
	}

	return nil
}

// GrantCert grants a cert to the user
func (u *user) GrantCert(ctx context.Context, cert *model.UserCert) error {
	cert.CreatedAt = time.Now().Unix()

	err := u.queryRowContext(ctx, `
		INSERT INTO user_cert (user_id, cert, expires_at, is_revoked, created_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, cert.UserID, cert.Cert, cert.ExpiresAt, cert.IsRevoked, cert.CreatedAt).
		Scan(&cert.ID)

	if err != nil {
		return fmt.Errorf("failed to grant cert: %w", err)
	}

	return nil
}

// UpdateCert updates a cert of the user
func (u *user) UpdateCert(ctx context.Context, id int64, cert string) error {
	_, err := u.execContext(ctx, `
		UPDATE user_cert SET cert = $1, updated_at = $2
		WHERE id = $3
	`, cert, time.Now().Unix(), id)

	if err != nil {
		return fmt.Errorf("failed to update cert: %w", err)
	}

	return nil
}

// RevokeCert revokes a cert of the user
func (u *user) RevokeCert(ctx context.Context, id int64) error {
	_, err := u.execContext(ctx, `
		UPDATE user_cert
		SET is_revoked = true,
			updated_at = $1
		WHERE id = $2
	`, time.Now().Unix(), id)

	if err != nil {
		return fmt.Errorf("failed to revoke cert: %w", err)
	}

	return nil
}

// RevokeAllCerts revokes all certs of the user
func (u *user) RevokeAllCerts(ctx context.Context, userID int64) error {
	_, err := u.execContext(ctx, `
		UPDATE user_cert
		SET is_revoked = true,
			updated_at = $1
		WHERE user_id = $2
	`, time.Now().Unix(), userID)

	if err != nil {
		return fmt.Errorf("failed to revoke all certs: %w", err)
	}

	return nil
}

// ListCerts lists all certs of the user
func (u *user) ListCerts(ctx context.Context, userID int64) ([]*model.UserCert, error) {
	rows, err := u.queryContext(ctx, `
		SELECT id, user_id, cert, expires_at, is_revoked, created_at, updated_at
		FROM user_cert
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list certs: %w", err)
	}
	defer rows.Close()

	certs := []*model.UserCert{}
	for rows.Next() {
		cert := &model.UserCert{}
		err := rows.Scan(&cert.ID, &cert.UserID, &cert.Cert, &cert.ExpiresAt, &cert.IsRevoked, &cert.CreatedAt, &cert.UpdateAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cert: %w", err)
		}
		certs = append(certs, cert)
	}

	return certs, nil
}
