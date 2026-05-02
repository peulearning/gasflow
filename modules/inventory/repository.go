package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gasflow/internal/domain/inventory"
)

var ErrItemNotFound = errors.New("inventory: item não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetItem(ctx context.Context, depositID, productID string) (inventory.Item, error) {
	const q = `SELECT id, deposit_id, product_id, quantity, reserved, updated_at
		         FROM inventory_items WHERE deposit_id = ? AND product_id = ?`
	row := r.db.QueryRowContext(ctx, q, depositID, productID)
	return scanItem(row)
}

func (r *Repository) GetItemByID(ctx context.Context, id string) (inventory.Item, error) {
	const q = `SELECT id, deposit_id, product_id, quantity, reserved, updated_at FROM inventory_items WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, id)
	return scanItem(row)
}

// UpdateItem atualiza quantity e reserved atomicamente com lock pessimista.
func (r *Repository) UpdateItem(ctx context.Context, item inventory.Item) error {
	const q = `UPDATE inventory_items SET quantity=?, reserved=?, updated_at=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, q, item.Quantity, item.Reserved, time.Now().UTC(), item.ID)
	return wrapErr("update_item", err)
}

// ReserveWithLock reserva estoque dentro de uma transação com SELECT FOR UPDATE.
func (r *Repository) ReserveWithLock(ctx context.Context, depositID, productID string, qty int) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return wrapErr("begin_tx", err)
	}
	defer tx.Rollback()

	const q = `SELECT id, deposit_id, product_id, quantity, reserved, updated_at
		         FROM inventory_items WHERE deposit_id = ? AND product_id = ? FOR UPDATE`
	row := tx.QueryRowContext(ctx, q, depositID, productID)

	var item inventory.Item
	if err := row.Scan(&item.ID, &item.DepositID, &item.ProductID, &item.Quantity, &item.Reserved, &item.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrItemNotFound
		}
		return err
	}

	if err := item.Reserve(qty); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE inventory_items SET reserved=?, updated_at=? WHERE id=?`,
		item.Reserved, time.Now().UTC(), item.ID); err != nil {
		return wrapErr("update_reserved", err)
	}

	return tx.Commit()
}

// ConsumeWithLock baixa o estoque físico após entrega (release + decrement).
func (r *Repository) ConsumeWithLock(ctx context.Context, depositID, productID string, qty int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return wrapErr("begin_tx", err)
	}
	defer tx.Rollback()

	const q = `SELECT id, deposit_id, product_id, quantity, reserved, updated_at
		         FROM inventory_items WHERE deposit_id = ? AND product_id = ? FOR UPDATE`
	row := tx.QueryRowContext(ctx, q, depositID, productID)

	var item inventory.Item
	if err := row.Scan(&item.ID, &item.DepositID, &item.ProductID, &item.Quantity, &item.Reserved, &item.UpdatedAt); err != nil {
		return err
	}

	if err := item.Consume(qty); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE inventory_items SET quantity=?, reserved=?, updated_at=? WHERE id=?`,
		item.Quantity, item.Reserved, time.Now().UTC(), item.ID)
	if err != nil {
		return wrapErr("consume", err)
	}
	return tx.Commit()
}

// ReleaseWithLock libera reserva sem baixar estoque (cancelamento).
func (r *Repository) ReleaseWithLock(ctx context.Context, depositID, productID string, qty int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const q = `SELECT id, deposit_id, product_id, quantity, reserved, updated_at
		         FROM inventory_items WHERE deposit_id = ? AND product_id = ? FOR UPDATE`
	row := tx.QueryRowContext(ctx, q, depositID, productID)

	var item inventory.Item
	if err := row.Scan(&item.ID, &item.DepositID, &item.ProductID, &item.Quantity, &item.Reserved, &item.UpdatedAt); err != nil {
		return err
	}

	if err := item.Release(qty); err != nil {
		return err
	}

	tx.ExecContext(ctx, `UPDATE inventory_items SET reserved=?, updated_at=? WHERE id=?`,
		item.Reserved, time.Now().UTC(), item.ID)
	return tx.Commit()
}

func (r *Repository) ListByDeposit(ctx context.Context, depositID string) ([]inventory.Item, error) {
	const q = `SELECT id, deposit_id, product_id, quantity, reserved, updated_at FROM inventory_items WHERE deposit_id = ?`
	rows, err := r.db.QueryContext(ctx, q, depositID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []inventory.Item
	for rows.Next() {
		var item inventory.Item
		rows.Scan(&item.ID, &item.DepositID, &item.ProductID, &item.Quantity, &item.Reserved, &item.UpdatedAt)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) LowStockItems(ctx context.Context) ([]inventory.Item, error) {
	const q = `SELECT id, deposit_id, product_id, quantity, reserved, updated_at
		         FROM inventory_items WHERE (quantity - reserved) < ?`
	rows, err := r.db.QueryContext(ctx, q, inventory.LowStockThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []inventory.Item
	for rows.Next() {
		var item inventory.Item
		rows.Scan(&item.ID, &item.DepositID, &item.ProductID, &item.Quantity, &item.Reserved, &item.UpdatedAt)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Receive(ctx context.Context, depositID, productID string, qty int) error {
	const q = `UPDATE inventory_items SET quantity = quantity + ?, updated_at = ? WHERE deposit_id = ? AND product_id = ?`
	res, err := r.db.ExecContext(ctx, q, qty, time.Now().UTC(), depositID, productID)
	if err != nil {
		return wrapErr("receive", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *Repository) ListDeposits(ctx context.Context) ([]inventory.Deposit, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, city, created_at FROM inventory_deposits`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []inventory.Deposit
	for rows.Next() {
		var d inventory.Deposit
		rows.Scan(&d.ID, &d.Name, &d.City, &d.CreatedAt)
		result = append(result, d)
	}
	return result, rows.Err()
}

func scanItem(row *sql.Row) (inventory.Item, error) {
	var item inventory.Item
	err := row.Scan(&item.ID, &item.DepositID, &item.ProductID, &item.Quantity, &item.Reserved, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return inventory.Item{}, ErrItemNotFound
	}
	return item, err
}

func wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("inventory.repo.%s: %w", op, err)
}