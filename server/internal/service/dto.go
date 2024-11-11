package service

import (
	"fmt"

	"github.com/sysarmor/guard/server/internal/service/errors"
	"github.com/sysarmor/guard/server/pkg/apis/dto"
	err "github.com/sysarmor/guard/server/pkg/errors"
)

type PrincipalList = dto.PrincipalList
type Principals = dto.Principals

type Node struct {
	ID            int64    `json:"id"`
	UniqueID      string   `json:"unique_id"`
	Secret        string   `json:"secret"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	SpaceID       int64    `json:"space_id"`
	IP            string   `json:"ip"`
	LastHeartbeat int64    `json:"last_heartbeat"`
	Accounts      []string `json:"accounts"`
	CreatedAt     int64    `json:"created_at"`
	UpdatedAt     int64    `json:"updated_at"`
}

type PageRequest struct {
	// Page is the page number, start from 1
	Page int64 `form:"page"`
	// Limit is the number of items per page,
	// must be less than or equal to 1000
	Limit int64 `form:"limit"`
}

func (pr *PageRequest) Validate() error {
	if pr.Page < 0 {
		pr.Page = 0
	}
	if pr.Limit <= 0 {
		return err.New(errors.ParamError, "limit must be greater than 0")
	}
	return nil
}

func (pr *PageRequest) Offset() int64 {
	return (pr.Page - 1) * pr.Limit
}

// ==== Space ====
type CreateSpaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (cs *CreateSpaceRequest) Validate() error {
	if cs.Name == "" {
		return err.New(errors.ParamError, "name is required")
	}
	return nil
}

type ListSpaceVO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

type ListSpaceResponse []*ListSpaceVO

type AddUserToSpaceRequest struct {
	SpaceID int64 `json:"-"`
	// UserIDs is the user id list, at
	// least one user id is required
	UserIDs []int64 `json:"user_ids"`
}

func (a *AddUserToSpaceRequest) Validate() error {
	if len(a.UserIDs) == 0 {
		return err.New(errors.ParamError, "at least one user id is required")
	}
	if a.SpaceID <= 0 {
		return err.New(errors.ParamError, "space id is required")
	}
	return nil
}

// ==== User ====
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	// PublicKey is the public key of the ssh key
	PublicKey string `json:"public_key"`
}

func (cur *CreateUserRequest) Validate() error {
	if cur.Username == "" {
		return err.New(errors.ParamError, "username is required")
	}
	if cur.Email == "" {
		return err.New(errors.ParamError, "email is required")
	}
	if cur.PublicKey == "" {
		return err.New(errors.ParamError, "public key is required")
	}
	return nil
}

type UserListVO struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	// Ban is the status of the user
	// If it is true, the user is banned
	Ban       bool  `json:"ban"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

type ListUserRequest struct {
	PageRequest
}

type ListUserResponse []*UserListVO

type UpdateUserPublicKeyRequest struct {
	UserID    int64  `json:"user_id"`
	PublicKey string `json:"public_key"`
}

func (uupr *UpdateUserPublicKeyRequest) Validate() error {
	if uupr.UserID <= 0 {
		return err.New(errors.ParamError, "user id is required")
	}
	if uupr.PublicKey == "" {
		return err.New(errors.ParamError, "public key is required")
	}
	return nil
}

type UserVO struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	PubKey    string `json:"public_key"`
	Ban       bool   `json:"ban"`
	CreatedAt int64  `json:"created_at"`
	UpdateAt  int64  `json:"updated_at"`
}

type GetUserResponse UserVO

// ==== Node ====

type CreateNodeRequest struct {
	SpaceID int64 `json:"-"`
	// Name is the name of the node
	Name        string `json:"name"`
	Description string `json:"description"`
	// IP is the ip address of the node
	IP string `json:"ip"`
	// Accounts is the account list of the node, it should
	// is account from the machine. If it is empty, the default
	// account is root.
	Accounts []string `json:"accounts"`
}

func (cnr *CreateNodeRequest) Validate() error {
	if cnr.SpaceID <= 0 {
		return err.New(errors.ParamError, "space id is required")
	}
	if cnr.Name == "" {
		return err.New(errors.ParamError, "name is required")
	}
	if cnr.IP == "" {
		return err.New(errors.ParamError, "ip is required")
	}

	return nil
}

type CreateNodeResponse struct {
	ID int64 `json:"id"`
	// UniqueID is the unique id of the node
	UniqueID string `json:"unique_id"`

	// Secret is the secret of the node, it only show once
	// when the node is created
	Secret string `json:"secret"`
}

type ListNodeRequest struct {
	PageRequest

	SpaceID int64 `json:"-"`
}

func (lnr *ListNodeRequest) Validate() error {
	if err := lnr.PageRequest.Validate(); err != nil {
		return err
	}

	if lnr.SpaceID <= 0 {
		return err.New(errors.ParamError, "space id is required")
	}
	return lnr.PageRequest.Validate()
}

