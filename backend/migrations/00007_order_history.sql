-- +goose Up
CREATE TABLE order_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  old_status TEXT NOT NULL,
  new_status TEXT NOT NULL,
  changed_by_user_id UUID REFERENCES users(id),
  changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  note TEXT
);

CREATE INDEX idx_order_history_order_id ON order_history(order_id);
CREATE INDEX idx_order_history_changed_at ON order_history(changed_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_order_history_changed_at;
DROP INDEX IF EXISTS idx_order_history_order_id;
DROP TABLE IF EXISTS order_history;
