package api

import (
	"Quest100Backend/internal/gotcha/domain"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func profileID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("profileID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	return v.(uuid.UUID), true
}

func campus(c *gin.Context) string {
	v, _ := c.Get("campus")
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func (h *GotchaHandler) getCampus(c *gin.Context, pid uuid.UUID) (string, error) {
	if cam := campus(c); cam != "" {
		return cam, nil
	}
	profile, err := h.profileService.GetProfileById(pid)
	if err != nil {
		return "", err
	}
	if profile.Campus == "" {
		return "", fmt.Errorf("campus not set on profile, sync first")
	}
	return profile.Campus, nil
}

func propToResponse(p *domain.GotchaProp) PropResponse {
	return PropResponse{ID: p.ID, NameEN: p.NameEN, NameNL: p.NameNL}
}
