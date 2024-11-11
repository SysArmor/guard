package model

type Node struct {
	ID int64 `json:"id"`
	// The node belongs to which space
	SpaceID int64 `json:"space_id"`

	// The name of the node, it's a human readable name
	Name string `json:"name"`

	// The description of the node
	Description string `json:"description"`

	// The unique identifier of the node
	UniqueID string `json:"unique_id"`

	// The secret of the node, it's used to verify the node
	Secret string `json:"secret"`

	// The ip address of the node, maybe a internal ip
	// or external ip
	IP string `json:"ip"`

	// The last heartbeat time of the node
	LastHeartbeat int64 `json:"last_heartbeat"`

	// Accounts is the list of the accounts, it's must exist
	// on the node. default use the first account as the default
	Accounts []string `json:"accounts"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}
