package model

// Space is the model of the space
type Space struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

// SpaceUser is the relation of the space and the user
type SpaceUser struct {
	ID        int64 `json:"id"`
	SpaceID   int64 `json:"space_id"`
	UserID    int64 `json:"user_id"`
	CreatedAt int64 `json:"created_at"`
}
