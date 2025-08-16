package tasks

import (
	"encoding/json/v2"
	"github.com/clockme/clockme-backend/internal/shared/common"
	db2 "github.com/clockme/clockme-backend/internal/shared/db"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type TaskHandler struct {
	db *db2.Queries
}

func NewTaskHandler(db *db2.Queries) *TaskHandler {
	return &TaskHandler{
		db: db,
	}
}
func (t *TaskHandler) RegisterRoutes(router *http.ServeMux, mw func(http.Handler) http.Handler) {
	router.Handle("POST /api/v1/projects/{projectID}/tasks", mw(http.HandlerFunc(t.CreateTask)))
	router.Handle("GET /api/v1/tasks", mw(http.HandlerFunc(t.GetTasks)))
	router.Handle("GET /api/v1/tasks/{taskID}", mw(http.HandlerFunc(t.GetTask)))
	router.Handle("PUT /api/v1/tasks/{taskID}", mw(http.HandlerFunc(t.UpdateTask)))
	router.Handle("DELETE /api/v1/tasks/{taskID}", mw(http.HandlerFunc(t.DeleteTask)))

	router.Handle("GET /api/v1/projects/{projectID}/tasks", mw(http.HandlerFunc(t.GetTasksForProject)))
	router.Handle("GET /api/v1/users/{userID}/tasks", mw(http.HandlerFunc(t.GetTasksForUser)))
}
func (t *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectIDString := r.PathValue("projectID")
	projectID, err := uuid.Parse(projectIDString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse project ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var request CreateTaskRequest
	if err = json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createParams := db2.CreateTaskParams{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      request.Name,
	}
	task, err := t.db.CreateTask(ctx, createParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create task in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response := TaskResponse{
		ID:        task.ID,
		ProjectID: task.ProjectID,
		Name:      task.Name,
	}
	common.RespondWithJSON(w, http.StatusCreated, response)
}
func (t *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var tasksResponse []TaskResponse
	tasks, err := t.db.GetAllTasks(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tasks from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	for _, task := range tasks {
		taskResponse := TaskResponse{
			ID:        task.ID,
			ProjectID: task.ProjectID,
			Name:      task.Name,
		}
		tasksResponse = append(tasksResponse, taskResponse)
	}
	common.RespondWithJSON(w, http.StatusOK, tasksResponse)
}
func (t *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("taskID")
	taskID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse task ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid task ID format")
		return
	}
	task, err := t.db.GetTask(ctx, taskID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get task from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	taskResponse := TaskResponse{
		ID:        task.ID,
		ProjectID: task.ProjectID,
		Name:      task.Name,
	}
	common.RespondWithJSON(w, http.StatusOK, taskResponse)
}
func (t *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("taskID")
	taskID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse task ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid task ID format")
		return
	}
	task, err := t.db.GetTask(ctx, taskID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get task from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
	}
	var request UpdateTaskRequest
	if err = json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updateParams := db2.UpdateTaskParams{
		ID:        task.ID,
		ProjectID: task.ProjectID,
		Name:      request.Name,
	}
	updatedTask, err := t.db.UpdateTask(ctx, updateParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update task in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response := TaskResponse{
		ID:        updatedTask.ID,
		ProjectID: updatedTask.ProjectID,
		Name:      updatedTask.Name,
	}
	common.RespondWithJSON(w, http.StatusOK, response)
}
func (t *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("taskID")
	taskID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse task ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid task ID format")
		return
	}
	err = t.db.DeleteTask(ctx, taskID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete task from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	common.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}
func (t *TaskHandler) GetTasksForProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("projectID")
	projectID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse project ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}
	var tasksResponse []TaskResponse
	tasks, err := t.db.ListTaskForProject(ctx, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tasks from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	for _, task := range tasks {
		taskResponse := TaskResponse{
			ID:        task.ID,
			ProjectID: task.ProjectID,
			Name:      task.Name,
		}
		tasksResponse = append(tasksResponse, taskResponse)
	}
	common.RespondWithJSON(w, http.StatusOK, tasksResponse)
}
func (t *TaskHandler) GetTasksForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idString := r.PathValue("userID")
	userID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	tasks, err := t.db.ListTaskForUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tasks for user from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var userTasksResponse []TaskResponse
	for _, task := range tasks {
		taskResponse := TaskResponse{
			ID:        task.ID,
			ProjectID: task.ProjectID,
			Name:      task.Name,
		}
		userTasksResponse = append(userTasksResponse, taskResponse)
	}

	common.RespondWithJSON(w, http.StatusOK, userTasksResponse)
}
