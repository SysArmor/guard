package model

// Role is the model of the role
type Role struct {
	ID          int64  `json:"id"`
	SpaceID     int64  `json:"space_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

// RoleNode is the relation of the role and the node
// It's subset of the space_node
type RoleNode struct {
	ID     int64 `json:"id"`
	RoleID int64 `json:"role_id"`
	NodeID int64 `json:"node_id"`
	// Account is the account of the node, user will use this account
	// to access the node
	Account   string `json:"account"`
	CreatedAt int64  `json:"created_at"`
}

// RoleUser is the relation of the role and the user
// It's subset of the space_user. Only the user who has the role
// can access the node in the role.
type RoleUser struct {
	ID        int64 `json:"id"`
	RoleID    int64 `json:"role_id"`
	UserID    int64 `json:"user_id"`
	CreatedAt int64 `json:"created_at"`
}

type RoleNodeView struct {
	Node

	Account   string `json:"account"`
	CreatedAt int64  `json:"created_at"`
}
