package clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gasflow/internal/domain/client"
)

var ErrClientNotFound = errors.New("clients: não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c client.Client) error {
	const q = `
		INSERT INTO clients (id, name, document, phone, email, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		c.ID, c.Name, c.Document, c.Phone, c.Email, c.Status, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("clients.repo: create: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (client.Client, error) {
	const q = `SELECT id, name, document, phone, email, status, created_at, updated_at FROM clients WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, id)
	return scanClient(row)
}

func (r *Repository) GetByDocument(ctx context.Context, doc string) (client.Client, error) {
	const q = `SELECT id, name, document, phone, email, status, created_at, updated_at FROM clients WHERE document = ?`
	row := r.db.QueryRowContext(ctx, q, doc)
	return scanClient(row)
}

type ListFilter struct {
	Status string
	Search string
	Limit  int
	Offset int
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]client.Client, int, error) {
	where := "WHERE 1=1"
	args := []any{}

	if f.Status != "" {
		where += " AND status = ?"
		args = append(args, f.Status)
	}
	if f.Search != "" {
		where += " AND (name LIKE ? OR document LIKE ?)"
		args = append(args, "%"+f.Search+"%", "%"+f.Search+"%")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM clients "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("clients.repo: count: %w", err)
	}

	if f.Limit <= 0 {
		f.Limit = 20
	}
	args = append(args, f.Limit, f.Offset)
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, document, phone, email, status, created_at, updated_at FROM clients "+where+" ORDER BY name LIMIT ? OFFSET ?",
		args...)
	if err != nil {
		return nil, 0, fmt.Errorf("clients.repo: list: %w", err)
	}
	defer rows.Close()

	var result []client.Client
	for rows.Next() {
		var c client.Client
		if err := rows.Scan(&c.ID, &c.Name, &c.Document, &c.Phone, &c.Email, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, c)
	}
	return result, total, rows.Err()
}

func (r *Repository) Update(ctx context.Context, c client.Client) error {
	const q = `UPDATE clients SET name=?, phone=?, email=?, status=?, updated_at=? WHERE id=?`
	res, err := r.db.ExecContext(ctx, q, c.Name, c.Phone, c.Email, c.Status, time.Now().UTC(), c.ID)
	if err != nil {
		return fmt.Errorf("clients.repo: update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrClientNotFound
	}
	return nil
}

func (r *Repository) CreateAddress(ctx context.Context, a client.Address) error {
	const q = `
		INSERT INTO addresses (id, client_id, street, city, state, zipcode, region, is_primary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, a.ID, a.ClientID, a.Street, a.City, a.State, a.Zipcode, a.Region, a.IsPrimary)
	return err
}

func (r *Repository) ListAddresses(ctx context.Context, clientID string) ([]client.Address, error) {
	const q = `SELECT id, client_id, street, city, state, zipcode, region, is_primary FROM addresses WHERE client_id = ?`
	rows, err := r.db.QueryContext(ctx, q, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []client.Address
	for rows.Next() {
		var a client.Address
		rows.Scan(&a.ID, &a.ClientID, &a.Street, &a.City, &a.State, &a.Zipcode, &a.Region, &a.IsPrimary)
		result = append(result, a)
	}
	return result, rows.Err()
}

func scanClient(row *sql.Row) (client.Client, error) {
	var c client.Client
	err := row.Scan(&c.ID, &c.Name, &c.Document, &c.Phone, &c.Email, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return client.Client{}, ErrClientNotFound
	}
	return c, err
}