type ListNodeVO struct {
	ID            int64    `json:"id"`
	UniqueID      string   `json:"unique_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	IP            string   `json:"ip"`
	Accounts      []string `json:"accounts"`
	LastHeartbeat int64    `json:"last_heartbeat"`
	CreatedAt     int64    `json:"created_at"`
}

type ListNodeResponse struct {
	Total int64         `json:"total"`
	Nodes []*ListNodeVO `json:"nodes"`
}

// ==== Role ====
type CreateRoleRequest struct {
	SpaceID     int64  `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (crr *CreateRoleRequest) Validate() error {
	if crr.SpaceID <= 0 {
		return err.New(errors.ParamError, "space id is required")
	}
	if crr.Name == "" {
		return err.New(errors.ParamError, "name is required")
	}
	return nil
}

type ListRoleRequest struct {
	SpaceID int64 `json:"space_id"`
}

func (lrr *ListRoleRequest) Validate() error {
	if lrr.SpaceID <= 0 {
		return fmt.Errorf("space id is required")
	}
	return nil
}

type ListRoleVO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

type ListRoleResponse []*ListRoleVO

type AddNodeToRoleRequest struct {
	RoleID int64 `json:"role_id"`
	// Nodes is the node list, at least one node is required
	Nodes RoleNodeListRequest `json:"nodes"`
}

type RoleNodeListRequest []RoleNodeRequest

type RoleNodeRequest struct {
	// NodeID is the node id, one of node id and unique id is required
	// If both are provided, the node id is used
	NodeID int64 `json:"node_id"`
	// UniqueID is the unique id of the node
	UniqueID string `json:"unique_id"`
	// Account is the account of the node, must from the node account list
	Account string `json:"account"`
}

func (ar *AddNodeToRoleRequest) Validate() error {
	if ar.RoleID <= 0 {
		return err.New(errors.ParamError, "role id is required")
	}

	if len(ar.Nodes) == 0 {
		return err.New(errors.ParamError, "at least one node is required")
	}

	for i, node := range ar.Nodes {
		if node.NodeID <= 0 && node.UniqueID == "" {
			return err.New(errors.ParamError,
				fmt.Sprintf("node id or unique id is required for node %d", i))
		}
	}
	return nil
}

// ListRoleNodeRequest list role node request
type ListRoleNodeRequest struct {
	RoleID int64 `json:"role_id"`
}

func (lrnr *ListRoleNodeRequest) Validate() error {
	if lrnr.RoleID <= 0 {
		return err.New(errors.ParamError, "role id is required")
	}
	return nil
}

type RoleNodeListVO struct {
	ID            int64  `json:"id"`
	UniqueID      string `json:"unique_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	IP            string `json:"ip"`
	Account       string `json:"account"`
	LastHeartbeat int64  `json:"last_heartbeat"`
}

type ListRoleNodeResponse []*RoleNodeListVO

type RemoveNodeFromRoleRequest struct {
	RoleID  int64   `json:"-"`
	NodeIDs []int64 `json:"node_ids"`
}

func (rnfr *RemoveNodeFromRoleRequest) Validate() error {
	if rnfr.RoleID <= 0 {
		return err.New(errors.ParamError, "role id is required")
	}
	if len(rnfr.NodeIDs) == 0 {
		return err.New(errors.ParamError, "at least one node id is required")
	}
	return nil
}

type AddUserToRoleRequest struct {
	// UserIDs is the user id list, at least one user id is required
	UserIDs []int64 `json:"user_ids"`
	RoleID  int64   `json:"-"`
}

func (aur *AddUserToRoleRequest) Validate() error {
	if aur.RoleID <= 0 {
		return err.New(errors.ParamError, "role id is required")
	}

	if len(aur.UserIDs) == 0 {
		return err.New(errors.ParamError, "at least one user id is required")
	}

	return nil
}

type ListRoleUserRequest struct {
	RoleID int64 `json:"role_id"`
}

func (lrur *ListRoleUserRequest) Validate() error {
	if lrur.RoleID <= 0 {
		return err.New(errors.ParamError, "role id is required")
	}
	return nil
}

type RoleUserListVO struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ListRoleUserResponse []*RoleUserListVO

type RemoveUserFromRoleRequest struct {
	RoleID  int64   `json:"-"`
	UserIDs []int64 `json:"user_ids"`
}

func (rur *RemoveUserFromRoleRequest) Validate() error {
	if len(rur.UserIDs) == 0 {
		return err.New(errors.ParamError, "at least one user id is required")
	}
	if rur.RoleID <= 0 {
		return err.New(errors.ParamError, "role id is required")
	}
	return nil
}

type GrantCertRequest struct {
	UserID int64 `json:"-"`
	// Effect in seconds
	Effect int64 `json:"effect"`

	// StartDate is the start time of the certificate
	// If it is 0, it means the current time
	StartDate int64 `json:"start_date"`
}

func (scr *GrantCertRequest) Validate() error {
	if scr.UserID <= 0 {
		return err.New(errors.ParamError, "user id is required")
	}
	if scr.Effect <= 0 {
		return err.New(errors.ParamError, "effect must be greater than 0")
	}
	return nil
}

type GrantCertResponse struct {
	// Cert is the certificate content
	Cert string `json:"cert"`
}
