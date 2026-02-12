package domain

import "fmt"

type InvalidQRCodeDataError struct {
	Message string
}

func (e *InvalidQRCodeDataError) Error() string {
	return fmt.Sprintf("invalid qr code data: %s", e.Message)
}

type QRCodeGenerationError struct {
	Message string
}

func (e *QRCodeGenerationError) Error() string {
	return fmt.Sprintf("qr code generation failed: %s", e.Message)
}
