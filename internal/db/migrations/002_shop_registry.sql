CREATE TABLE IF NOT EXISTS user_shops (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shop_name      TEXT        NOT NULL,
    shop_namespace TEXT        NOT NULL DEFAULT 'default',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(shop_name, shop_namespace)
);
