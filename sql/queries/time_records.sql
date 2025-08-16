-- name: CreateTimeRecord :one
INSERT INTO time_records (
    id,
    name,
    start_time,
    end_time,
    task_id
) VALUES (
             $1, $2, $3, $4, $5
         )
RETURNING *;

-- name: GetAllTimeRecords :many
SELECT * FROM time_records;

-- name: GetTimeRecord :one
SELECT * FROM time_records
WHERE id = $1;

-- name: UpdateTimeRecord :one
UPDATE time_records
SET
    name = $2,
    start_time = $3,
    end_time = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteTimeRecord :exec
DELETE FROM time_records
WHERE id = $1;

-- name: ListTimeRecordsForTask :many
SELECT * FROM time_records
WHERE task_id = $1
ORDER BY created_at;

-- name: ListTimeRecordsForProject :many
SELECT tr.* FROM time_records AS tr
                     JOIN tasks AS t ON tr.task_id = t.id
WHERE t.project_id = $1
ORDER BY tr.created_at;

-- name: ListTimeRecordsForUser :many
SELECT tr.*
FROM time_records AS tr
         JOIN tasks AS t ON tr.task_id = t.id
         JOIN projects_users AS pu ON t.project_id = pu.project_id
WHERE pu.user_id = $1
ORDER BY tr.created_at;