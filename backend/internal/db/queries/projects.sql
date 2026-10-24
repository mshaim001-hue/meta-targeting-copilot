-- name: CreateProject :one
INSERT INTO projects (
    user_id, business_name, website, product, country, city, price, objective, budget,
    has_customer_list, has_pixel, has_capi, has_website_visitors, has_instagram,
    has_facebook_page, has_video_views, has_lead_forms, has_purchasers, has_crm
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19
)
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1;

-- name: ListProjectsByUser :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects SET
    business_name = $2,
    website = $3,
    product = $4,
    country = $5,
    city = $6,
    price = $7,
    objective = $8,
    budget = $9,
    has_customer_list = $10,
    has_pixel = $11,
    has_capi = $12,
    has_website_visitors = $13,
    has_instagram = $14,
    has_facebook_page = $15,
    has_video_views = $16,
    has_lead_forms = $17,
    has_purchasers = $18,
    has_crm = $19,
    updated_at = now()
WHERE id = $1
RETURNING *;
