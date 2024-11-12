package apis

import (
	"context"

	"github.com/sysarmor/guard/server/pkg/apis/dto"
)

type Guard interface {
	GetCA(ctx context.Context) (string, error)
	GetPrincipals(ctx context.Context) ([]*dto.Principals, error)
	GetKRL(ctx context.Context) (string, error)
	GetAuthorizedKeys(ctx context.Context) ([]string, error)
}
