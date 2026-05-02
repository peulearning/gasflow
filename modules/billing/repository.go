package billing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/peulearning/gasflow/internal/domain/billing"
)

var ErrChargeNotFound = errors.New("billing: cobrança não encontrada")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c billing.Charge) error {
	const q = `INSERT INTO charges (id, order_id, client_id, amount_cents, status, due_date, created_at)
		         VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, c.ID, c.OrderID, c.ClientID, c.Amount.Cents(), c.Status, c.DueDate, c.CreatedAt)
	return wrapErr("create", err)
}

func (r *Repository) GetByID(ctx context.Context, id string) (billing.Charge, error) {
	const q = `SELECT id, order_id, client_id, amount_cents, status, due_date, paid_at, created_at FROM charges WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, id)
	return scanCharge(row)
}

func (r *Repository) MarkPaid(ctx context.Context, id string) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, `UPDATE charges SET status='paid', paid_at=? WHERE id=? AND status != 'paid'`, now, id)
	if err != nil {
		return wrapErr("mark_paid", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrChargeNotFound
	}
	return nil
}

func (r *Repository) MarkOverdueAll(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE charges SET status='overdue' WHERE status='pending' AND due_date < CURDATE()`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

type ListFilter struct {
	ClientID string
	Status   string
	Limit    int
	Offset   int
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]billing.Charge, int, error) {
	where := "WHERE 1=1"
	var args []any
	if f.ClientID != "" {
		where += " AND client_id = ?"
		args = append(args, f.ClientID)
	}
	if f.Status != "" {
		where += " AND status = ?"
		args = append(args, f.Status)
	}
	var total int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM charges "+where, args...).Scan(&total)

	if f.Limit == 0 {
		f.Limit = 20
	}
	args = append(args, f.Limit, f.Offset)
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, order_id, client_id, amount_cents, status, due_date, paid_at, created_at FROM charges "+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?",
		args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []billing.Charge
	for rows.Next() {
		var c billing.Charge
		var paidAt sql.NullTime
		var amountCents int64
		var status billing.ChargeStatus
		rows.Scan(&c.ID, &c.OrderID, &c.ClientID, &amountCents, &status, &c.DueDate, &paidAt, &c.CreatedAt)
		c.Amount, _ = billing.MoneyFromCents(amountCents)
		c.Status = status
		if paidAt.Valid {
			c.PaidAt = &paidAt.Time
		}
		result = append(result, c)
	}
	return result, total, rows.Err()
}

func scanCharge(row *sql.Row) (billing.Charge, error) {
	var c billing.Charge
	var paidAt sql.NullTime
	var amountCents int64
	var status billing.ChargeStatus
	err := row.Scan(&c.ID, &c.OrderID, &c.ClientID, &amountCents, &status, &c.DueDate, &paidAt, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return billing.Charge{}, ErrChargeNotFound
	}
	if err != nil {
		return billing.Charge{}, err
	}
	c.Amount, _ = billing.MoneyFromCents(amountCents)
	c.Status = status
	if paidAt.Valid {
		c.PaidAt = &paidAt.Time
	}
	return c, nil
}

func wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("billing.repo.%s: %w", op, err)
}