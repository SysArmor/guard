package errors

import (
	"net/http"

	"github.com/sysarmor/guard/server/pkg/errors"
)

const (
	ParamError = 400
)

var (
	ErrSpaceNotFound          = errors.NewWithHTTPCode(http.StatusNotFound, 100001, "space not found")
	ErrNodeNotFound           = errors.NewWithHTTPCode(http.StatusNotFound, 100002, "node not found")
	ErrRoleNotFound           = errors.NewWithHTTPCode(http.StatusNotFound, 100003, "role not found")
	ErrUserNotFound           = errors.NewWithHTTPCode(http.StatusNotFound, 100004, "user not found")
	ErrSpaceNameAlreadyExists = errors.New(100005, "space name already exists")
	ErrUserBanned             = errors.NewWithHTTPCode(http.StatusForbidden, 100006, "user is banned")
	ErrUserAlreadyExists      = errors.New(100007, "user already exists")
)
