package projects

import "github.com/google/uuid"

type CreateProjectRequest struct {
	Name string `json:"name"`
}
type UpdateProjectRequest struct {
	Name string `json:"name"`
}
type AddUserToProjectRequest struct {
	UserID uuid.UUID `json:"userID"`
}
type ProjectResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
