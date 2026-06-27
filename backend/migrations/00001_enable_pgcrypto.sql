-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE health_marker (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO health_marker DEFAULT VALUES;

-- +goose Down
DROP TABLE health_marker;
DROP EXTENSION IF EXISTS pgcrypto;
