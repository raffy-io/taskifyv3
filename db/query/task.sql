-- name: ListTasks :many
SELECT *
FROM tasks
WHERE user_id = $1;
   

-- name: CreateTask :one
INSERT INTO tasks (title,user_id) 
VALUES ($1, $2)
RETURNING *;

-- name: ToggleTaskStatus :one
UPDATE tasks
SET is_done = NOT is_done
WHERE id = $1 AND user_id = $2
RETURNING *;


-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1 AND user_id = $2;

