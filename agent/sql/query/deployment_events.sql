-- name: CreateDeploymentEvent :one
INSERT INTO deployment_events (
    user_id, deployment_id, event_type, message
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListUserEvents :many
SELECT * FROM deployment_events
WHERE user_id = $1
ORDER BY created_at;

-- name: ListDeploymentEvents :many
SELECT * FROM deployment_events
WHERE user_id = $1 and deployment_id = $2
ORDER BY created_at;
