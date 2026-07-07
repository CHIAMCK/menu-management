CREATE TYPE merchant_status AS ENUM ('ACTIVE', 'INACTIVE');

CREATE TYPE menu_status AS ENUM ('ACTIVE', 'INACTIVE');

CREATE TYPE category_status AS ENUM ('ACTIVE', 'INACTIVE');

CREATE TYPE item_status AS ENUM ('ACTIVE', 'INACTIVE', 'ARCHIVED');

CREATE TYPE item_availability AS ENUM ('AVAILABLE', 'OUT_OF_STOCK');

CREATE TYPE order_status AS ENUM ('PENDING', 'CONFIRMED', 'COMPLETED', 'CANCELLED');

CREATE TABLE merchants (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255)     NOT NULL,
    status     merchant_status  NOT NULL DEFAULT 'INACTIVE',
    created_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE menus (
    id          BIGSERIAL PRIMARY KEY,
    merchant_id BIGINT       NOT NULL REFERENCES merchants (id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    status      menu_status  NOT NULL DEFAULT 'INACTIVE',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_menus_merchant_id_status ON menus (merchant_id, status);

CREATE UNIQUE INDEX idx_menus_one_active_per_merchant
    ON menus (merchant_id)
    WHERE status = 'ACTIVE';

CREATE TABLE categories (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255)     NOT NULL,
    menu_id    BIGINT           NOT NULL REFERENCES menus (id) ON DELETE CASCADE,
    status     category_status  NOT NULL DEFAULT 'INACTIVE',
    created_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_menu_id_status_created_at ON categories (menu_id, status, created_at);

CREATE TABLE items (
    id           BIGSERIAL PRIMARY KEY,
    merchant_id  BIGINT            NOT NULL REFERENCES merchants (id),
    name         VARCHAR(255)      NOT NULL,
    price        NUMERIC(12, 2)    NOT NULL CHECK (price >= 0),
    status       item_status       NOT NULL DEFAULT 'INACTIVE',
    availability item_availability NOT NULL DEFAULT 'AVAILABLE',
    category_id  BIGINT            NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ       NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_category_id_status_created_at ON items (category_id, status, created_at);
CREATE INDEX idx_items_merchant_id_status ON items (merchant_id, status);

CREATE TABLE orders (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT         NOT NULL REFERENCES users (id),
    merchant_id  BIGINT         NOT NULL REFERENCES merchants (id),
    status       order_status   NOT NULL DEFAULT 'PENDING',
    total_amount NUMERIC(12, 2) NOT NULL CHECK (total_amount >= 0),
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_merchant_id ON orders (merchant_id);

CREATE TABLE order_items (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT         NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    item_id    BIGINT         NOT NULL REFERENCES items (id),
    quantity   INTEGER        NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12, 2) NOT NULL CHECK (unit_price >= 0),
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (order_id, item_id)
);

CREATE INDEX idx_order_items_item_id ON order_items (item_id);
