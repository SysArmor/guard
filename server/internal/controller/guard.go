package controller

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sysarmor/guard/server/internal/service"
	"github.com/sysarmor/guard/server/internal/service/errors"
)

type Guard struct {
	svc service.Guard
}

func New(svc service.Guard) *Guard {
	return &Guard{svc: svc}
}

// @Summary GetCA
// @Description Get CA certificate
// @Tags Guard
// @Param node_id query string true "Node ID"
// @Param X-Timestamp header string true "unix timestamp, seconds"
// @Param X-Signature header string true "signature"
// @Router /api/v1/guard/ca [get]
// @Router /api/v1/guard/ca [get]
// @Success 200 {string} string "CA certificate"
func (g *Guard) GetCA(c *gin.Context) {
	ctx := c.Request.Context()

	ca := g.svc.GetCA(ctx)
	response(c, string(ca), nil)
}

// @Summary GetPrincipals
// @Description Get principals
// @Tags Guard
// @Param node_id query string true "Node ID"
// @Param X-Timestamp header string true "unix timestamp, seconds"
// @Param X-Signature header string true "signature"
// @Success 200 {object} service.PrincipalList "principals"
// @Router /api/v1/guard/principals [get]
func (g *Guard) GetPrincipals(c *gin.Context) {
	nodeID := c.Query("nodeID")
	if nodeID == "" {
		c.AbortWithError(http.StatusBadRequest, errors.ErrNodeNotFound)
		return
	}

	ctx := c.Request.Context()
	principals, err := g.svc.GetPrincipals(ctx, nodeID)
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "node id", nodeID)
		response(c, nil, err)
		return
	}

	if principals == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	response(c, principals, nil)
}

// @Summary GetKRL
// @Description Get revoked key list (KRL)
// @Tags Guard
// @Param node_id query string true "Node ID"
// @Param X-Timestamp header string true "unix timestamp, seconds"
// @Param X-Signature header string true "signature"
// @Success 200 {object} []byte "KRL"
// @Router /api/v1/guard/krl [get]
func (g *Guard) GetKRL(c *gin.Context) {
	nodeID := c.Query("nodeID")
	if nodeID == "" {
		c.AbortWithError(http.StatusBadRequest, errors.ErrNodeNotFound)
		return
	}

	ctx := c.Request.Context()
	krl, err := g.svc.GetKRL(ctx, nodeID)
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "node id", nodeID)
		response(c, nil, err)
		return
	}

	response(c, krl, nil)
}

// @Summary CreateUser
// @Description Create user
// @Tags user
// @Router /api/v1/guard/user [post]
// @Param body body service.CreateUserRequest true "Create user request"
// @Success 200 {object} nil
func (g *Guard) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	id, err := g.svc.CreateUser(ctx, &req)
	response(c, id, err)
}

// @Summary ListUser
// @Description List users
// @Tags user
// @Param page query int true "page"
// @Param limit query int tue "limit"
// @Success 200 {object} service.ListUserResponse
// @Router /api/v1/guard/user [get]
func (g *Guard) ListUser(c *gin.Context) {
	ctx := c.Request.Context()
	req := service.ListUserRequest{}
	if err := c.BindQuery(&req.PageRequest); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	users, err := g.svc.ListUser(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, users, nil)
}

// @Summary GetUser
// @Description Get user
// @Tags user
// @Param userID path int true "User ID"
// @Success 200 {object} service.UserVO
// @Router /api/v1/guard/user/{userID} [get]
func (g *Guard) GetUser(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := getUserID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	user, err := g.svc.GetUser(ctx, id)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, user, nil)
}

// @Summary QueryUser
// @Description Query user
// @Tags user
// @Param email query string true "Email"
// @Success 200 {object} service.UserVO
// @Router /api/v1/guard/user [get]
func (g *Guard) QueryUser(c *gin.Context) {
	ctx := c.Request.Context()
	email := c.Query("email")
	if email == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("email is required"))
		return
	}

	user, err := g.svc.GetUserByEmail(ctx, email)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, user, nil)
}

// @Summary BanUser
// @Description Ban user
// @Tags user
// @Param userID path int true "User ID"
// @Success 200 {object} nil
// @Router /api/v1/guard/user/{userID}/ban [post]
func (g *Guard) BanUser(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := getUserID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := g.svc.BanUser(ctx, id); err != nil {
		slog.ErrorContext(ctx, err.Error())
		response(c, nil, err)
		return
	}

	response(c, nil, nil)
}

