-- name: CreateSegment :one
INSERT INTO audience_segments (
    brief_id, name, description, intent_level, pains, motivations,
    semantic_signals, offer_angles, creative_angles, position
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: ListSegmentsByBrief :many
SELECT * FROM audience_segments
WHERE brief_id = $1
ORDER BY position ASC, created_at ASC;

-- name: GetSegment :one
SELECT * FROM audience_segments WHERE id = $1;
