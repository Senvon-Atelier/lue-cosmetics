-- +goose Up
CREATE TABLE carts (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid REFERENCES users(id) ON DELETE CASCADE,
    guest_token  text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    -- exactly one of user_id / guest_token must be set
    CHECK ((user_id IS NULL) <> (guest_token IS NULL))
);
CREATE UNIQUE INDEX idx_carts_user_id ON carts(user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_carts_guest_token ON carts(guest_token) WHERE guest_token IS NOT NULL;

CREATE TABLE cart_items (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id              uuid NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id           uuid NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    qty                  int  NOT NULL CHECK (qty >= 1),
    unit_price_ghs_minor bigint NOT NULL CHECK (unit_price_ghs_minor >= 0),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    UNIQUE (cart_id, product_id)
);
CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);

-- +goose Down
DROP TABLE cart_items;
DROP TABLE carts;
