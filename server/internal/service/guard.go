package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/sysarmor/guard/server/internal/repo"
	"github.com/sysarmor/guard/server/pkg/certificate"
)

type Guard interface {
	GetCA(ctx context.Context) []byte
	GetPrincipals(ctx context.Context, uniqueID string) (PrincipalList, error)
	GetNodeByUniqueID(ctx context.Context, uniqueID string) (*Node, error)
	GetKRL(ctx context.Context, uniqueID string) (string, error)
	GetAuthorizedKeys(ctx context.Context, uniqueID string) ([]string, error)

	CreateUser(ctx context.Context, in *CreateUserRequest) (int64, error)
	ListUser(ctx context.Context, in *ListUserRequest) (ListUserResponse, error)
	GetUser(ctx context.Context, id int64) (*GetUserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*GetUserResponse, error)
	BanUser(ctx context.Context, id int64) error
	UpdateUserPublicKey(ctx context.Context, in *UpdateUserPublicKeyRequest) error
	GrantCert(ctx context.Context, in *GrantCertRequest) (*GrantCertResponse, error)

	CreateSpace(ctx context.Context, in *CreateSpaceRequest) (int64, error)
	ListSpace(ctx context.Context) (ListSpaceResponse, error)

	CreateNode(ctx context.Context, in *CreateNodeRequest) (*CreateNodeResponse, error)
	ListNode(ctx context.Context, in *ListNodeRequest) (*ListNodeResponse, error)
	DeleteNode(ctx context.Context, id int64) error
	UpdateLastHeartbeat(ctx context.Context, uniqueID string) error

	CreateRole(ctx context.Context, in *CreateRoleRequest) (int64, error)
	ListRole(ctx context.Context, in *ListRoleRequest) (ListRoleResponse, error)
	DeleteRole(ctx context.Context, roleID int64) error
	AddNodeToRole(ctx context.Context, in *AddNodeToRoleRequest) error
	ListRoleNode(ctx context.Context, in *ListRoleNodeRequest) (ListRoleNodeResponse, error)
	RemoveNodeFromRole(ctx context.Context, in *RemoveNodeFromRoleRequest) error
	AddUserToRole(ctx context.Context, in *AddUserToRoleRequest) error
	ListRoleUser(ctx context.Context, in *ListRoleUserRequest) (ListRoleUserResponse, error)
	RemoveUserFromRole(ctx context.Context, in *RemoveUserFromRoleRequest) error
}

type Config struct {
	CaPassphrase string `yaml:"ca_passphrase"`
	PubKeyPath   string `yaml:"public_key_path"`
	PrivKeyPath  string `yaml:"private_key_path"`
}

func (c *Config) Validate() error {
	if c.PubKeyPath == "" {
		return fmt.Errorf("public key path is required")
	}

	if c.PrivKeyPath == "" {
		return fmt.Errorf("private key path is required")
	}

	if c.CaPassphrase == "" {
		return fmt.Errorf("ca passphrase is required")
	}

	return nil
}

type guard struct {
	publicKey         []byte
	certificateSigner *certificate.Certificate

	repo          repo.Repo
	getPassphrase func(ctx context.Context) string
}

func New(cfg Config, repo repo.Repo) (Guard, error) {
	guard := &guard{
		repo: repo,
	}

	if err := guard.init(&cfg); err != nil {
		return nil, fmt.Errorf("failed to init guard: %w", err)
	}

	return guard, nil
}

func (g *guard) init(cfg *Config) error {
	privateKey, err := os.ReadFile(cfg.PrivKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	publicKey, err := os.ReadFile(cfg.PubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	g.publicKey = publicKey
	g.certificateSigner = certificate.New(privateKey, publicKey)
	g.getPassphrase = func(_ context.Context) string {
		return cfg.CaPassphrase
	}

	return nil
}

// GetCA returns the CA certificate.
func (g *guard) GetCA(ctx context.Context) []byte {
	return g.publicKey
}

// GetPrincipals returns the principals of the node with the given unique id.
func (g *guard) GetPrincipals(ctx context.Context, uniqueID string) (PrincipalList, error) {
	node, err := g.repo.Node().GetByUniqueID(ctx, uniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node by unique id: %w", err)
	}

	roleNodes, err := g.repo.Role().ListRoleNodeByNodeID(ctx, node.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles by node id: %w", err)
	}

	index := make(map[string]*Principals, len(roleNodes))
	principals := make(PrincipalList, 0, len(roleNodes))
	for _, roleNode := range roleNodes {
		users, err := g.repo.Role().ListUserByRoleID(ctx, roleNode.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list users by role id: %w", err)
		}

		var p *Principals
		var ok bool
		if p, ok = index[roleNode.Account]; !ok {
			p = &Principals{
				Role:       roleNode.Account,
				Principals: make([]string, 0),
			}
			index[roleNode.Account] = p
			principals = append(principals, p)
		}

		for _, user := range users {
			p.Principals = append(p.Principals, user.Email)
		}

	}

	return principals, nil
}

// GetNodeByUniqueID returns the node with the given unique id.
func (g *guard) GetNodeByUniqueID(ctx context.Context, uniqueID string) (*Node, error) {
	node, err := g.repo.Node().GetByUniqueID(ctx, uniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node by unique id: %w", err)
	}

	if node == nil {
		return nil, fmt.Errorf("node not found")
	}

	return &Node{
		ID:            node.ID,
		UniqueID:      node.UniqueID,
		Secret:        node.Secret,
		Name:          node.Name,
		Description:   node.Description,
		SpaceID:       node.SpaceID,
		IP:            node.IP,
		LastHeartbeat: node.LastHeartbeat,
		Accounts:      node.Accounts,
		CreatedAt:     node.CreatedAt,
	}, nil
}

// GetKRL returns the key revocation list.
func (g *guard) GetKRL(ctx context.Context, uniqueID string) (string, error) {
	node, err := g.repo.Node().GetByUniqueID(ctx, uniqueID)
	if err != nil {
		return "", fmt.Errorf("failed to get node by unique id: %w", err)
	}

	revokedKeys, err := g.repo.Role().ListRevokedKeys(ctx, node.ID)
	if err != nil {
		return "", fmt.Errorf("failed to list revoked keys: %w", err)
	}

	if len(revokedKeys) == 0 {
		return "", nil
	}

	crl, err := g.certificateSigner.RevokeKeys(revokedKeys...)
	if err != nil {
		return "", fmt.Errorf("failed to revoke keys: %w", err)
	}

	return base64.StdEncoding.EncodeToString(crl), nil
}

// GetAuthorizedKeys returns the public keys of the user with the given unique id.
// If ssh server not support certificate authentication, the public key is used for authentication.
func (g *guard) GetAuthorizedKeys(ctx context.Context, uniqueID string) ([]string, error) {
	node, err := g.repo.Node().GetByUniqueID(ctx, uniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node by unique id: %w", err)
	}

	roleIDs, err := g.repo.Role().ListRoleNodeByNodeID(ctx, node.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list role node by node id: %w", err)
	}

	publicKeys := make([]string, 0)
	publicKeysIndex := make(map[string]struct{})

	for _, roleID := range roleIDs {
		keys, err := g.repo.Role().ListUserPublicKeyByRoleID(ctx, roleID.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list user public key by role id: %w", err)
		}

		for _, key := range keys {
			if _, ok := publicKeysIndex[key]; ok {
				continue
			}

			publicKeys = append(publicKeys, key)
			publicKeysIndex[key] = struct{}{}
		}
	}

	return publicKeys, nil
}
