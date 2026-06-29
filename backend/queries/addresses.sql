-- name: CreateAddress :one
INSERT INTO addresses (user_id, label, line1, line2, city, region, phone, is_default)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at;

-- name: GetAddressByID :one
SELECT id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at
FROM addresses
WHERE id = $1;

-- name: ListAddressesByUserID :many
SELECT id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at
FROM addresses
WHERE user_id = $1
ORDER BY is_default DESC, created_at DESC;

-- name: UpdateAddress :one
UPDATE addresses
SET label = $2, line1 = $3, line2 = $4, city = $5, region = $6, phone = $7, updated_at = now()
WHERE id = $1
RETURNING id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at;

-- name: DeleteAddress :exec
DELETE FROM addresses WHERE id = $1;

-- name: CountAddressesByUserID :one
SELECT count(*) FROM addresses WHERE user_id = $1;

-- name: ClearDefaultForUser :exec
UPDATE addresses SET is_default = false, updated_at = now()
WHERE user_id = $1 AND is_default = true;

-- name: SetDefaultAddress :one
UPDATE addresses SET is_default = true, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at;

-- name: GetOldestOtherAddress :one
SELECT id
FROM addresses
WHERE user_id = $1 AND id != $2
ORDER BY created_at ASC
LIMIT 1;
