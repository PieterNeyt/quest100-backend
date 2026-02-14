package api

import (
	"Quest100Backend/internal/util/qrcode/application"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QRCodeHandler struct {
	qrService application.QRCodeService
}

func NewQRCodeHandler(qrService application.QRCodeService) *QRCodeHandler {
	return &QRCodeHandler{
		qrService: qrService,
	}
}

type GenerateQRRequest struct {
	ID string `json:"id" binding:"required"`
}

type GenerateQRResponse struct {
	QRCode string `json:"qrCode"`
	ID     string `json:"id"`
}

func (h *QRCodeHandler) GenerateQRCode(c *gin.Context) {
	var req GenerateQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	qrCodeBase64, err := h.qrService.GenerateAttendanceQRCode(req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GenerateQRResponse{
		QRCode: qrCodeBase64,
		ID:     req.ID,
	})
}
