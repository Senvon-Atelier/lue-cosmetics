-- name: CreateOrder :one
INSERT INTO orders (user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
                    total_ghs_minor, paystack_reference, shipping_address)
VALUES ($1, 'pending', $2, $3, $4, $5, $6)
RETURNING id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
          total_ghs_minor, paystack_reference, paystack_transaction_id,
          shipping_address, created_at, updated_at;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, qty, unit_price_ghs_minor,
                         product_name_snapshot, product_brand_snapshot,
                         product_image_snapshot)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, order_id, product_id, qty, unit_price_ghs_minor,
          product_name_snapshot, product_brand_snapshot,
          product_image_snapshot, created_at;

-- name: GetOrderByReference :one
SELECT id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
       total_ghs_minor, paystack_reference, paystack_transaction_id,
       shipping_address, created_at, updated_at
FROM orders WHERE paystack_reference = $1;

-- name: GetOrderByID :one
SELECT id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
       total_ghs_minor, paystack_reference, paystack_transaction_id,
       shipping_address, created_at, updated_at
FROM orders WHERE id = $1;

-- name: ListOrderItems :many
SELECT id, order_id, product_id, qty, unit_price_ghs_minor,
       product_name_snapshot, product_brand_snapshot,
       product_image_snapshot, created_at
FROM order_items
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: GetOrderByReferenceForUpdate :one
SELECT id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
       total_ghs_minor, paystack_reference, paystack_transaction_id,
       shipping_address, created_at, updated_at
FROM orders WHERE paystack_reference = $1
FOR UPDATE;

-- name: MarkOrderPaid :exec
UPDATE orders
SET status = 'paid',
    paystack_transaction_id = $2,
    updated_at = now()
WHERE id = $1 AND status = 'pending';

-- name: CountOrdersByStatus :one
SELECT count(*) FROM orders WHERE status = $1;
