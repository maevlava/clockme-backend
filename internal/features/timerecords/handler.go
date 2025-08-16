package timerecords

import (
	"encoding/json/v2"
	"github.com/clockme/clockme-backend/internal/shared/common"
	"github.com/clockme/clockme-backend/internal/shared/db"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type TimeRecordHandler struct {
	db *db.Queries
}

func NewTimeRecordHandler(db *db.Queries) *TimeRecordHandler {
	return &TimeRecordHandler{db: db}
}
func (t *TimeRecordHandler) RegisterRoutes(router *http.ServeMux, mw func(http.Handler) http.Handler) {
	router.Handle("POST /api/v1/tasks/{taskID}/timerecords", mw(http.HandlerFunc(t.CreateTimeRecord)))
	router.Handle("GET /api/v1/timerecords", mw(http.HandlerFunc(t.GetTimeRecords)))
	router.Handle("GET /api/v1/timerecords/{timeRecordID}", mw(http.HandlerFunc(t.GetTimeRecord)))
	router.Handle("PUT /api/v1/timerecords/{timeRecordID}", mw(http.HandlerFunc(t.UpdateTimeRecord)))
	router.Handle("DELETE /api/v1/timerecords/{timeRecordID}", mw(http.HandlerFunc(t.DeleteTimeRecord)))

	router.Handle("GET /api/v1/tasks/{taskID}/timerecords", mw(http.HandlerFunc(t.GetTimeRecordsForTask)))
	router.Handle("GET /api/v1/projects/{projectID}/timerecords", mw(http.HandlerFunc(t.GetTimeRecordsForProject)))
	router.Handle("GET /api/v1/users/{userID}/timerecords", mw(http.HandlerFunc(t.GetTimeRecordsForUser)))
}
func (t *TimeRecordHandler) CreateTimeRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("taskID")
	taskID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse task ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid task ID format")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var request CreateTimeRecordRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.EndTime.Time.Before(request.StartTime.Time) {
		common.RespondWithError(w, http.StatusBadRequest, "endTime must be after startTime")
		return
	}

	createParams := db.CreateTimeRecordParams{
		ID:        uuid.New(),
		TaskID:    taskID,
		Name:      request.Name,
		StartTime: request.StartTime.Time,
		EndTime:   request.EndTime.Time,
	}
	newTimeRecord, err := t.db.CreateTimeRecord(ctx, createParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create time record in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response := TimeRecordResponse{
		ID:        newTimeRecord.ID,
		TaskID:    newTimeRecord.TaskID,
		Name:      newTimeRecord.Name,
		StartTime: common.CustomTime{Time: newTimeRecord.StartTime},
		EndTime:   common.CustomTime{Time: newTimeRecord.EndTime},
	}
	common.RespondWithJSON(w, http.StatusCreated, response)
}
func (t *TimeRecordHandler) GetTimeRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	timeRecords, err := t.db.GetAllTimeRecords(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get time records from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var timeRecordsResponse []TimeRecordResponse
	for _, timeRecord := range timeRecords {
		timeRecordResponse := TimeRecordResponse{
			ID:        timeRecord.ID,
			TaskID:    timeRecord.TaskID,
			Name:      timeRecord.Name,
			StartTime: common.CustomTime{Time: timeRecord.StartTime},
			EndTime:   common.CustomTime{Time: timeRecord.EndTime},
		}
		timeRecordsResponse = append(timeRecordsResponse, timeRecordResponse)
	}
	common.RespondWithJSON(w, http.StatusOK, timeRecordsResponse)
}
func (t *TimeRecordHandler) GetTimeRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("timeRecordID")
	timeRecordID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse time record ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid time record ID format")
		return
	}
	timeRecord, err := t.db.GetTimeRecord(ctx, timeRecordID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get time record from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	timeRecordResponse := TimeRecordResponse{
		ID:        timeRecord.ID,
		TaskID:    timeRecord.TaskID,
		Name:      timeRecord.Name,
		StartTime: common.CustomTime{Time: timeRecord.StartTime},
		EndTime:   common.CustomTime{Time: timeRecord.EndTime},
	}
	common.RespondWithJSON(w, http.StatusOK, timeRecordResponse)
}
func (t *TimeRecordHandler) UpdateTimeRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("timeRecordID")
	timeRecordID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse time record ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid time record ID format")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var request UpdateTimeRecordRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updateParams := db.UpdateTimeRecordParams{
		ID:        timeRecordID,
		Name:      request.Name,
		StartTime: request.StartTime.Time,
		EndTime:   request.EndTime.Time,
	}
	updatedTimeRecord, err := t.db.UpdateTimeRecord(ctx, updateParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update time record in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := TimeRecordResponse{
		ID:        updatedTimeRecord.ID,
		TaskID:    updatedTimeRecord.TaskID,
		Name:      updatedTimeRecord.Name,
		StartTime: common.CustomTime{Time: updatedTimeRecord.StartTime},
		EndTime:   common.CustomTime{Time: updatedTimeRecord.EndTime},
	}

	common.RespondWithJSON(w, http.StatusOK, response)
}
func (t *TimeRecordHandler) DeleteTimeRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("timeRecordID")
	timeRecordID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse time record ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid time record ID format")
		return
	}
	err = t.db.DeleteTimeRecord(ctx, timeRecordID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete time record from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	common.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Time record deleted successfully"})
}

