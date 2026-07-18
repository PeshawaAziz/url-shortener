package http

import (
	"net/http"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/auth"
	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	tokenService auth.TokenService
}

func NewTokenHandler(tokenService auth.TokenService) *TokenHandler {
	return &TokenHandler{tokenService: tokenService}
}

func (h *TokenHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not provided"})
		return
	}

	newAccessToken, newRefreshToken, expiry, err := h.tokenService.RotateRefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		switch err {
		case auth.ErrRefreshTokenReused:
			c.SetCookie("refresh_token", "", -1, "/", "", true, true)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token reuse detected, all sessions logged out"})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		}
		return
	}

	c.SetCookie(
		"refresh_token",
		newRefreshToken,
		int(time.Until(expiry).Seconds()),
		"/",
		"",   // domain (empty = current domain)
		true, // Secure
		true, // HttpOnly
	)
	c.SetSameSite(http.SameSiteStrictMode)

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"token_type":   "Bearer",
		"expires_in":   int(time.Until(expiry).Seconds()),
	})
}

func (h *TokenHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && refreshToken != "" {
		_ = h.tokenService.RevokeRefreshTokenFamily(c.Request.Context(), refreshToken)
	}
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
