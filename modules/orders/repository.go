package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gasflow/internal/domain/order"
)

var ErrOrderNotFound = errors.New("orders: não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, o order.Order) error {
	const q = `
		INSERT INTO orders (id, client_id, address_id, product_id, quantity, status, driver_id, scheduled_at, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		o.ID, o.ClientID, o.AddressID, o.ProductID, o.Quantity,
		o.Status, nullStr(o.DriverID), o.ScheduledAt, o.Notes,
		o.CreatedAt, o.UpdatedAt)
	return wrapErr("create", err)
}

func (r *Repository) GetByID(ctx context.Context, id string) (order.Order, error) {
	const q = `
		SELECT id, client_id, address_id, product_id, quantity, status,
		       COALESCE(driver_id,''), scheduled_at, delivered_at, COALESCE(notes,''), created_at, updated_at
		FROM orders WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, id)
	return scanOrder(row)
}

type ListFilter struct {
	Status   string
	ClientID string
	DriverID string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]order.Order, int, error) {
	where, args := buildWhere(f)

	var total int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders "+where, args...).Scan(&total)

	if f.Limit <= 0 {
		f.Limit = 20
	}
	q := `SELECT id, client_id, address_id, product_id, quantity, status,
		         COALESCE(driver_id,''), scheduled_at, delivered_at, COALESCE(notes,''), created_at, updated_at
		  FROM orders ` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, wrapErr("list", err)
	}
	defer rows.Close()

	var result []order.Order
	for rows.Next() {
		var o order.Order
		var scheduledAt, deliveredAt sql.NullTime
		err := rows.Scan(&o.ID, &o.ClientID, &o.AddressID, &o.ProductID, &o.Quantity,
			&o.Status, &o.DriverID, &scheduledAt, &deliveredAt, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		if scheduledAt.Valid {
			o.ScheduledAt = &scheduledAt.Time
		}
		if deliveredAt.Valid {
			o.DeliveredAt = &deliveredAt.Time
		}
		result = append(result, o)
	}
	return result, total, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, o order.Order) error {
	const q = `UPDATE orders SET status=?, driver_id=?, delivered_at=?, updated_at=? WHERE id=?`
	res, err := r.db.ExecContext(ctx, q,
		o.Status, nullStr(o.DriverID), o.DeliveredAt, time.Now().UTC(), o.ID)
	if err != nil {
		return wrapErr("update_status", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *Repository) SaveStatusHistory(ctx context.Context, h order.StatusHistory) error {
	const q = `
		INSERT INTO order_status_history (id, order_id, from_status, to_status, changed_by, reason, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		h.ID, h.OrderID, string(h.FromStatus), string(h.ToStatus),
		h.ChangedBy, h.Reason, h.CreatedAt)
	return wrapErr("save_history", err)
}

func (r *Repository) GetHistory(ctx context.Context, orderID string) ([]order.StatusHistory, error) {
	const q = `SELECT id, order_id, COALESCE(from_status,''), to_status, COALESCE(changed_by,''), COALESCE(reason,''), created_at
		         FROM order_status_history WHERE order_id = ? ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, q, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []order.StatusHistory
	for rows.Next() {
		var h order.StatusHistory
		rows.Scan(&h.ID, &h.OrderID, &h.FromStatus, &h.ToStatus, &h.ChangedBy, &h.Reason, &h.CreatedAt)
		result = append(result, h)
	}
	return result, rows.Err()
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func buildWhere(f ListFilter) (string, []any) {
	where := "WHERE 1=1"
	var args []any
	if f.Status != "" {
		where += " AND status = ?"
		args = append(args, f.Status)
	}
	if f.ClientID != "" {
		where += " AND client_id = ?"
		args = append(args, f.ClientID)
	}
	if f.DriverID != "" {
		where += " AND driver_id = ?"
		args = append(args, f.DriverID)
	}
	if f.From != nil {
		where += " AND created_at >= ?"
		args = append(args, *f.From)
	}
	if f.To != nil {
		where += " AND created_at <= ?"
		args = append(args, *f.To)
	}
	return where, args
}

func scanOrder(row *sql.Row) (order.Order, error) {
	var o order.Order
	var scheduledAt, deliveredAt sql.NullTime
	err := row.Scan(&o.ID, &o.ClientID, &o.AddressID, &o.ProductID, &o.Quantity,
		&o.Status, &o.DriverID, &scheduledAt, &deliveredAt, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return order.Order{}, ErrOrderNotFound
	}
	if err != nil {
		return order.Order{}, err
	}
	if scheduledAt.Valid {
		o.ScheduledAt = &scheduledAt.Time
	}
	if deliveredAt.Valid {
		o.DeliveredAt = &deliveredAt.Time
	}
	return o, nil
}

func nullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("orders.repo.%s: %w", op, err)
}