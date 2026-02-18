-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ─── Users ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL CHECK (role IN ('customer', 'business_owner')),
    fcm_token     VARCHAR(255),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ─── Businesses ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS businesses (
    id          UUID             PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id    UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(100)     NOT NULL,
    description TEXT,
    address     VARCHAR(255)     NOT NULL,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    -- PostGIS column for spatial queries (fallback / analytics)
    location    GEOGRAPHY(POINT, 4326) GENERATED ALWAYS AS (
                    ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography
                ) STORED,
    category    VARCHAR(50)  NOT NULL,
    fcm_token   VARCHAR(255),
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_businesses_location ON businesses USING GIST(location);
CREATE INDEX IF NOT EXISTS idx_businesses_category ON businesses(category);
CREATE INDEX IF NOT EXISTS idx_businesses_owner    ON businesses(owner_id);

-- ─── Orders ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS orders (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id       UUID        NOT NULL REFERENCES users(id),
    business_id       UUID        NOT NULL REFERENCES businesses(id),
    total_amount      DECIMAL(10,2) NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','paid','ready','completed','cancelled')),
    pin               CHAR(6),
    stripe_payment_id VARCHAR(255),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_customer  ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_business  ON orders(business_id);
CREATE INDEX IF NOT EXISTS idx_orders_status    ON orders(status);

-- ─── Order Items ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS order_items (
    id           UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id     UUID          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_name VARCHAR(255)  NOT NULL,
    quantity     INT           NOT NULL CHECK (quantity > 0),
    unit_price   DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0)
);