func (t *TimeRecordHandler) GetTimeRecordsForTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("taskID")
	taskID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse task ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid task ID format")
		return
	}
	timeRecords, err := t.db.ListTimeRecordsForTask(ctx, taskID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get time records from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	var timeRecordsResponse []TimeRecordResponse
	for _, timeRecord := range timeRecords {
		timeRecordResponse := TimeRecordResponse{
			ID:        timeRecord.ID,
			TaskID:    timeRecord.TaskID,
			Name:      timeRecord.Name,
			StartTime: common.CustomTime{Time: timeRecord.StartTime},
			EndTime:   common.CustomTime{Time: timeRecord.EndTime},
		}
		timeRecordsResponse = append(timeRecordsResponse, timeRecordResponse)
	}
	common.RespondWithJSON(w, http.StatusOK, timeRecordsResponse)
}
func (t *TimeRecordHandler) GetTimeRecordsForProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("projectID")
	projectID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse project ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}
	timeRecords, err := t.db.ListTimeRecordsForProject(ctx, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get time records from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
	var timeRecordsResponse []TimeRecordResponse
	for _, timeRecord := range timeRecords {
		timeRecordResponse := TimeRecordResponse{
			ID:        timeRecord.ID,
			TaskID:    timeRecord.TaskID,
			Name:      timeRecord.Name,
			StartTime: common.CustomTime{Time: timeRecord.StartTime},
			EndTime:   common.CustomTime{Time: timeRecord.EndTime},
		}
		timeRecordsResponse = append(timeRecordsResponse, timeRecordResponse)
	}
	common.RespondWithJSON(w, http.StatusOK, timeRecordsResponse)
}
func (t *TimeRecordHandler) GetTimeRecordsForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("userID")
	userID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	timeRecords, err := t.db.ListTimeRecordsForUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get time records from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	var timeRecordsResponse []TimeRecordResponse
	for _, timeRecord := range timeRecords {
		timeRecordResponse := TimeRecordResponse{
			ID:        timeRecord.ID,
			TaskID:    timeRecord.TaskID,
			Name:      timeRecord.Name,
			StartTime: common.CustomTime{Time: timeRecord.StartTime},
			EndTime:   common.CustomTime{Time: timeRecord.EndTime},
		}
		timeRecordsResponse = append(timeRecordsResponse, timeRecordResponse)
	}
	common.RespondWithJSON(w, http.StatusOK, timeRecordsResponse)
}
