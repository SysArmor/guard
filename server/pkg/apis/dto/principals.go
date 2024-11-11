package dto

import "strings"

// PrincipalList represents a list of principals
type PrincipalList []*Principals

func (p PrincipalList) String() string {
	var res []string
	for _, principal := range p {
		res = append(res, principal.String())
	}
	return strings.Join(res, ";")
}

// Principals represents the principals of a role
type Principals struct {
	// Role is the role of the principals
	Role string `json:"role"`
	// Principals are the principals of the role
	// the principals are the email of the users
	Principals []string `json:"principals"`
}

func (p *Principals) String() string {
	return p.Role + ":" + strings.Join(p.Principals, ",")
}
