-- +goose Up
CREATE TABLE addresses (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label       text NOT NULL DEFAULT 'Home',
    line1       text NOT NULL CHECK (length(line1) > 0),
    line2       text NOT NULL DEFAULT '',
    city        text NOT NULL CHECK (length(city) > 0),
    region      text NOT NULL CHECK (length(region) > 0),
    phone       text NOT NULL CHECK (length(phone) > 0),
    is_default  boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_addresses_user_id ON addresses(user_id);
CREATE UNIQUE INDEX idx_addresses_one_default_per_user
    ON addresses(user_id) WHERE is_default = true;

-- +goose Down
DROP TABLE addresses;
