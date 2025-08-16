package timerecords

import (
	"github.com/clockme/clockme-backend/internal/shared/common"
	"github.com/google/uuid"
)

type CreateTimeRecordRequest struct {
	Name      string            `json:"name"`
	StartTime common.CustomTime `json:"startTime"`
	EndTime   common.CustomTime `json:"endTime"`
}
type UpdateTimeRecordRequest struct {
	Name      string            `json:"name"`
	StartTime common.CustomTime `json:"startTime"`
	EndTime   common.CustomTime `json:"endTime"`
}
type TimeRecordResponse struct {
	ID        uuid.UUID         `json:"id"`
	TaskID    uuid.UUID         `json:"taskID"`
	Name      string            `json:"name"`
	StartTime common.CustomTime `json:"startTime"`
	EndTime   common.CustomTime `json:"endTime"`
}
