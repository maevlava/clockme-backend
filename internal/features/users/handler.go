package users

import (
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"github.com/clockme/clockme-backend/internal/features/auth"
	"github.com/clockme/clockme-backend/internal/shared/common"
	db "github.com/clockme/clockme-backend/internal/shared/db"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type UserStore interface {
	GetUser(ctx context.Context, userID uuid.UUID) (db.User, error)
	CreateUser(ctx context.Context, createParams db.CreateUserParams) (db.User, error)
	UpdateUser(ctx context.Context, updateParams db.UpdateUserParams) (db.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetAllUsers(ctx context.Context) ([]db.User, error)
}
type UserHandler struct {
	store UserStore
}

func NewUserHandler(store UserStore) *UserHandler {
	return &UserHandler{
		store: store,
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
	createParams := db.CreateUserParams{
		ID:             uuid.New(),
		Name:           request.Name,
		Email:          request.Email,
		HashedPassword: hashedPassword,
	}
	newUser, err := u.store.CreateUser(ctx, createParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user in database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := UserResponse{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}
	common.RespondWithJSON(w, http.StatusCreated, response)
}
func (u *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var usersResponse []UserResponse

	users, err := u.store.GetAllUsers(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg("No users found in database")
			common.RespondWithError(w, http.StatusNotFound, "No users found")
			return
		}
		log.Error().Err(err).Msg("Failed to get users from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	for _, user := range users {
		userResponse := UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
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

	user, err := u.store.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Str("userID", userID.String()).Msg("User not found in database")
			common.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Error().Err(err).Str("userID", userID.String()).Msg("Failed to get user from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
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
	updateParams := db.UpdateUserParams{
		ID:    userID,
		Name:  request.Name,
		Email: request.Email,
	}
	updatedUser, err := u.store.UpdateUser(ctx, updateParams)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Str("userID", userID.String()).Msg("User not found in database")
			common.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Error().Err(err).Str("userID", userID.String()).Msg("Failed to update from database")
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
		common.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	err = u.store.DeleteUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			common.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		log.Error().Err(err).Msg("Failed to delete user from database")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
