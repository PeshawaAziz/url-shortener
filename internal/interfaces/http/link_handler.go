package http

import (
	"net/http"

	"github.com/PeshawaAziz/url-shortener/internal/domain/url"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LinkHandler struct {
	shortenerSvc *url.ShortenerService
}

func NewLinkHandler(svc *url.ShortenerService) *LinkHandler {
	return &LinkHandler{shortenerSvc: svc}
}

func (h *LinkHandler) HandleCreate(c *gin.Context) {
	var req struct {
		OriginalURL string `json:"original_url" binding:"required"`
		DesiredSlug string `json:"desired_slug"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := uuid.Parse(c.GetHeader("X-Tenant-ID"))
	if err != nil {
		tenantID = uuid.New()
	}
	userID, err := uuid.Parse(c.GetHeader("X-User-ID"))
	if err != nil {
		userID = uuid.New()
	}
	idemKey := c.GetHeader("X-Idempotency-Key")

	input := url.ShortenURLInput{
		TenantID:       tenantID,
		UserID:         userID,
		OriginalURL:    req.OriginalURL,
		DesiredSlug:    req.DesiredSlug,
		IdempotencyKey: idemKey,
	}

	result, err := h.shortenerSvc.ShortenURL(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           result.ID,
		"slug":         result.Slug,
		"original_url": result.OriginalURL,
	})
}
