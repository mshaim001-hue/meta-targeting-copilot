-- name: CreateOffer :one
INSERT INTO offers (
    segment_id, awareness_level, problem, motivation, objection, offer, hook, cta
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListOffersByBrief :many
SELECT o.* FROM offers o
JOIN audience_segments s ON s.id = o.segment_id
WHERE s.brief_id = $1
ORDER BY s.position ASC;

-- name: CreateCreative :one
INSERT INTO creatives (
    segment_id, angle, hook, visual_idea, copy_angle, cta, format
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ListCreativesByBrief :many
SELECT c.* FROM creatives c
JOIN audience_segments s ON s.id = c.segment_id
WHERE s.brief_id = $1
ORDER BY s.position ASC;

-- name: CreateExperiment :one
INSERT INTO experiments (project_id, brief_id, hypothesis, audience, variable, variants)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListExperimentsByBrief :many
SELECT * FROM experiments WHERE brief_id = $1 ORDER BY created_at ASC;
