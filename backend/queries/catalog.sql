-- name: ListCategories :many
SELECT id, slug, label, sort_order
FROM categories
ORDER BY sort_order ASC, label ASC;

-- name: ListBrands :many
SELECT id, slug, name
FROM brands
ORDER BY name ASC;

-- name: GetBrandByID :one
SELECT id, slug, name FROM brands WHERE id = $1;

-- name: GetProductBySlug :one
SELECT id, slug, name, brand_id, category_id, price_ghs_minor, was_price_ghs_minor,
       tone, size, rating, review_count, tags, image_path, created_at
FROM products
WHERE slug = $1;

-- name: GetProductByID :one
SELECT id, slug, name, brand_id, category_id, price_ghs_minor, was_price_ghs_minor,
       tone, size, rating, review_count, tags, image_path, created_at
FROM products
WHERE id = $1;

-- name: CountProducts :one
SELECT count(*)
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
LEFT JOIN brands     b ON b.id = p.brand_id
WHERE (sqlc.narg('category_slug')::text IS NULL OR c.slug = sqlc.narg('category_slug'))
  AND (sqlc.narg('brand_slug')::text    IS NULL OR b.slug = sqlc.narg('brand_slug'))
  AND (sqlc.narg('tag')::text           IS NULL OR p.tags && ARRAY[sqlc.narg('tag')::text])
  AND (sqlc.narg('q')::text             IS NULL OR p.name ILIKE '%' || sqlc.narg('q') || '%');

-- name: ListProductsByNewest :many
SELECT p.id, p.slug, p.name, p.brand_id, p.category_id, p.price_ghs_minor, p.was_price_ghs_minor,
       p.tone, p.size, p.rating, p.review_count, p.tags, p.image_path, p.created_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
LEFT JOIN brands     b ON b.id = p.brand_id
WHERE (sqlc.narg('category_slug')::text IS NULL OR c.slug = sqlc.narg('category_slug'))
  AND (sqlc.narg('brand_slug')::text    IS NULL OR b.slug = sqlc.narg('brand_slug'))
  AND (sqlc.narg('tag')::text           IS NULL OR p.tags && ARRAY[sqlc.narg('tag')::text])
  AND (sqlc.narg('q')::text             IS NULL OR p.name ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListProductsByPriceAsc :many
SELECT p.id, p.slug, p.name, p.brand_id, p.category_id, p.price_ghs_minor, p.was_price_ghs_minor,
       p.tone, p.size, p.rating, p.review_count, p.tags, p.image_path, p.created_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
LEFT JOIN brands     b ON b.id = p.brand_id
WHERE (sqlc.narg('category_slug')::text IS NULL OR c.slug = sqlc.narg('category_slug'))
  AND (sqlc.narg('brand_slug')::text    IS NULL OR b.slug = sqlc.narg('brand_slug'))
  AND (sqlc.narg('tag')::text           IS NULL OR p.tags && ARRAY[sqlc.narg('tag')::text])
  AND (sqlc.narg('q')::text             IS NULL OR p.name ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY p.price_ghs_minor ASC
LIMIT $1 OFFSET $2;

-- name: ListProductsByPriceDesc :many
SELECT p.id, p.slug, p.name, p.brand_id, p.category_id, p.price_ghs_minor, p.was_price_ghs_minor,
       p.tone, p.size, p.rating, p.review_count, p.tags, p.image_path, p.created_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
LEFT JOIN brands     b ON b.id = p.brand_id
WHERE (sqlc.narg('category_slug')::text IS NULL OR c.slug = sqlc.narg('category_slug'))
  AND (sqlc.narg('brand_slug')::text    IS NULL OR b.slug = sqlc.narg('brand_slug'))
  AND (sqlc.narg('tag')::text           IS NULL OR p.tags && ARRAY[sqlc.narg('tag')::text])
  AND (sqlc.narg('q')::text             IS NULL OR p.name ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY p.price_ghs_minor DESC
LIMIT $1 OFFSET $2;

-- name: ListProductsByRating :many
SELECT p.id, p.slug, p.name, p.brand_id, p.category_id, p.price_ghs_minor, p.was_price_ghs_minor,
       p.tone, p.size, p.rating, p.review_count, p.tags, p.image_path, p.created_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
LEFT JOIN brands     b ON b.id = p.brand_id
WHERE (sqlc.narg('category_slug')::text IS NULL OR c.slug = sqlc.narg('category_slug'))
  AND (sqlc.narg('brand_slug')::text    IS NULL OR b.slug = sqlc.narg('brand_slug'))
  AND (sqlc.narg('tag')::text           IS NULL OR p.tags && ARRAY[sqlc.narg('tag')::text])
  AND (sqlc.narg('q')::text             IS NULL OR p.name ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY p.rating DESC NULLS LAST
LIMIT $1 OFFSET $2;

-- name: ListProductsByName :many
SELECT p.id, p.slug, p.name, p.brand_id, p.category_id, p.price_ghs_minor, p.was_price_ghs_minor,
       p.tone, p.size, p.rating, p.review_count, p.tags, p.image_path, p.created_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
LEFT JOIN brands     b ON b.id = p.brand_id
WHERE (sqlc.narg('category_slug')::text IS NULL OR c.slug = sqlc.narg('category_slug'))
  AND (sqlc.narg('brand_slug')::text    IS NULL OR b.slug = sqlc.narg('brand_slug'))
  AND (sqlc.narg('tag')::text           IS NULL OR p.tags && ARRAY[sqlc.narg('tag')::text])
  AND (sqlc.narg('q')::text             IS NULL OR p.name ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY p.name ASC
LIMIT $1 OFFSET $2;
  
-- name: CreateCategory :one
INSERT INTO categories (slug, label, sort_order)
VALUES ($1, $2, 0)
RETURNING id, slug, label, sort_order;

-- name: CreateBrand :one
INSERT INTO brands (slug, name)
VALUES ($1, $2)
RETURNING id, slug, name;

-- name: CreateProduct :one
INSERT INTO products (slug, name, brand_id, category_id, price_ghs_minor, tags, image_path)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, slug, name, brand_id, category_id, price_ghs_minor, was_price_ghs_minor,
           tone, size, rating, review_count, tags, image_path, created_at;
