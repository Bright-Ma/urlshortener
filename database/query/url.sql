-- name: CreateURL :exec
INSERT INTO urls (
    original_url,
    short_code,
    is_custom,
    expired_at,
    user_id
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: IsShortCodeAvailable :one
SELECT NOT EXISTS(
    SELECT 1 FROM urls
    WHERE short_code = $1
) AS is_available;

-- name: GetUrlByShortCode :one
SELECT original_url, short_code, views, is_custom FROM urls 
WHERE short_code = $1
AND expired_at > CURRENT_TIMESTAMP;


-- name: UpdateViewsByShortCode :exec
UPDATE urls
SET views = views + $1
WHERE short_code = $2;

-- name: GetURLsByUserID :many
SELECT original_url, short_code, views, is_custom, expired_at FROM urls r
WHERE r.user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
