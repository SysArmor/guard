package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sysarmor/guard/server/pkg/errors"
	"github.com/sysarmor/guard/server/pkg/signature"
)

const (
	HeaderTimestamp = "X-Timestamp"
	HeaderSignature = "X-Signature"
)

type ctxIn struct {
	secret string

	*gin.Context
}

func (g *Guard) IsAllowedNode(c *gin.Context) {
	nodeID := c.Query("nodeID")
	if nodeID == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("node id is required"))
		return
	}

	timeStamp := c.GetHeader(HeaderTimestamp)
	if timeStamp == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("timestamp is required"))
		return
	}

	sign := c.GetHeader(HeaderSignature)
	if sign == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("signature is required"))
		return
	}

	ctx := c.Request.Context()
	node, err := g.svc.GetNodeByUniqueID(ctx, nodeID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		slog.ErrorContext(ctx, err.Error(), "nodeID", nodeID)
		return
	}

	if node == nil {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("node not found"))
		return
	}

	expectedSign := signature.SimpleSignature(signature.SimpleString(timeStamp), []byte(node.Secret))

	if expectedSign != sign {
		slog.Debug("expected sign: %s, actual sign: %s", expectedSign, sign)
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("signature is invalid"))
		return
	}

	ctxIn := &ctxIn{
		secret:  node.Secret,
		Context: c,
	}

	c.Request = c.Request.WithContext(ctxIn)

	c.Next()
}

type writer struct {
	gin.ResponseWriter

	secret string
}

// Write implements the http.ResponseWriter interface
func (w *writer) Write(b []byte) (int, error) {
	sign := signature.SimpleSignature(signature.SimpleString(b), []byte(w.secret))
	w.Header().Set(HeaderSignature, sign)
	return w.ResponseWriter.Write(b)
}

// Signature is a middleware to sign the response
// with the secret of the node
func (g *Guard) Signature(c *gin.Context) {
	ctx, ok := c.Request.Context().(*ctxIn)
	if !ok {
		return
	}

	c.Writer = &writer{
		ResponseWriter: c.Writer,
		secret:         ctx.secret,
	}

	c.Next()
}

func (g *Guard) UpdateNodeLastHeartbeat(c *gin.Context) {
	c.Next()

	nodeID := c.Query("nodeID")
	if nodeID == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("node id is required"))
		return
	}

	ctx := c.Request.Context()
	if err := g.svc.UpdateLastHeartbeat(ctx, nodeID); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		slog.ErrorContext(ctx, err.Error(), "nodeID", nodeID)
		return
	}
}

func getSpaceID(c *gin.Context) (int64, error) {
	spaceID, err := strconv.ParseInt(c.Param("spaceID"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse space id: %w", err)
	}
	return spaceID, nil
}

func getNodeID(c *gin.Context) (int64, error) {
	nodeID, err := strconv.ParseInt(c.Param("nodeID"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse node id: %w", err)
	}
	return nodeID, nil
}

func getUserID(c *gin.Context) (int64, error) {
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse user id: %w", err)
	}
	return userID, nil
}

func getRoleID(c *gin.Context) (int64, error) {
	roleID, err := strconv.ParseInt(c.Param("roleID"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse role id: %w", err)
	}
	return roleID, nil
}

// response is a helper function to send response to the client
func response(c *gin.Context, data interface{}, err error) {
	if err != nil {
		if err, ok := err.(*errors.Error); ok {
			if err.HTTPCode != nil {
				c.JSON(err.GetHTTPCode(), err)
				return
			}
			c.JSON(http.StatusBadRequest, err)
			return
		}

		slog.Error(err.Error())

		// do not expose the error message to the client
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if data == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, data)
}
