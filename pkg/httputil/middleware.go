package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantStr := c.GetHeader("X-Tenant-ID")
		if tenantStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header required"})
			return
		}
		tenantID, err := uuid.Parse(tenantStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
			return
		}
		SetTenant(c, tenantID)
		c.Next()
	}
}
