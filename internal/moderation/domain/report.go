package domain

import "github.com/google/uuid"

type Report struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey;" json:"id"`
	UserID      uuid.UUID   `gorm:"type:uuid;" json:"userId"`
	TargetID    uuid.UUID   `gorm:"type:uuid;" json:"targetId"`
	ContextID   *uuid.UUID  `gorm:"type:uuid;" json:"contextId,omitempty"`
	Resolved    bool        `gorm:"default:false" json:"resolved"`
	ChannelType ChannelType `json:"channelType"`
	ReportType  ReportType  `json:"reportType"`
	Message     string      `json:"message"`
}

type ReportType int

const (
	ReportTypeHarassment ReportType = iota
	ReportTypeRacism
	ReportTypeSpam
	ReportTypeHateSpeech
	ReportTypeOther
)

type ChannelType int

const (
	ChannelTypeEvent ChannelType = iota
	ChannelTypeMessage
	ChannelTypeAward
)

type ModerationRepository interface {
	SaveReport(report *Report) error
	GetReports() (*[]Report, error)
	GetReportById(id uuid.UUID) (*Report, error)
	ResolveReport(id uuid.UUID) error
	HasOpenReport(targetID uuid.UUID, channelType ChannelType) (bool, error)
	HasOpenReportByContextID(contextID uuid.UUID, channelType ChannelType) (bool, error)
}

func NewReport(userID, targetID uuid.UUID, channelType ChannelType, reportType ReportType, message string, contextID *uuid.UUID) (*Report, error) {
	if message == "" {
		return nil, &InvalidReportError{Message: "message cannot be empty"}
	}

	return &Report{
		ID:          uuid.New(),
		UserID:      userID,
		TargetID:    targetID,
		ContextID:   contextID,
		Resolved:    false,
		ChannelType: channelType,
		ReportType:  reportType,
		Message:     message,
	}, nil
}
