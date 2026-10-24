-- name: CreateBrief :one
INSERT INTO audience_briefs (project_id, prompt)
VALUES ($1, $2)
RETURNING *;

-- name: GetBrief :one
SELECT * FROM audience_briefs WHERE id = $1;

-- name: GetLatestBriefByProject :one
SELECT * FROM audience_briefs
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateBriefAnalysis :one
UPDATE audience_briefs
SET analyzed_json = $2, missing_info = $3
WHERE id = $1
RETURNING *;
