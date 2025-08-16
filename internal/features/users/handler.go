package users

import (
	"context"
	"encoding/json/v2"
	"github.com/clockme/clockme-backend/internal/features/auth"
	"github.com/clockme/clockme-backend/internal/shared/common"
	db2 "github.com/clockme/clockme-backend/internal/shared/db"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type UserHandler struct {
	db *db2.Queries
}

func NewUserHandler(db *db2.Queries) *UserHandler {
	return &UserHandler{
		db: db,
	}
}
func (u *UserHandler) RegisterRoutes(router *http.ServeMux, mw func(http.Handler) http.Handler) {
	router.Handle("POST /api/v1/users", mw(http.HandlerFunc(u.CreateUser)))
	router.Handle("GET /api/v1/users", mw(http.HandlerFunc(u.GetUsers)))
	router.Handle("GET /api/v1/users/{userID}", mw(http.HandlerFunc(u.GetUser)))
	router.Handle("PUT /api/v1/users/{userID}", mw(http.HandlerFunc(u.UpdateUser)))
	router.Handle("DELETE /api/v1/users/{userID}", mw(http.HandlerFunc(u.DeleteUser)))
}

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var request CreateUserRequest

	if err := json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password during user creation")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	createParams := db2.CreateUserParams{
		ID:             uuid.New(),
		Name:           request.Name,
		Email:          request.Email,
		HashedPassword: hashedPassword,
	}
	newUser, err := u.db.CreateUser(ctx, createParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := UserResponse{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt.Time,
		UpdatedAt: newUser.UpdatedAt.Time,
	}
	common.RespondWithJSON(w, http.StatusCreated, response)
}
func (u *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var usersResponse []UserResponse

	users, err := u.db.GetAllUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users from database")
	}
	for _, user := range users {
		userResponse := UserResponse{
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
func (u *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	idString := r.PathValue("userID")

	userID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	user, err := u.db.GetUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}

	common.RespondWithJSON(w, http.StatusOK, response)
}
func (u *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("userID")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var request UpdateUserRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	userID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	updateParams := db2.UpdateUserParams{
		ID:    userID,
		Name:  request.Name,
		Email: request.Email,
	}
	updatedUser, err := u.db.UpdateUser(ctx, updateParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response := UserResponse{
		ID:    updatedUser.ID,
		Name:  updatedUser.Name,
		Email: updatedUser.Email,
	}
	common.RespondWithJSON(w, http.StatusOK, response)
}
func (u *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("userID")
	userID, err := uuid.Parse(idString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	err = u.db.DeleteUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	common.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}