// @Summary UpdateUserPublicKey
// @Description Update user public key
// @Tags user
// @Param userID path int true "User ID"
// @Param body body string true "Public key"
// @Success 200 {object} nil
// @Router /api/v1/guard/user/{userID}/publicKey [put]
func (g *Guard) UpdateUserPublicKey(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.UpdateUserPublicKeyRequest
	if err := c.ShouldBindJSON(&req.PublicKey); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var err error
	req.UserID, err = getUserID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	if err := g.svc.UpdateUserPublicKey(ctx, &req); err != nil {
		response(c, nil, err)
		return
	}

	response(c, nil, nil)
}

// @Summary GrantCert
// @Description Grant certificate
// @Tags user
// @Param userID path int true "User ID"
// @Param body body service.GrantCertRequest true "Grant certificate request"
// @Success 200 {object} service.GrantCertResponse
// @Router /api/v1/guard/user/{userID}/cert [post]
func (g *Guard) GrantCert(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.GrantCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var err error
	req.UserID, err = getUserID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	cert, err := g.svc.GrantCert(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, cert, nil)
}

// @Summary CreateSpace
// @Description Create space
// @Tags space
// @Param body body service.CreateSpaceRequest true "Create space request"
// @Success 200 {object} nil
// @Router /api/v1/guard/space [post]
func (g *Guard) CreateSpace(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	id, err := g.svc.CreateSpace(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, id, nil)
}

// @Summary ListSpace
// @Description List spaces
// @Tags space
// @Success 200 {object} service.ListSpaceResponse
// @Router /api/v1/guard/space [get]
func (g *Guard) ListSpace(c *gin.Context) {
	ctx := c.Request.Context()
	spaces, err := g.svc.ListSpace(ctx)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, spaces, nil)
}

