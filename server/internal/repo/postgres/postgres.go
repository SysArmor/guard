package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sysarmor/guard/server/internal/repo"
)

// Config is the configuration for the postgres database
type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.Port == "" {
		c.Port = "5432"
	}

	if c.User == "" {
		return fmt.Errorf("user is required")
	}

	if c.Database == "" {
		return fmt.Errorf("database is required")
	}

	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}

	return nil
}

// New creates a new postgres database connection
func New(ctx context.Context, cfg *Config) (repo.Repo, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	var br baseRepo
	br = baseRepo{
		db: db,

		node:  NewNode(&br),
		role:  NewRole(&br),
		space: NewSpace(&br),
		user:  NewUser(&br),
	}

	return &br, nil
}

type baseRepo struct {
	db *sql.DB
	tx *sql.Tx

	node  repo.NodeRepo
	role  repo.RoleRepo
	space repo.SpaceRepo
	user  repo.UserRepo
}

func (br *baseRepo) Node() repo.NodeRepo {
	return br.node
}

func (br *baseRepo) Role() repo.RoleRepo {
	return br.role
}

func (br *baseRepo) Space() repo.SpaceRepo {
	return br.space
}

func (br *baseRepo) User() repo.UserRepo {
	return br.user
}

func (br *baseRepo) execContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if br.tx != nil {
		return br.tx.ExecContext(ctx, query, args...)
	}

	return br.db.ExecContext(ctx, query, args...)
}

func (br *baseRepo) queryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if br.tx != nil {
		return br.tx.QueryRowContext(ctx, query, args...)
	}

	return br.db.QueryRowContext(ctx, query, args...)
}

func (br *baseRepo) queryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if br.tx != nil {
		return br.tx.QueryContext(ctx, query, args...)
	}

	return br.db.QueryContext(ctx, query, args...)
}

func (br *baseRepo) BeginTx(ctx context.Context) (repo.Repo, error) {
	tx, err := br.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &baseRepo{
		db:    br.db,
		tx:    tx,
		node:  br.node,
		role:  br.role,
		space: br.space,
		user:  br.user,
	}, nil
}

func (br *baseRepo) CommitTx(ctx context.Context) error {
	if br.tx == nil {
		return fmt.Errorf("transaction not started")
	}

	if err := br.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction")
	}

	br.tx = nil
	return nil
}

func (br *baseRepo) RollbackTx(ctx context.Context) error {
	if br.tx == nil {
		return fmt.Errorf("transaction not started")
	}

	if err := br.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction")
	}

	br.tx = nil
	return nil
}
