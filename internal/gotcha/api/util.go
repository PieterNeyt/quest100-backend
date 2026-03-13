package api

import (
	"Quest100Backend/internal/gotcha/domain"
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

func propToResponse(p *domain.GotchaProp) PropResponse {
	return PropResponse{ID: p.ID, NameEN: p.NameEN, NameNL: p.NameNL}
}

func (h *GotchaHandler) profile(c *gin.Context) (uuid.UUID, bool) {
	pid, ok := profileID(c)
	if !ok {
		return uuid.Nil, false
	}
	return pid, true
}
func parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}
