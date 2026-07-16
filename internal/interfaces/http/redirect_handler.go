package http

import (
	"net/http"

	"github.com/PeshawaAziz/url-shortener/internal/domain/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RedirectHandler struct {
	redirectSvc *url.RedirectService
}

func NewRedirectHandler(svc *url.RedirectService) *RedirectHandler {
	return &RedirectHandler{redirectSvc: svc}
}

// HandleRedirect handles GET /:slug
func (h *RedirectHandler) HandleRedirect(c *gin.Context) {
	slug := c.Param("slug")

	tenantIDStr := c.GetHeader("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		tenantID = uuid.New()
	}

	meta := url.ClickMetadata{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
	}

	urlEntity, err := h.redirectSvc.Resolve(c.Request.Context(), tenantID, slug, meta)
	if err != nil {
		if err == url.ErrURLNotFound {
			c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte("<h1>404 - Link Not Found</h1><p>The link you clicked does not exist.</p>"))
			return
		}
		if err == url.ErrLinkExpired {
			c.Data(http.StatusGone, "text/html; charset=utf-8", []byte("<h1>410 - Link Expired</h1>"))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	statusCode := http.StatusFound // 302 (Temporary)
	if urlEntity.RedirectType == "permanent" {
		statusCode = http.StatusMovedPermanently // 301 (Permanent)
	}

	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Robots-Tag", "noindex, nofollow")

	c.Redirect(statusCode, string(urlEntity.OriginalURL))
}
