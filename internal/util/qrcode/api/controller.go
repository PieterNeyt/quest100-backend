package api

import (
	"Quest100Backend/internal/util/qrcode/application"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QRCodeHandler struct {
	qrService application.QRCodeService
}

func NewQRCodeHandler(qrService application.QRCodeService) *QRCodeHandler {
	return &QRCodeHandler{
		qrService: qrService,
	}
}

func (h *QRCodeHandler) GenerateQRCode(c *gin.Context) {
	profileId, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no profile id found"})
		return
	}

	qrCodeBase64, err := h.qrService.GenerateAttendanceQRCode(profileId.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GenerateQRResponse{
		QRCode: qrCodeBase64,
	})
}
