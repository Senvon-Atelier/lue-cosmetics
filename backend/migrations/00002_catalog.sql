-- +goose Up
CREATE TABLE categories (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        text NOT NULL UNIQUE,
    label       text NOT NULL,
    sort_order  int  NOT NULL DEFAULT 0
);

CREATE TABLE brands (
    id    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug  text NOT NULL UNIQUE,
    name  text NOT NULL
);

CREATE TABLE products (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                 text NOT NULL UNIQUE,
    name                 text NOT NULL,
    brand_id             uuid NOT NULL REFERENCES brands(id) ON DELETE RESTRICT,
    category_id          uuid NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    price_ghs_minor      bigint NOT NULL CHECK (price_ghs_minor >= 0),
    was_price_ghs_minor  bigint CHECK (was_price_ghs_minor IS NULL OR was_price_ghs_minor >= 0),
    tone                 text NOT NULL DEFAULT 'lavender',
    size                 text NOT NULL DEFAULT '',
    rating               numeric(2,1),
    review_count         int NOT NULL DEFAULT 0,
    tags                 text[] NOT NULL DEFAULT '{}',
    image_path           text NOT NULL DEFAULT '',
    created_at           timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_brand_id    ON products(brand_id);
CREATE INDEX idx_products_created_at  ON products(created_at DESC);
CREATE INDEX idx_products_price       ON products(price_ghs_minor);
CREATE INDEX idx_products_tags        ON products USING GIN (tags);

-- +goose Down
DROP TABLE products;
DROP TABLE brands;
DROP TABLE categories;
