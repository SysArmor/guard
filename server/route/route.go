package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sysarmor/guard/server/internal/controller"
)

type Route struct {
	server *http.Server

	cc *controller.Guard
}

func New(cc *controller.Guard) *Route {
	r := &Route{
		server: &http.Server{},

		cc: cc,
	}

	r.Register()
	return r
}

func (r *Route) Register() {
	e := gin.Default()
	r.server.Handler = e

	sg := e.Group("/api/v1/guard", r.cc.IsAllowedNode, r.cc.Signature, r.cc.UpdateNodeLastHeartbeat)
	{
		sg.GET("/ca", r.cc.GetCA)
		sg.GET("/principals", r.cc.GetPrincipals)
		sg.GET("/krl", r.cc.GetKRL)
	}

	space := e.Group("/api/v1/guard/space")
	{
		space.GET("", r.cc.ListSpace)
		space.POST("", r.cc.CreateSpace)
	}

	user := e.Group("/api/v1/guard")
	{
		user.POST("/user", r.cc.CreateUser)
		user.GET("/users", r.cc.ListUser)
		user.GET("/user", r.cc.QueryUser)
		user.GET("/user/:userID", r.cc.GetUser)
		user.POST("/user/:userID/ban", r.cc.BanUser)
		user.PUT("/user/:userID/publicKey", r.cc.UpdateUserPublicKey)
		user.POST("/user/:userID/cert", r.cc.GrantCert)
	}

	node := e.Group("/api/v1/guard/space/:spaceID/node")
	{
		node.GET("", r.cc.ListNode)
		node.POST("", r.cc.CreateNode)
		node.DELETE("/:nodeID", r.cc.DeleteNode)
	}

	role := e.Group("/api/v1/guard/space/:spaceID/role")
	{
		role.GET("", r.cc.ListRole)
		role.POST("", r.cc.CreateRole)
		role.DELETE("/:roleID", r.cc.DeleteRole)
		role.POST("/:roleID/node", r.cc.AddNodeToRole)
		role.GET("/:roleID/node", r.cc.ListRoleNode)
		role.POST("/:roleID/node/batch/delete", r.cc.BatchRemoveNodeFromRole)
		role.POST("/:roleID/user", r.cc.AddUserToRole)
		role.GET("/:roleID/user", r.cc.ListRoleUser)
		role.POST("/:roleID/user/batch/delete", r.cc.BatchRemoveUserFromRole)
	}
}

func (r *Route) Run(addr string) error {
	r.server.Addr = addr
	return r.server.ListenAndServe()
}

func (r *Route) Close() error {
	return r.server.Close()
}
