package http

import (
	"net/http"
	"strings"

	"github.com/PeshawaAziz/url-shortener/internal/domain/auth"
	"github.com/PeshawaAziz/url-shortener/pkg/httputil"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(tokenService auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed Authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims, err := tokenService.ValidateAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired access token"})
			return
		}

		httputil.SetUser(c, claims.UserID)
		httputil.SetTenant(c, claims.TenantID)

		c.Next()
	}
}
