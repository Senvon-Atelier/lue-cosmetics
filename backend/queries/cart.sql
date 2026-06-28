-- name: GetCartByUserID :one
SELECT id, user_id, guest_token, created_at, updated_at FROM carts WHERE user_id = $1;

-- name: GetCartByGuestToken :one
SELECT id, user_id, guest_token, created_at, updated_at FROM carts WHERE guest_token = $1;

-- name: CreateCartForUser :one
INSERT INTO carts (user_id) VALUES ($1) RETURNING id, user_id, guest_token, created_at, updated_at;

-- name: CreateCartForGuest :one
INSERT INTO carts (guest_token) VALUES ($1) RETURNING id, user_id, guest_token, created_at, updated_at;

-- name: TouchCart :exec
UPDATE carts SET updated_at = now() WHERE id = $1;

-- name: DeleteCart :exec
DELETE FROM carts WHERE id = $1;

-- name: ListCartItems :many
SELECT id, cart_id, product_id, qty, unit_price_ghs_minor, created_at, updated_at
FROM cart_items
WHERE cart_id = $1
ORDER BY created_at ASC;

-- name: GetCartItemByID :one
SELECT id, cart_id, product_id, qty, unit_price_ghs_minor, created_at, updated_at
FROM cart_items
WHERE id = $1 AND cart_id = $2;

-- name: GetCartItemByProduct :one
SELECT id, cart_id, product_id, qty, unit_price_ghs_minor, created_at, updated_at
FROM cart_items
WHERE cart_id = $1 AND product_id = $2;

-- name: UpsertCartItemAddQty :one
INSERT INTO cart_items (cart_id, product_id, qty, unit_price_ghs_minor)
VALUES ($1, $2, $3, $4)
ON CONFLICT (cart_id, product_id) DO UPDATE
SET qty = cart_items.qty + EXCLUDED.qty, updated_at = now()
RETURNING id, cart_id, product_id, qty, unit_price_ghs_minor, created_at, updated_at;

-- name: SetCartItemQty :execrows
UPDATE cart_items SET qty = $3, updated_at = now()
WHERE id = $1 AND cart_id = $2;

-- name: DeleteCartItem :execrows
DELETE FROM cart_items WHERE id = $1 AND cart_id = $2;
