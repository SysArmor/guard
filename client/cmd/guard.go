package cmd

import (
	"context"
	"fmt"

	flag "github.com/spf13/pflag"
	"github.com/sysarmor/guard/server/pkg/apis"
	"github.com/sysarmor/guard/server/pkg/apis/dto"
)

type guard struct {
	apis.Guard

	address    string
	nodeID     string
	nodeSecret string
}

func (g *guard) PersistentFlags(flagSet *flag.FlagSet) {
	flagSet.StringVar(&g.address, "address", "", "Address of the guard server")
	flagSet.StringVar(&g.nodeID, "node-id", "", "Node ID")
	flagSet.StringVar(&g.nodeSecret, "node-secret", "", "Node secret")
}

// initEndpoint initializes the guard endpoint
func (g *guard) initEndpoint() error {
	if g.Guard == nil {
		if err := g.validate(); err != nil {
			return err
		}

		var err error
		g.Guard, err = apis.NewHTTPGuard(g.address, g.nodeID, g.nodeSecret)
		if err != nil {
			return fmt.Errorf("failed to create guard: %w", err)
		}
	}

	return nil
}

func (g *guard) validate() error {
	if g.address == "" {
		return fmt.Errorf("guard server address is required")
	}

	if g.nodeID == "" {
		return fmt.Errorf("node ID is required")
	}

	if g.nodeSecret == "" {
		return fmt.Errorf("node secret is required")
	}

	return nil
}

func (g *guard) GetCA(ctx context.Context) (string, error) {
	err := g.initEndpoint()
	if err != nil {
		return "", err
	}
	return g.Guard.GetCA(ctx)
}
func (g *guard) GetPrincipals(ctx context.Context) ([]*dto.Principals, error) {
	err := g.initEndpoint()
	if err != nil {
		return nil, err
	}
	return g.Guard.GetPrincipals(ctx)
}

func (g *guard) GetKRL(ctx context.Context) (string, error) {
	err := g.initEndpoint()
	if err != nil {
		return "", err
	}
	return g.Guard.GetKRL(ctx)
}
