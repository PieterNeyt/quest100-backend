package api

type GenerateQRRequest struct {
	ID string `json:"id" binding:"required"`
}

type GenerateQRResponse struct {
	QRCode string `json:"qrCode"`
	ID     string `json:"id"`
}
