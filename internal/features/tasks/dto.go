package tasks

import "github.com/google/uuid"

type CreateTaskRequest struct {
	Name string `json:"name"`
}
type UpdateTaskRequest struct {
	Name string `json:"name"`
}
type TaskResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"projectID"`
	Name      string    `json:"name"`
}
