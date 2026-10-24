-- name: CreateTargetingEntity :one
INSERT INTO targeting_entities (
    segment_id, brief_id, query, meta_id, meta_name, meta_type,
    status, source, generated_from, explanation, audience_size, raw
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: ListTargetingByBrief :many
SELECT * FROM targeting_entities
WHERE brief_id = $1
ORDER BY created_at ASC;

-- name: ListTargetingBySegment :many
SELECT * FROM targeting_entities
WHERE segment_id = $1
ORDER BY created_at ASC;
