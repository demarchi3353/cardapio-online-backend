-- Habilita extensão para geração de UUIDs
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. ESTABELECIMENTOS
CREATE TABLE establishments (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name          VARCHAR(255) NOT NULL,
  description   TEXT,
  address       TEXT,
  image_key     VARCHAR(512),
  banner_key    VARCHAR(512),
  phone         VARCHAR(20),
  created_at    TIMESTAMP   NOT NULL DEFAULT now(),
  updated_at    TIMESTAMP   NOT NULL DEFAULT now()
);

-- 2. CATEGORIAS DE PRODUTOS
CREATE TABLE product_categories (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  establishment_id UUID        NOT NULL
    REFERENCES establishments(id)
    ON DELETE CASCADE,
  name             VARCHAR(100) NOT NULL,
  description      TEXT,
  created_at       TIMESTAMP    NOT NULL DEFAULT now()
);

-- 3. PRODUTOS
CREATE TABLE products (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  establishment_id UUID        NOT NULL
    REFERENCES establishments(id)
    ON DELETE CASCADE,
  category_id      UUID
    REFERENCES product_categories(id)
    ON DELETE SET NULL,
  name             VARCHAR(255) NOT NULL,
  description      TEXT,
  price_cents      INTEGER     NOT NULL,
  image_key        VARCHAR(512),
  banner_key       VARCHAR(512),
  is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMP   NOT NULL DEFAULT now(),
  updated_at       TIMESTAMP   NOT NULL DEFAULT now()
);

-- 4. INGREDIENTES
CREATE TABLE ingredients (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name          VARCHAR(100) NOT NULL UNIQUE,
  description   TEXT
);

-- 5. ASSOCIAÇÃO PRODUTO ↔ INGREDIENTE (N:N)
CREATE TABLE product_ingredients (
  product_id    UUID        NOT NULL
    REFERENCES products(id)
    ON DELETE CASCADE,
  ingredient_id UUID        NOT NULL
    REFERENCES ingredients(id)
    ON DELETE RESTRICT,
  quantity      VARCHAR(50),
  PRIMARY KEY (product_id, ingredient_id)
);

-- 6. CLIENTES
CREATE TABLE customers (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name          VARCHAR(255) NOT NULL,
  email         VARCHAR(255) NOT NULL UNIQUE,
  phone         VARCHAR(20),
  password_hash VARCHAR(255) NOT NULL,
  created_at    TIMESTAMP   NOT NULL DEFAULT now(),
  updated_at    TIMESTAMP   NOT NULL DEFAULT now()
);

-- 7. CUPONS DE DESCONTO
CREATE TABLE coupons (
  code          VARCHAR(50) PRIMARY KEY,
  description   TEXT,
  discount_type VARCHAR(10) NOT NULL
    CHECK (discount_type IN ('percent','fixed')),
  discount_value INTEGER    NOT NULL,
  valid_from    DATE        NOT NULL,
  valid_until   DATE        NOT NULL,
  max_uses      INTEGER,
  created_at    TIMESTAMP   NOT NULL DEFAULT now()
);

-- 8. REGISTRO DE USO DE CUPONS
CREATE TABLE coupon_redemptions (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  coupon_code   VARCHAR(50) NOT NULL
    REFERENCES coupons(code)
    ON DELETE CASCADE,
  customer_id   UUID        NOT NULL
    REFERENCES customers(id)
    ON DELETE CASCADE,
  order_id      UUID,
  redeemed_at   TIMESTAMP   NOT NULL DEFAULT now(),
  UNIQUE (coupon_code, customer_id, order_id)
);

-- 9. CONTAS DE FIDELIDADE
CREATE TABLE loyalty_accounts (
  customer_id    UUID        PRIMARY KEY
    REFERENCES customers(id)
    ON DELETE CASCADE,
  points_balance INTEGER     NOT NULL DEFAULT 0,
  tier           VARCHAR(50) NOT NULL DEFAULT 'standard',
  created_at     TIMESTAMP   NOT NULL DEFAULT now(),
  updated_at     TIMESTAMP   NOT NULL DEFAULT now()
);

-- 10. TRANSAÇÕES DE PONTOS
CREATE TABLE loyalty_transactions (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id   UUID        NOT NULL
    REFERENCES customers(id)
    ON DELETE CASCADE,
  points_delta  INTEGER     NOT NULL,
  reason        VARCHAR(100),
  created_at    TIMESTAMP   NOT NULL DEFAULT now()
);

-- 11. PEDIDOS
CREATE TABLE orders (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id       UUID        NOT NULL
    REFERENCES customers(id)
    ON DELETE RESTRICT,
  establishment_id  UUID        NOT NULL
    REFERENCES establishments(id)
    ON DELETE RESTRICT,
  coupon_code       VARCHAR(50)
    REFERENCES coupons(code)
    ON DELETE SET NULL,
  loyalty_points    INTEGER     NOT NULL DEFAULT 0,
  total_cents       BIGINT      NOT NULL,
  status            VARCHAR(20) NOT NULL DEFAULT 'PENDING'
    CHECK (status IN ('PENDING','PROCESSING','COMPLETED','CANCELLED','FAILED')),
  ordered_at        TIMESTAMP   NOT NULL DEFAULT now(),
  processed_at      TIMESTAMP,
  completed_at      TIMESTAMP,
  updated_at        TIMESTAMP   NOT NULL DEFAULT now()
);

-- 12. ITENS DO PEDIDO
CREATE TABLE order_items (
  order_id          UUID        NOT NULL
    REFERENCES orders(id)
    ON DELETE CASCADE,
  product_id        UUID        NOT NULL
    REFERENCES products(id)
    ON DELETE RESTRICT,
  quantity          INTEGER     NOT NULL CHECK (quantity > 0),
  unit_price_cents  BIGINT      NOT NULL,
  total_price_cents BIGINT      NOT NULL,
  PRIMARY KEY (order_id, product_id)
);

-- 13. EVENTOS DE PEDIDO (audit trail)
CREATE TABLE order_events (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id    UUID        NOT NULL
    REFERENCES orders(id)
    ON DELETE CASCADE,
  event_type  VARCHAR(20) NOT NULL,
  payload     JSONB,
  occurred_at TIMESTAMP   NOT NULL DEFAULT now()
);

-- Índices adicionais para performance (exemplos)
CREATE INDEX idx_products_estab ON products(establishment_id);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
