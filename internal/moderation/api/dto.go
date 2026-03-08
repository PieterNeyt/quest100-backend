package api

import "github.com/google/uuid"

type CreateReportRequest struct {
	TargetID    uuid.UUID `json:"targetId" binding:"required"`
	ChannelType int       `json:"channelType" binding:"min=0,max=2"`
	ReportType  int       `json:"reportType" binding:"min=0,max=4"`
	Message     string    `json:"message" binding:"required"`
}
