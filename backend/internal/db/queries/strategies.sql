-- name: CreateStrategy :one
INSERT INTO audience_strategies (
    project_id, brief_id, name, type, geo, ages, gender,
    included_targeting, excluded_targeting, reach_estimate,
    confidence, confidence_level, basis, explanation, quality_score
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10,
    $11, $12, $13, $14, $15
)
RETURNING *;

-- name: ListStrategiesByBrief :many
SELECT * FROM audience_strategies
WHERE brief_id = $1
ORDER BY created_at ASC;

-- name: UpdateStrategyReach :one
UPDATE audience_strategies
SET reach_estimate = $2
WHERE id = $1
RETURNING *;
