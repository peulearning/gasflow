package analytics

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// ── KPIs ─────────────────────────────────────────────────────────────────────

type KPISummary struct {
	Period     PeriodInfo    `json:"period"`
	Deliveries DeliveryKPIs  `json:"deliveries"`
	Inventory  InventoryKPIs `json:"inventory"`
	Billing    BillingKPIs   `json:"billing"`
}

type PeriodInfo    struct {
	From string `json:"from"`
	To   string `json:"to"`
}
type DeliveryKPIs  struct {
	Total       int     `json:"total"`
	Delivered   int     `json:"delivered"`
	Delayed     int     `json:"delayed"`
	Rescheduled int     `json:"rescheduled"`
	SLARate     float64 `json:"sla_rate"`
	VolumeKG    float64 `json:"volume_kg"`
}
type InventoryKPIs struct {
	TotalUnits     int `json:"total_units"`
	Reserved       int `json:"reserved"`
	Available      int `json:"available"`
	LowStockAlerts int `json:"low_stock_alerts"`
}
type BillingKPIs   struct {
	RevenueCents       int64 `json:"revenue_cents"`
	OverdueCount       int   `json:"overdue_count"`
	OverdueAmountCents int64 `json:"overdue_amount_cents"`
}

func (r *Repository) GetKPIs(ctx context.Context, from, to time.Time) (KPISummary, error) {
	kpi := KPISummary{Period: PeriodInfo{From: from.Format("2006-01-02"), To: to.Format("2006-01-02")}}

	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(status='delivered'),0),
		       COALESCE(SUM(status='rescheduled'),0),
		       COALESCE(SUM(status='in_route' AND scheduled_at IS NOT NULL AND scheduled_at < NOW()),0)
		FROM orders WHERE created_at BETWEEN ? AND ?`, from, to,
	).Scan(&kpi.Deliveries.Total, &kpi.Deliveries.Delivered,
		&kpi.Deliveries.Rescheduled, &kpi.Deliveries.Delayed)

	if kpi.Deliveries.Total > 0 {
		kpi.Deliveries.SLARate = float64(kpi.Deliveries.Delivered) / float64(kpi.Deliveries.Total) * 100
	}

	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(o.quantity * p.weight_kg),0)
		FROM orders o JOIN products p ON p.id=o.product_id
		WHERE o.status='delivered' AND o.delivered_at BETWEEN ? AND ?`, from, to,
	).Scan(&kpi.Deliveries.VolumeKG)

	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity),0),COALESCE(SUM(reserved),0) FROM inventory_items`,
	).Scan(&kpi.Inventory.TotalUnits, &kpi.Inventory.Reserved)
	kpi.Inventory.Available = kpi.Inventory.TotalUnits - kpi.Inventory.Reserved

	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_items WHERE (quantity-reserved)<10`,
	).Scan(&kpi.Inventory.LowStockAlerts)

	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_cents),0) FROM charges WHERE status='paid' AND paid_at BETWEEN ? AND ?`,
		from, to).Scan(&kpi.Billing.RevenueCents)

	r.db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(amount_cents),0) FROM charges WHERE status='overdue'`,
	).Scan(&kpi.Billing.OverdueCount, &kpi.Billing.OverdueAmountCents)

	return kpi, nil
}

// ── Entregas ──────────────────────────────────────────────────────────────────

type DeliveryRow struct {
	OrderID     string  `json:"order_id"`
	ClientName  string  `json:"client_name"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Status      string  `json:"status"`
	DriverName  string  `json:"driver_name"`
	ScheduledAt *string `json:"scheduled_at,omitempty"`
	DeliveredAt *string `json:"delivered_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type DeliveryFilter struct {
	From     *time.Time
	To       *time.Time
	Status   string
	DriverID string
	Limit    int
	Offset   int
}

