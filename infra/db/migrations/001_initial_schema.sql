-- 001_initial_schema.sql
-- GasFlow — schema inicial completo

CREATE TABLE IF NOT EXISTS users (
    id          CHAR(36) PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    email       VARCHAR(200) UNIQUE NOT NULL,
    password    VARCHAR(200) NOT NULL,
    role        ENUM('admin','operational','financial') NOT NULL DEFAULT 'operational',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_role  (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Clientes ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS clients (
    id          CHAR(36) PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    document    VARCHAR(20)  NOT NULL,
    phone       VARCHAR(20),
    email       VARCHAR(200),
    status      ENUM('active','inactive','blocked') NOT NULL DEFAULT 'active',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_document (document),
    INDEX idx_status (status),
    INDEX idx_name   (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS addresses (
    id          CHAR(36) PRIMARY KEY,
    client_id   CHAR(36) NOT NULL,
    street      VARCHAR(300) NOT NULL,
    city        VARCHAR(100) NOT NULL,
    state       CHAR(2)      NOT NULL,
    zipcode     VARCHAR(10),
    region      VARCHAR(50),
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    INDEX idx_client (client_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS contracts (
    id              CHAR(36) PRIMARY KEY,
    client_id       CHAR(36) NOT NULL,
    product_id      CHAR(36) NOT NULL,
    price_cents     INT      NOT NULL,
    payment_method  ENUM('cash','credit','billing') NOT NULL,
    valid_until     DATE,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id) REFERENCES clients(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Produtos ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS products (
    id               CHAR(36) PRIMARY KEY,
    name             VARCHAR(100) NOT NULL,
    weight_kg        DECIMAL(6,2) NOT NULL,
    unit_price_cents INT          NOT NULL,
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Motoristas e veículos ───────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS drivers (
    id          CHAR(36) PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    license     VARCHAR(20),
    phone       VARCHAR(20),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_license (license)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS vehicles (
    id           CHAR(36) PRIMARY KEY,
    plate        VARCHAR(10) NOT NULL,
    driver_id    CHAR(36),
    capacity_kg  DECIMAL(8,2),
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_plate (plate),
    FOREIGN KEY (driver_id) REFERENCES drivers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Pedidos ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS orders (
    id           CHAR(36) PRIMARY KEY,
    client_id    CHAR(36) NOT NULL,
    address_id   CHAR(36) NOT NULL,
    product_id   CHAR(36) NOT NULL,
    quantity     INT      NOT NULL,
    status       ENUM('received','approved','separated','in_route','delivered','cancelled','rescheduled') NOT NULL DEFAULT 'received',
    driver_id    CHAR(36),
    scheduled_at DATETIME,
    delivered_at DATETIME,
    notes        TEXT,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id)  REFERENCES clients(id),
    FOREIGN KEY (address_id) REFERENCES addresses(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    INDEX idx_status_created  (status, created_at),
    INDEX idx_client          (client_id),
    INDEX idx_driver          (driver_id),
    INDEX idx_scheduled       (scheduled_at),
    INDEX idx_delivered       (delivered_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS order_status_history (
    id          CHAR(36) PRIMARY KEY,
    order_id    CHAR(36)     NOT NULL,
    from_status VARCHAR(30),
    to_status   VARCHAR(30)  NOT NULL,
    changed_by  CHAR(36),
    reason      VARCHAR(300),
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    INDEX idx_order (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Estoque ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS inventory_deposits (
    id         CHAR(36) PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    city       VARCHAR(100),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS inventory_items (
    id         CHAR(36) PRIMARY KEY,
    deposit_id CHAR(36) NOT NULL,
    product_id CHAR(36) NOT NULL,
    quantity   INT NOT NULL DEFAULT 0,
    reserved   INT NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (deposit_id) REFERENCES inventory_deposits(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    UNIQUE KEY uq_deposit_product (deposit_id, product_id),
    INDEX idx_deposit (deposit_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Cobrança ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS charges (
    id           CHAR(36) PRIMARY KEY,
    order_id     CHAR(36) NOT NULL,
    client_id    CHAR(36) NOT NULL,
    amount_cents INT      NOT NULL,
    status       ENUM('pending','paid','overdue','cancelled') NOT NULL DEFAULT 'pending',
    due_date     DATE     NOT NULL,
    paid_at      DATETIME,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id)  REFERENCES orders(id),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    INDEX idx_status_due (status, due_date),
    INDEX idx_client     (client_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Analytics (read model) ───────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS analytics_daily (
    date            DATE     NOT NULL,
    deposit_id      CHAR(36) NOT NULL DEFAULT '',
    total_orders    INT      NOT NULL DEFAULT 0,
    delivered       INT      NOT NULL DEFAULT 0,
    delayed         INT      NOT NULL DEFAULT 0,
    rescheduled     INT      NOT NULL DEFAULT 0,
    volume_kg       DECIMAL(10,2) NOT NULL DEFAULT 0,
    revenue_cents   BIGINT   NOT NULL DEFAULT 0,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (date, deposit_id),
    INDEX idx_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Auditoria ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS audit_logs (
    id         CHAR(36) PRIMARY KEY,
    entity     VARCHAR(50) NOT NULL,
    entity_id  CHAR(36)    NOT NULL,
    action     VARCHAR(50) NOT NULL,
    user_id    CHAR(36),
    payload    JSON,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_entity  (entity, entity_id),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;