// @Summary CreateNode
// @Description Create node
// @Tags node
// @Param spaceID path int true "Space ID"
// @Param body body service.CreateNodeRequest true "Create node request"
// @Success 200 {object} service.CreateNodeResponse
// @Router /api/v1/guard/space/{spaceID}/node [post]
func (g *Guard) CreateNode(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var err error
	req.SpaceID, err = getSpaceID(c)
	if err != nil {
		response(c, nil, err)
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	resp, err := g.svc.CreateNode(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, resp, nil)
}

// @Summary ListNode
// @Description List nodes
// @Tags node
// @Param spaceID path int true "Space ID"
// @Param page query int true "page"
// @Param limit query int true "limit"
// @Success 200 {object} service.ListNodeResponse
// @Router /api/v1/guard/space/{spaceID}/node [get]
func (g *Guard) ListNode(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.ListNodeRequest
	var err error
	if err := c.BindQuery(&req.PageRequest); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	req.SpaceID, err = getSpaceID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	nodes, err := g.svc.ListNode(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, nodes, nil)
}

// @Summary DeleteNode
// @Description Delete node
// @Tags node
// @Param spaceID path int true "Space ID"
// @Param nodeID path int true "Node ID"
// @Success 200 {object} nil
// @Router /api/v1/guard/space/{spaceID}/node/{nodeID} [delete]
func (g *Guard) DeleteNode(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := getNodeID(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := g.svc.DeleteNode(ctx, id); err != nil {
		response(c, nil, err)
		return
	}

	response(c, nil, nil)
}

// @Summary CreateRole
// @Description Create role
// @Tags role
// @Param spaceID path int true "Space ID"
// @Param body body service.CreateRoleRequest true "Create role request"
// @Success 200 {object} nil
func (g *Guard) CreateRole(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var err error
	req.SpaceID, err = getSpaceID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	id, err := g.svc.CreateRole(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, id, nil)
}

// @Summary ListRole
// @Description List roles
// @Tags role
// @Param spaceID path int true "Space ID"
// @Success 200 {object} service.ListRoleResponse
// @Router /api/v1/guard/space/{spaceID}/role [get]
func (g *Guard) ListRole(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.ListRoleRequest
	var err error
	req.SpaceID, err = getSpaceID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	roles, err := g.svc.ListRole(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}
	response(c, roles, err)
}

// @Summary DeleteRole
// @Description Delete role
// @Tags role
// @Param spaceID path int true "Space ID"
// @Param roleID path int true "Role ID"
// @Success 200 {object} nil
// @Router /api/v1/guard/space/{spaceID}/role/{roleID} [delete]
func (g *Guard) DeleteRole(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := getRoleID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := g.svc.DeleteRole(ctx, id); err != nil {
		response(c, nil, err)
		return
	}

	response(c, nil, nil)
}

// @Summary AddNodeToRole
// @Description Add node to role
// @Tags role
// @Param spaceID path int true "Space ID"
// @Param roleID path int true "Role ID"
// @Param body body service.RoleNodeListRequest true "node list"
// @Success 200 {object} nil
// @Router /api/v1/guard/space/{spaceID}/role/{roleID}/node [post]
func (g *Guard) AddNodeToRole(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.AddNodeToRoleRequest
	if err := c.ShouldBindJSON(&req.Nodes); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var err error
	req.RoleID, err = getRoleID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	if err := g.svc.AddNodeToRole(ctx, &req); err != nil {
		response(c, nil, err)
		return
	}

	response(c, nil, nil)
}

// @Summary ListRoleNode
// @Description List role nodes
// @Tags role
// @Param spaceID path int true "Space ID"
// @Param roleID path int true "Role ID"
// @Success 200 {object} service.ListRoleNodeResponse
// @Router /api/v1/guard/space/{spaceID}/role/{roleID}/node [get]
func (g *Guard) ListRoleNode(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.ListRoleNodeRequest
	var err error

	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	req.RoleID, err = getRoleID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	nodes, err := g.svc.ListRoleNode(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, nodes, nil)
}

// @Summary BatchRemoveNodeFromRole
// @Description Batch remove node from role
// @Tags role
// @Param spaceID path int true "Space ID"
// @Param roleID path int true "Role ID"
// @Param body body []int64 true "Node IDs"
// @Success 200 {object} nil
// @Router /api/v1/guard/space/{spaceID}/role/{roleID}/node/batch/delete [post]
func (g *Guard) BatchRemoveNodeFromRole(c *gin.Context) {
	ctx := c.Request.Context()
	var req = service.RemoveNodeFromRoleRequest{
		NodeIDs: make([]int64, 0),
	}
	var err error
	if err := c.BindJSON(&req.NodeIDs); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	req.RoleID, err = getRoleID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	if err := g.svc.RemoveNodeFromRole(ctx, &req); err != nil {
		response(c, nil, err)
		return
	}

	response(c, nil, nil)
}

// @Summary AddUserToRole
// @Description Add user to role
// @Tags role
// @Param spaceID path int true "Space ID"
// @Param roleID path int true "Role ID"
// @Param body body []int64 true "User IDs"
// @Success 200 {object} nil
// @Router /api/v1/guard/space/{spaceID}/role/{roleID}/user [post]
func (g *Guard) AddUserToRole(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.AddUserToRoleRequest
	if err := c.ShouldBindJSON(&req.UserIDs); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var err error
	req.RoleID, err = getRoleID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	if err := g.svc.AddUserToRole(ctx, &req); err != nil {
		response(c, nil, err)
		return
	}

	response(c, nil, nil)
}

// @Summary ListRoleUser
// @Description List role users
// @Tags role
// @Param spaceID path int true "Space ID"
// @Param roleID path int true "Role ID"
// @Success 200 {object} service.ListRoleUserResponse
// @Router /api/v1/guard/space/{spaceID}/role/{roleID}/user [get]
func (g *Guard) ListRoleUser(c *gin.Context) {
	ctx := c.Request.Context()
	var req service.ListRoleUserRequest
	var err error
	req.RoleID, err = getRoleID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	users, err := g.svc.ListRoleUser(ctx, &req)
	if err != nil {
		response(c, nil, err)
		return
	}

	response(c, users, nil)
}

// @Summary BatchRemoveUserFromRole
// @Description Batch remove user from role
// @Tags role
// @Param spaceID path int true "Space ID"
// @Param roleID path int true "Role ID"
// @Param body body []int64 true "User IDs"
// @Success 200 {object} nil
// @Router /api/v1/guard/space/{spaceID}/role/{roleID}/user/batch/delete [post]
func (g *Guard) BatchRemoveUserFromRole(c *gin.Context) {
	ctx := c.Request.Context()
	var req = service.RemoveUserFromRoleRequest{
		UserIDs: make([]int64, 0),
	}
	var err error
	if err := c.ShouldBind(&req.UserIDs); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	req.RoleID, err = getRoleID(c)
	if err != nil {
		response(c, nil, err)
		return
	}

	if err := req.Validate(); err != nil {
		response(c, nil, err)
		return
	}

	if err := g.svc.RemoveUserFromRole(ctx, &req); err != nil {
		response(c, nil, err)
		return
	}

	response(c, nil, nil)
}
