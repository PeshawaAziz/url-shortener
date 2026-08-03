package httputil

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	UserIDKey   contextKey = "user_id"
)

var (
	ErrMissingTenantID = errors.New("missing X-Tenant-ID header")
	ErrInvalidTenantID = errors.New("invalid tenant ID format")
)

func SetTenant(c *gin.Context, tenantID uuid.UUID) {
	c.Set(string(TenantIDKey), tenantID)
}

func GetTenant(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(string(TenantIDKey))
	if !exists {
		return uuid.Nil, errors.New("tenant ID not found in context")
	}
	tid, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("tenant ID in context is not a UUID")
	}
	return tid, nil
}

func SetUser(c *gin.Context, userID uuid.UUID) {
	c.Set(string(UserIDKey), userID)
}

func GetUser(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(string(UserIDKey))
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	uid, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID in context is not a UUID")
	}
	return uid, nil
}
