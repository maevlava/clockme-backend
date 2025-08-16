package projects

import (
	"encoding/json/v2"
	"github.com/clockme/clockme-backend/internal/features/users"
	"github.com/clockme/clockme-backend/internal/shared/common"
	db2 "github.com/clockme/clockme-backend/internal/shared/db"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type ProjectHandler struct {
	db *db2.Queries
}

func NewProjectHandler(db *db2.Queries) *ProjectHandler {
	return &ProjectHandler{
		db: db,
	}
}
func (p *ProjectHandler) RegisterRoutes(router *http.ServeMux, mw func(http.Handler) http.Handler) {
	router.Handle("POST /api/v1/projects", mw(http.HandlerFunc(p.CreateProject)))
	router.Handle("GET /api/v1/projects", mw(http.HandlerFunc(p.GetProjects)))
	router.Handle("GET /api/v1/projects/{projectID}", mw(http.HandlerFunc(p.GetProject)))
	router.Handle("PUT /api/v1/projects/{projectID}", mw(http.HandlerFunc(p.UpdateProject)))
	router.Handle("DELETE /api/v1/projects/{projectID}", mw(http.HandlerFunc(p.DeleteProject)))

	// Users - Projects
	router.Handle("POST /api/v1/projects/{projectID}/users", mw(http.HandlerFunc(p.AddUserToProject)))
	router.Handle("GET /api/v1/projects/{projectID}/users", mw(http.HandlerFunc(p.GetUsersInProject)))
	router.Handle("DELETE /api/v1/projects/{projectID}/users/{userID}", mw(http.HandlerFunc(p.DeleteUserFromProject)))
}

func (p *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var request CreateProjectRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	createParams := db2.CreateProjectParams{
		ID:   uuid.New(),
		Name: request.Name,
	}
	newProject, err := p.db.CreateProject(ctx, createParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create project in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response := ProjectResponse{
		ID:   newProject.ID,
		Name: newProject.Name,
	}
	common.RespondWithJSON(w, http.StatusCreated, response)
}
func (p *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	var projectsResponse []ProjectResponse
	ctx := r.Context()
	projects, err := p.db.GetAllProjects(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get projects from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	for _, project := range projects {
		projectResponse := ProjectResponse{
			ID:   project.ID,
			Name: project.Name,
		}
		projectsResponse = append(projectsResponse, projectResponse)
	}
	common.RespondWithJSON(w, http.StatusOK, projectsResponse)
}
func (p *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("projectID")
	projectID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse project ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}
	project, err := p.db.GetProject(ctx, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get project from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	projectResponse := ProjectResponse{
		ID:   project.ID,
		Name: project.Name,
	}
	common.RespondWithJSON(w, http.StatusOK, projectResponse)
}
func (p *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("projectID")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var request UpdateProjectRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	projectID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse project ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}
	updateParams := db2.UpdateProjectParams{
		ID:   projectID,
		Name: request.Name,
	}
	updatedProject, err := p.db.UpdateProject(ctx, updateParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update project in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	projectResponse := ProjectResponse{
		ID:   updatedProject.ID,
		Name: updatedProject.Name,
	}
	common.RespondWithJSON(w, http.StatusOK, projectResponse)
}
func (p *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("projectID")
	projectID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse project ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}
	err = p.db.DeleteProject(ctx, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete project from database")
	}
	common.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}

func (p *ProjectHandler) AddUserToProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("projectID")
	projectID, err := uuid.Parse(idString)
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
	var request AddUserToProjectRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	addUserToProjectParams := db2.AddUserToProjectParams{
		UserID:    request.UserID,
		ProjectID: projectID,
	}
	_, err = p.db.AddUserToProject(ctx, addUserToProjectParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add user to project in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	common.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User added to project successfully"})

}
func (p *ProjectHandler) GetUsersInProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("projectID")
	projectID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse project ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}
	usersInProject, err := p.db.ListUsersForProject(ctx, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get usersInProject in project from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	var usersResponse []users.UserResponse

	for _, user := range usersInProject {
		userResponse := users.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
		}
		usersResponse = append(usersResponse, userResponse)
	}
	common.RespondWithJSON(w, http.StatusOK, usersResponse)

}
func (p *ProjectHandler) DeleteUserFromProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("projectID")
	projectID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse project ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}
	idString = r.PathValue("userID")
	userID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
	}
	deleteUserFromProjectParams := db2.RemoveUserFromProjectParams{
		UserID:    userID,
		ProjectID: projectID,
	}
	err = p.db.RemoveUserFromProject(ctx, deleteUserFromProjectParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user from project in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	common.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted from project successfully"})
}
