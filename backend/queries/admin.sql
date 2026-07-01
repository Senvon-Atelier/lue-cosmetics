-- Admin queries for Rue Cosmetics dashboard and analytics
-- These queries provide aggregate data for admin views

-- name: GetDashboardStats :one
-- Returns summary statistics for the admin dashboard
SELECT
  -- Total revenue from paid orders
  (SELECT COALESCE(SUM(total_ghs_minor), 0) FROM orders WHERE status != 'pending' AND status != 'cancelled')::bigint AS total_revenue_ghs_minor,
  -- Total order count
  (SELECT COUNT(*) FROM orders)::int AS total_orders,
  -- Pending order count
  (SELECT COUNT(*) FROM orders WHERE status = 'pending')::int AS pending_orders,
  -- Paid order count
  (SELECT COUNT(*) FROM orders WHERE status = 'paid')::int AS paid_orders,
  -- Shipped order count
  (SELECT COUNT(*) FROM orders WHERE status = 'shipped')::int AS shipped_orders,
  -- Delivered order count
  (SELECT COUNT(*) FROM orders WHERE status = 'delivered')::int AS delivered_orders,
  -- Total customers (users with customer role)
  (SELECT COUNT(DISTINCT user_id) FROM user_roles WHERE role = 'customer')::int AS total_customers,
  -- Total products
  (SELECT COUNT(*) FROM products)::int AS total_products;

-- name: GetRecentOrders :many
-- Returns the most recent orders with customer information
SELECT
  o.id,
  o.user_id,
  o.status,
  o.total_ghs_minor,
  o.paystack_reference,
  o.created_at,
  u.email AS customer_email,
  u.name AS customer_name
FROM orders o
JOIN users u ON u.id = o.user_id
ORDER BY o.created_at DESC
LIMIT $1;

-- name: ListAllOrders :many
-- Returns paginated list of all orders with optional status filter
SELECT
  o.id,
  o.user_id,
  o.status,
  o.subtotal_ghs_minor,
  o.shipping_ghs_minor,
  o.total_ghs_minor,
  o.paystack_reference,
  o.created_at,
  o.updated_at,
  u.email AS customer_email,
  u.name AS customer_name,
  -- Extract phone from shipping_address JSONB
  (o.shipping_address->>'phone') AS customer_phone
FROM orders o
JOIN users u ON u.id = o.user_id
WHERE (sqlc.narg('status')::text = '' OR o.status = sqlc.narg('status'))
  AND (sqlc.narg('date_from')::timestamptz IS NULL OR o.created_at >= sqlc.narg('date_from')::timestamptz)
  AND (sqlc.narg('date_to')::timestamptz IS NULL OR o.created_at <= sqlc.narg('date_to')::timestamptz)
ORDER BY o.created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountAllOrders :one
-- Counts orders with optional status and date filters
SELECT COUNT(*)
FROM orders
WHERE (sqlc.narg('status')::text = '' OR status = sqlc.narg('status'))
  AND (sqlc.narg('date_from')::timestamptz IS NULL OR created_at >= sqlc.narg('date_from')::timestamptz)
  AND (sqlc.narg('date_to')::timestamptz IS NULL OR created_at <= sqlc.narg('date_to')::timestamptz);

-- name: GetOrderAnalytics :one
-- Returns order counts and revenue grouped by status
SELECT
  (SELECT COUNT(*) FROM orders WHERE status = 'pending')::int AS pending_count,
  (SELECT COUNT(*) FROM orders WHERE status = 'paid')::int AS paid_count,
  (SELECT COUNT(*) FROM orders WHERE status = 'fulfilled')::int AS fulfilled_count,
  (SELECT COUNT(*) FROM orders WHERE status = 'shipped')::int AS shipped_count,
  (SELECT COUNT(*) FROM orders WHERE status = 'delivered')::int AS delivered_count,
  (SELECT COUNT(*) FROM orders WHERE status = 'cancelled')::int AS cancelled_count,
  (SELECT COUNT(*) FROM orders WHERE status = 'refunded')::int AS refunded_count,
  (SELECT COALESCE(SUM(total_ghs_minor), 0) FROM orders WHERE status = 'paid')::bigint AS paid_revenue_ghs_minor,
  (SELECT COALESCE(SUM(total_ghs_minor), 0) FROM orders WHERE status = 'delivered')::bigint AS delivered_revenue_ghs_minor,
  (SELECT COALESCE(SUM(total_ghs_minor), 0) FROM orders WHERE status IN ('paid', 'delivered'))::bigint AS total_completed_revenue_ghs_minor;

-- name: ListAllCustomers :many
-- Returns paginated list of all customers with order counts and lifetime value
SELECT DISTINCT
  u.id,
  u.email,
  u.name,
  u.email_verified,
  u.created_at,
  (SELECT COUNT(*) FROM orders WHERE user_id = u.id)::int AS order_count,
  (SELECT COALESCE(SUM(total_ghs_minor), 0) FROM orders WHERE user_id = u.id AND status IN ('paid', 'delivered'))::bigint AS lifetime_value_ghs_minor,
  -- Last order date
  (SELECT MAX(created_at) FROM orders WHERE user_id = u.id) AS last_order_at
