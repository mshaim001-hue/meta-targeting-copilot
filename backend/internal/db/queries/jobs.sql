-- name: CreateJob :one
INSERT INTO jobs (project_id, brief_id, type, status, stage, payload)
VALUES ($1, $2, $3, 'queued', $4, $5)
RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = $1;

-- name: ListJobsByProject :many
SELECT * FROM jobs WHERE project_id = $1 ORDER BY created_at DESC;

-- name: GetLatestJobByProject :one
SELECT * FROM jobs WHERE project_id = $1 ORDER BY created_at DESC LIMIT 1;

-- name: ClaimJobs :many
UPDATE jobs
SET status = 'running', locked_at = now(), attempts = attempts + 1, updated_at = now()
WHERE id IN (
    SELECT id FROM jobs
    WHERE status = 'queued' AND run_after <= now()
    ORDER BY created_at
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: UpdateJobStage :one
UPDATE jobs
SET stage = $2, status = $3, error = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CompleteJob :one
UPDATE jobs
SET status = 'complete', stage = 'COMPLETE', result = $2, error = NULL, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: FailJob :one
UPDATE jobs
SET status = 'failed', error = $2, updated_at = now()
WHERE id = $1
RETURNING *;
