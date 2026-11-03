-- name: UpdateHomepage :exec
INSERT INTO homepage (id, title, subtitle,home_image,home_image_alt)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (id)
DO UPDATE SET
    title = EXCLUDED.title,
    subtitle = EXCLUDED.subtitle,
    home_image = EXCLUDED.home_image,
    home_image_alt = EXCLUDED.home_image_alt;

-- name: GetHomepage :one
SELECT *
FROM homepage
WHERE id = $1
LIMIT 1;