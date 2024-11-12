package apis

import (
	"context"

	"github.com/sysarmor/guard/server/pkg/apis/dto"
)

type FakeGuard struct{}

func (g *FakeGuard) GetCA(ctx context.Context) (string, error) {
	return "fake-ca", nil
}

func (g *FakeGuard) GetPrincipals(ctx context.Context) ([]*dto.Principals, error) {
	return []*dto.Principals{
		{
			Role:       "fake-role",
			Principals: []string{"fake-principal", "fake"},
		},
		{
			Role:       "fake-role1",
			Principals: []string{"fake-principal1"},
		},
	}, nil
}

func (g *FakeGuard) GetKRL(ctx context.Context) (string, error) {
	return "", nil
}

func (g *FakeGuard) GetAuthorizedKeys(ctx context.Context) ([]string, error) {
	return []string{
		"fake-authorized-key",
		"fake-authorized",
	}, nil
}
