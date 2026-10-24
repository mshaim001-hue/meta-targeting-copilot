-- name: GetTargetingCache :one
SELECT * FROM targeting_cache
WHERE query = $1 AND country = $2 AND kind = $3 AND api_version = $4
LIMIT 1;

-- name: UpsertTargetingCache :one
INSERT INTO targeting_cache (query, country, kind, result, api_version)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (query, country, kind, api_version)
DO UPDATE SET result = EXCLUDED.result, created_at = now()
RETURNING *;

-- name: CreateAuditEvent :one
INSERT INTO audit_events (project_id, entity_type, entity_id, value, source, meta_id, generated_from, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListAuditByProject :many
SELECT * FROM audit_events
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT 200;
