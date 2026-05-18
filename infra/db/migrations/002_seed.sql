-- 002_seed.sql — dados iniciais para desenvolvimento e testes do dashboard

INSERT IGNORE INTO users (id, name, email, password, role) VALUES
('u-admin-001', 'Admin GasFlow', 'admin@gasflow.com',
 '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password
 'admin'),
('u-op-001', 'Operador Silva', 'operador@gasflow.com',
 '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
 'operational'),
('u-fin-001', 'Financeiro Costa', 'financeiro@gasflow.com',
 '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
 'financial');

INSERT IGNORE INTO products (id, name, weight_kg, unit_price_cents, is_active) VALUES
('prod-p13',  'P13',  13.0, 10500, TRUE),
('prod-p20',  'P20',  20.0, 18000, TRUE),
('prod-p45',  'P45',  45.0, 38000, TRUE),
('prod-p190', 'P190', 190.0, 95000, TRUE);

INSERT IGNORE INTO inventory_deposits (id, name, city) VALUES
('dep-sp-001', 'Depósito Central SP', 'São Paulo'),
('dep-rj-001', 'Depósito RJ Sul',     'Rio de Janeiro'),
('dep-mg-001', 'Depósito BH Norte',   'Belo Horizonte');

INSERT IGNORE INTO inventory_items (id, deposit_id, product_id, quantity, reserved) VALUES
('item-001', 'dep-sp-001', 'prod-p13',  500, 30),
('item-002', 'dep-sp-001', 'prod-p45',  200, 15),
('item-003', 'dep-sp-001', 'prod-p190',  50,  5),
('item-004', 'dep-rj-001', 'prod-p13',  300, 20),
('item-005', 'dep-rj-001', 'prod-p45',  150,  8),
('item-006', 'dep-mg-001', 'prod-p13',   12,  3),  -- abaixo do threshold (low stock)
('item-007', 'dep-mg-001', 'prod-p45',    8,  0);  -- low stock

INSERT IGNORE INTO drivers (id, name, license, phone, is_active) VALUES
('drv-001', 'Carlos Andrade', 'SP-12345', '(11) 99001-0001', TRUE),
('drv-002', 'Marcos Lima',    'SP-67890', '(11) 99001-0002', TRUE),
('drv-003', 'João Ferreira',  'RJ-11111', '(21) 99001-0003', TRUE);

INSERT IGNORE INTO clients (id, name, document, phone, email, status) VALUES
('cli-001', 'Restaurante Sabor Brasil', '12345678000195', '(11) 3333-0001', 'contato@saborbrasil.com', 'active'),
('cli-002', 'Padaria Estrela',          '98765432000188', '(11) 3333-0002', 'padaria@estrela.com',     'active'),
('cli-003', 'Churrascaria do Zé',       '11122233000144', '(11) 3333-0003', 'ze@churrascaria.com',     'active'),
('cli-004', 'Hotel Paraíso',            '55566677000122', '(21) 3333-0004', 'hotel@paraiso.com',       'active'),
('cli-005', 'Bar do Nino',              '99988877000111', '(11) 3333-0005', 'nino@bar.com',            'blocked');

INSERT IGNORE INTO addresses (id, client_id, street, city, state, zipcode, region, is_primary) VALUES
('addr-001', 'cli-001', 'Rua das Flores, 100',  'São Paulo',       'SP', '01310-000', 'centro', TRUE),
('addr-002', 'cli-002', 'Av. Paulista, 500',    'São Paulo',       'SP', '01311-000', 'centro', TRUE),
('addr-003', 'cli-003', 'Rua XV de Nov., 200',  'São Paulo',       'SP', '01313-000', 'sul',    TRUE),
('addr-004', 'cli-004', 'Av. Atlântica, 1500',  'Rio de Janeiro',  'RJ', '22010-000', 'sul',    TRUE),
('addr-005', 'cli-005', 'Rua Augusta, 800',     'São Paulo',       'SP', '01305-000', 'centro', TRUE);


-- ─── Veículos ────────────────────────────────────────────────────────────────

INSERT IGNORE INTO vehicles (
    id,
    plate,
    driver_id,
    capacity_kg
) VALUES
('veh-001', 'ABC1D23', 'drv-001', 5000.00),
('veh-002', 'XYZ9K88', 'drv-002', 4500.00),
('veh-003', 'QWE7R65', 'drv-003', 6000.00);

-- ─── Contratos ───────────────────────────────────────────────────────────────

INSERT IGNORE INTO contracts (
    id,
    client_id,
    product_id,
    price_cents,
    payment_method,
    valid_until
) VALUES
('ctr-001', 'cli-001', 'prod-p13', 10500, 'billing', '2026-12-31'),
('ctr-002', 'cli-002', 'prod-p45', 38000, 'cash', '2026-12-31'),
('ctr-003', 'cli-003', 'prod-p20', 18000, 'credit', '2026-12-31');

-- ─── Orders ──────────────────────────────────────────────────────────────────

INSERT IGNORE INTO orders (
    id,
    client_id,
    address_id,
    product_id,
    quantity,
    status,
    driver_id,
    scheduled_at,
    delivered_at,
    notes
) VALUES
(
    'ord-001',
    'cli-001',
    'addr-001',
    'prod-p13',
    10,
    'delivered',
    'drv-001',
    NOW(),
    NOW(),
    'Entrega concluída'
),
(
    'ord-002',
    'cli-002',
    'addr-002',
    'prod-p45',
    4,
    'in_route',
    'drv-002',
    NOW(),
    NULL,
    'Saiu para entrega'
),
(
    'ord-003',
    'cli-003',
    'addr-003',
    'prod-p20',
    7,
    'approved',
    'drv-003',
    NOW(),
    NULL,
    'Pedido aprovado'
);

-- ─── Histórico de Status ─────────────────────────────────────────────────────

INSERT IGNORE INTO order_status_history (
    id,
    order_id,
    from_status,
    to_status,
    changed_by,
    reason
) VALUES
(
    'hist-001',
    'ord-001',
    'approved',
    'delivered',
    'u-admin-001',
    'Entrega realizada'
),
(
    'hist-002',
    'ord-002',
    'approved',
    'in_route',
    'u-op-001',
    'Motorista saiu para entrega'
);

-- ─── Cobranças ───────────────────────────────────────────────────────────────

INSERT IGNORE INTO charges (
    id,
    order_id,
    client_id,
    amount_cents,
    status,
    due_date,
    paid_at
) VALUES
(
    'chg-001',
    'ord-001',
    'cli-001',
    105000,
    'paid',
    CURDATE(),
    NOW()
),
(
    'chg-002',
    'ord-002',
    'cli-002',
    152000,
    'pending',
    CURDATE(),
    NULL
);

-- ─── Analytics ───────────────────────────────────────────────────────────────

INSERT IGNORE INTO analytics_daily (
    reference_date,
    deposit_id,
    total_orders,
    delivered,
    delayed_count,
    rescheduled_count,
    volume_kg,
    revenue_cents
) VALUES
(
    CURDATE(),
    'dep-sp-001',
    30,
    24,
    2,
    1,
    540.50,
    850000
);

-- ─── Audit Logs ──────────────────────────────────────────────────────────────

INSERT IGNORE INTO audit_logs (
    id,
    entity,
    entity_id,
    action,
    user_id,
    payload
) VALUES
(
    'audit-001',
    'orders',
    'ord-001',
    'CREATE',
    'u-admin-001',
    JSON_OBJECT('status', 'delivered')
),
(
    'audit-002',
    'clients',
    'cli-001',
    'UPDATE',
    'u-op-001',
    JSON_OBJECT('field', 'phone')
);