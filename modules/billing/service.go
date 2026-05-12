package billing

import (
	"context"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

type Invoice struct {
	ID        string  `json:"id"`
	ClientID  string  `json:"client_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

func (s *Service) CreateInvoice(ctx context.Context, invoice Invoice) error {
	// TODO: implementar persistência
	return nil
}

func (s *Service) ListInvoices(ctx context.Context) ([]Invoice, error) {
	return []Invoice{}, nil
}

func (s *Service) GetInvoiceByID(ctx context.Context, id string) (*Invoice, error) {
	return &Invoice{
		ID:       id,
		Status:   "pending",
		ClientID: "client-001",
		Amount:   100,
	}, nil
}