-- name: CreateTask :one
INSERT INTO tasks(id, name, project_id)
VALUES (
        $1,
        $2,
        $3
       )
RETURNING *;

-- name: GetAllTasks :many
SELECT * FROM tasks;

-- name: GetTask :one
SELECT * FROM tasks
WHERE tasks.id = $1;

-- name: UpdateTask :one
UPDATE tasks
SET
    updated_at = now(),
    name = $2,
    project_id = $3
WHERE tasks.id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE tasks.id = $1;

-- name: DeleteAllTasks :exec
DELETE FROM tasks;

-- name: ListTaskForProject :many
SELECT * FROM tasks
WHERE project_id = $1
ORDER BY created_at;

-- name: ListTaskForUser :many
SELECT t.* FROM tasks AS t
JOIN projects_users AS pu ON pu.project_id = t.project_id
WHERE pu.user_id = $1
ORDER BY created_at;