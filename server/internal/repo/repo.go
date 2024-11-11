package repo

import "context"

type Repo interface {
	BeginTx(ctx context.Context) (Repo, error)
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error

	Node() NodeRepo
	Role() RoleRepo
	User() UserRepo
	Space() SpaceRepo
}