FROM users u
WHERE EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id AND ur.role = 'customer')
ORDER BY u.created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountAllCustomers :one
-- Counts total customers
SELECT COUNT(DISTINCT user_id)
FROM user_roles
WHERE role = 'customer';

-- name: GetCustomerStats :one
-- Returns customer statistics
SELECT
  (SELECT COUNT(DISTINCT user_id) FROM user_roles WHERE role = 'customer')::int AS total_customers,
  (SELECT COUNT(DISTINCT user_id) FROM user_roles ur
   JOIN users u ON u.id = ur.user_id
   WHERE ur.role = 'customer'
   AND EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id AND o.created_at >= NOW() - INTERVAL '30 days')
  )::int AS active_customers_30d,
  -- Customers who have placed at least one order
  (SELECT COUNT(DISTINCT o.user_id) FROM orders o
   JOIN user_roles ur ON ur.user_id = o.user_id
   WHERE ur.role = 'customer'
  )::int AS customers_with_orders;

-- name: GetProductStats :one
-- Returns product statistics
SELECT
  (SELECT COUNT(*) FROM products)::int AS total_products,
  -- Products with no stock (assuming stock would be tracked; for now using placeholder)
  (SELECT COUNT(*) FROM products WHERE tags && ARRAY['out-of-stock'])::int AS out_of_stock_count,
  -- Products tagged as low stock
  (SELECT COUNT(*) FROM products WHERE tags && ARRAY['low-stock'])::int AS low_stock_count,
  -- Count by tone
  (SELECT COUNT(*) FROM products WHERE tone = 'lavender')::int AS lavender_count,
  (SELECT COUNT(*) FROM products WHERE tone = 'rose')::int AS rose_count,
  (SELECT COUNT(*) FROM products WHERE tone = 'cream')::int AS cream_count,
  (SELECT COUNT(*) FROM products WHERE tone = 'ink')::int AS ink_count;

-- name: GetTopProducts :many
-- Returns top-selling products by revenue
SELECT
  p.id,
  p.slug,
  p.name,
  p.brand_id,
  p.category_id,
  p.price_ghs_minor,
  p.tone,
  p.image_path,
  c.name AS brand_name,
  cat.label AS category_label,
  COALESCE(SUM(oi.qty), 0)::int AS total_sold,
  COALESCE(SUM(oi.qty * oi.unit_price_ghs_minor), 0)::bigint AS total_revenue_ghs_minor
FROM products p
LEFT JOIN order_items oi ON oi.product_id = p.id
LEFT JOIN orders o ON o.id = oi.order_id AND o.status IN ('paid', 'delivered')
LEFT JOIN brands c ON c.id = p.brand_id
LEFT JOIN categories cat ON cat.id = p.category_id
GROUP BY p.id, p.slug, p.name, p.brand_id, p.category_id, p.price_ghs_minor, p.tone, p.image_path, c.name, cat.label
HAVING COALESCE(SUM(oi.qty), 0) > 0
ORDER BY total_revenue_ghs_minor DESC
LIMIT $1;

-- name: GetRevenueByDate :many
-- Returns revenue aggregated by date (daily, weekly, or monthly)
SELECT
  date_trunc(sqlc.narg('granularity'), o.created_at)::date AS revenue_date,
  COUNT(*)::int AS order_count,
  COALESCE(SUM(o.total_ghs_minor), 0)::bigint AS revenue_ghs_minor
FROM orders o
WHERE o.status IN ('paid', 'delivered')
  AND o.created_at >= sqlc.narg('date_from')::timestamptz
  AND o.created_at <= sqlc.narg('date_to')::timestamptz
GROUP BY date_trunc(sqlc.narg('granularity'), o.created_at)
ORDER BY revenue_date DESC;

-- name: GetRevenueByCategory :many
-- Returns revenue breakdown by product category
SELECT
  cat.id,
  cat.slug,
  cat.label AS category_name,
  COUNT(DISTINCT o.id)::int AS order_count,
  COALESCE(SUM(oi.qty * oi.unit_price_ghs_minor), 0)::bigint AS revenue_ghs_minor
FROM categories cat
LEFT JOIN products p ON p.category_id = cat.id
LEFT JOIN order_items oi ON oi.product_id = p.id
LEFT JOIN orders o ON o.id = oi.order_id AND o.status IN ('paid', 'delivered')
GROUP BY cat.id, cat.slug, cat.label
ORDER BY revenue_ghs_minor DESC;

-- name: UpdateOrderStatus :one
-- Updates the status of an order
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
          total_ghs_minor, paystack_reference, paystack_transaction_id,
          shipping_address, created_at, updated_at;

-- name: InsertOrderHistory :one
-- Creates an audit record for an order status change
INSERT INTO order_history (order_id, old_status, new_status, changed_by_user_id, note)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, order_id, old_status, new_status, changed_by_user_id, changed_at, note;