func (r *Repository) ListDeliveries(ctx context.Context, f DeliveryFilter) ([]DeliveryRow, int, error) {
	where, args := "WHERE 1=1", []any{}
	if f.From != nil {
		where += " AND o.created_at>=?"
		args = append(args, *f.From)
	}
	if f.To != nil {
		where += " AND o.created_at<=?"
		args = append(args, *f.To)
	}
	if f.Status != "" {
		where += " AND o.status=?"
		args = append(args, f.Status)
	}
	if f.DriverID != "" {
		where += " AND o.driver_id=?"
		args = append(args, f.DriverID)
	}
	var total int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders o "+where, args...).Scan(&total)
	if f.Limit <= 0 {
		f.Limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.id, c.name, p.name, o.quantity, o.status,
		       COALESCE(d.name,''), o.scheduled_at, o.delivered_at, o.created_at
		FROM orders o
		JOIN clients  c ON c.id=o.client_id
		JOIN products p ON p.id=o.product_id
		LEFT JOIN drivers d ON d.id=o.driver_id
		`+where+` ORDER BY o.created_at DESC LIMIT ? OFFSET ?`,
		append(args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []DeliveryRow
	for rows.Next() {
		var row DeliveryRow
		var scheduledAt, deliveredAt, createdAt sql.NullTime
		rows.Scan(&row.OrderID, &row.ClientName, &row.ProductName, &row.Quantity, &row.Status,
			&row.DriverName, &scheduledAt, &deliveredAt, &createdAt)
		if scheduledAt.Valid {
			s := scheduledAt.Time.Format(time.RFC3339)
			row.ScheduledAt = &s
		}
		if deliveredAt.Valid {
			s := deliveredAt.Time.Format(time.RFC3339)
			row.DeliveredAt = &s
		}
		if createdAt.Valid {
			row.CreatedAt = createdAt.Time.Format(time.RFC3339)
		}
		result = append(result, row)
	}
	return result, total, rows.Err()
}

// ── Performance motoristas ────────────────────────────────────────────────────

type DriverPerf struct {
	DriverID   string  `json:"driver_id"`
	DriverName string  `json:"driver_name"`
	Total      int     `json:"total"`
	Delivered  int     `json:"delivered"`
	Delayed    int     `json:"delayed"`
	SLARate    float64 `json:"sla_rate"`
}

func (r *Repository) DriverPerformance(ctx context.Context, from, to time.Time) ([]DriverPerf, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.name, COUNT(o.id),
		       COALESCE(SUM(o.status='delivered'),0),
		       COALESCE(SUM(o.status='in_route' AND o.scheduled_at IS NOT NULL AND o.scheduled_at < NOW()),0)
		FROM orders o JOIN drivers d ON d.id=o.driver_id
		WHERE o.created_at BETWEEN ? AND ?
		GROUP BY d.id, d.name ORDER BY 3 DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []DriverPerf
	for rows.Next() {
		var dp DriverPerf
		rows.Scan(&dp.DriverID, &dp.DriverName, &dp.Total, &dp.Delivered, &dp.Delayed)
		if dp.Total > 0 {
			dp.SLARate = float64(dp.Delivered) / float64(dp.Total) * 100
		}
		result = append(result, dp)
	}
	return result, rows.Err()
}

// ── Top clientes ──────────────────────────────────────────────────────────────

type TopClient struct {
	ClientID    string `json:"client_id"`
	ClientName  string `json:"client_name"`
	TotalOrders int    `json:"total_orders"`
	TotalCents  int64  `json:"total_cents"`
}

func (r *Repository) TopClientsByVolume(ctx context.Context, limit int) ([]TopClient, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.name, COUNT(o.id), COALESCE(SUM(ch.amount_cents),0)
		FROM clients c
		LEFT JOIN orders  o  ON o.client_id=c.id
		LEFT JOIN charges ch ON ch.client_id=c.id AND ch.status='paid'
		GROUP BY c.id, c.name ORDER BY 4 DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []TopClient
	for rows.Next() {
		var tc TopClient
		rows.Scan(&tc.ClientID, &tc.ClientName, &tc.TotalOrders, &tc.TotalCents)
		result = append(result, tc)
	}
	return result, rows.Err()
}