package http

import (
	"github.com/PeshawaAziz/url-shortener/internal/domain/url"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RedirectHandler struct {
	redirectSvc *url.RedirectService
	passwordSvc *url.PasswordService
}

func NewRedirectHandler(rs *url.RedirectService, ps *url.PasswordService) *RedirectHandler {
	return &RedirectHandler{redirectSvc: rs, passwordSvc: ps}
}

func (h *RedirectHandler) HandleRedirect(c *gin.Context) {
	slug := c.Param("slug")
	tenantID, _ := uuid.Parse(c.GetHeader("X-Tenant-ID"))

	reqCtx := url.RequestContext{
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		CountryCode: "US",
		DeviceType:  "mobile",
	}

	urlEntity, finalDest, err := h.redirectSvc.Resolve(c.Request.Context(), tenantID, slug, reqCtx)
	if err != nil {
		if err == url.ErrURLNotFound {
			c.Data(404, "text/html", []byte("404 Not Found"))
			return
		}
		if err == url.ErrRateLimited {
			c.Data(429, "text/html", []byte("429 Too Many Requests - Quota Exceeded"))
			return
		}
		if err == url.ErrPasswordRequired {
			cookieName := "unlock_" + slug
			cookie, err := c.Cookie(cookieName)
			if err != nil || cookie != "valid" {
				c.JSON(401, gin.H{"error": "password required"})
				return
			}
			finalDest = string(urlEntity.OriginalURL)
		}
	}

	statusCode := 302
	if urlEntity.RedirectType == "permanent" {
		statusCode = 301
	}
	c.Redirect(statusCode, finalDest)
}

// HandleUnlock handles POST /v1/links/:slug/unlock (4.16)
func (h *RedirectHandler) HandleUnlock(c *gin.Context) {
	slug := c.Param("slug")
	tenantID, _ := uuid.Parse(c.GetHeader("X-Tenant-ID"))

	var body struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	valid, err := h.passwordSvc.VerifyPassword(c.Request.Context(), tenantID, slug, body.Password)
	if err != nil || !valid {
		c.JSON(401, gin.H{"error": "invalid password"})
		return
	}

	c.SetCookie("unlock_"+slug, "valid", 3600, "/", "", false, true)
	c.JSON(200, gin.H{"status": "unlocked"})
}
