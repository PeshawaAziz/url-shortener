package http

import (
	"net/http"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/auth"
	"github.com/PeshawaAziz/url-shortener/internal/domain/user"
	"github.com/PeshawaAziz/url-shortener/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *auth.UserAuthService
}

func NewAuthHandler(authService *auth.UserAuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	tenantID, err := httputil.GetTenant(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID required"})
		return
	}

	var body struct {
		Email       string `json:"email" binding:"required,email"`
		DisplayName string `json:"display_name"`
		Password    string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.authService.Register(c.Request.Context(), auth.RegisterInput{
		TenantID:    tenantID,
		Email:       body.Email,
		DisplayName: body.DisplayName,
		Password:    body.Password,
	})
	if err != nil {
		switch err {
		case user.ErrEmailAlreadyTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		}
		return
	}

	resp := gin.H{"user_id": output.User.ID.String()}
	if output.VerificationToken != "" {
		// !!! In production, this token would be sent via email.
		// !!! For development, we return it in the response.
		resp["verification_token"] = output.VerificationToken
		resp["message"] = "user registered. Please verify your email."
	} else {
		resp["message"] = "user registered and verified."
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	tenantID, err := httputil.GetTenant(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID required"})
		return
	}

	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.authService.Login(c.Request.Context(), auth.LoginInput{
		TenantID: tenantID,
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		switch err {
		case user.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		case user.ErrAccountLocked:
			c.JSON(http.StatusLocked, gin.H{"error": "account temporarily locked"})
		case user.ErrEmailNotVerified:
			c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		}
		return
	}

	c.SetCookie(
		"refresh_token",
		output.RefreshToken,
		int(time.Until(output.ExpiresAt).Seconds()),
		"/",
		"",
		true, // Secure
		true, // HttpOnly
	)
	c.SetSameSite(http.SameSiteStrictMode)

	c.JSON(http.StatusOK, gin.H{
		"access_token": output.AccessToken,
		"token_type":   "Bearer",
		"expires_in":   int(time.Until(output.AccessExpiry).Seconds()),
		"user_id":      output.User.ID.String(),
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	err := h.authService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired verification token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}
