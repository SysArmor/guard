package model

// User is the model of the user
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	PubKey    string `json:"pub_key"`
	Ban       bool   `json:"ban"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// UserCert is the model of the user cert
type UserCert struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Cert   string `json:"cert"`
	// ExpiresAt is the time when the cert will be expired
	// if it is 0, it means the cert will never be expired
	ExpiresAt int64 `json:"expires_at"`
	// IsRevoked is the flag to indicate whether the cert is revoked
	IsRevoked bool  `json:"is_revoked"`
	CreatedAt int64 `json:"created_at"`
	UpdateAt  int64 `json:"updated_at"`
}
