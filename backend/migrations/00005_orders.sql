-- +goose Up
CREATE TABLE orders (
    id                        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                   uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status                    text NOT NULL CHECK (status IN ('pending','paid','fulfilled','shipped','delivered','cancelled','refunded')),
    subtotal_ghs_minor        bigint NOT NULL CHECK (subtotal_ghs_minor >= 0),
    shipping_ghs_minor        bigint NOT NULL CHECK (shipping_ghs_minor >= 0),
    total_ghs_minor           bigint NOT NULL CHECK (total_ghs_minor >= 0),
    paystack_reference        text NOT NULL UNIQUE,
    paystack_transaction_id   text,
    shipping_address          jsonb NOT NULL,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

CREATE TABLE order_items (
    id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id                    uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id                  uuid NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    qty                         int NOT NULL CHECK (qty >= 1),
    unit_price_ghs_minor        bigint NOT NULL CHECK (unit_price_ghs_minor >= 0),
    product_name_snapshot       text NOT NULL,
    product_brand_snapshot      text NOT NULL DEFAULT '',
    product_image_snapshot      text NOT NULL DEFAULT '',
    created_at                  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- +goose Down
DROP TABLE order_items;
DROP TABLE orders;
