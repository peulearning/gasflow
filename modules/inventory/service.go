package inventory

import (
	"context"
	"fmt"

	"gasflow/internal/domain/inventory"
	"gasflow/infra/messaging"
	"github.com/rs/zerolog/log"
)

// stockReservedEvent é publicado após reserva bem-sucedida.
type stockReservedEvent struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type stockReleasedEvent struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Service struct {
	repo *Repository
	mq   *messaging.Client
	// depositID padrão — em produção, virá do pedido ou do contrato.
	defaultDepositID string
}

func NewService(repo *Repository, mq *messaging.Client, defaultDepositID string) *Service {
	return &Service{repo: repo, mq: mq, defaultDepositID: defaultDepositID}
}

type ReceiveInput struct {
	DepositID string `json:"deposit_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (s *Service) Receive(ctx context.Context, in ReceiveInput) error {
	if in.Quantity <= 0 {
		return inventory.ErrNegativeQuantity
	}
	return s.repo.Receive(ctx, in.DepositID, in.ProductID, in.Quantity)
}

func (s *Service) Reserve(ctx context.Context, orderID, productID string, qty int) error {
	depositID := s.defaultDepositID
	if err := s.repo.ReserveWithLock(ctx, depositID, productID, qty); err != nil {
		return fmt.Errorf("inventory.service: reserve: %w", err)
	}

	evt := stockReservedEvent{OrderID: orderID, ProductID: productID, Quantity: qty}
	if err := s.mq.Publish(ctx, messaging.ExchangeInventory, "stock.reserved", evt); err != nil {
		log.Error().Err(err).Msg("inventory.service: publish stock.reserved failed")
	}
	return nil
}

func (s *Service) Consume(ctx context.Context, productID string, qty int) error {
	return s.repo.ConsumeWithLock(ctx, s.defaultDepositID, productID, qty)
}

func (s *Service) Release(ctx context.Context, orderID, productID string, qty int) error {
	if err := s.repo.ReleaseWithLock(ctx, s.defaultDepositID, productID, qty); err != nil {
		return err
	}
	evt := stockReleasedEvent{OrderID: orderID, ProductID: productID, Quantity: qty}
	s.mq.Publish(ctx, messaging.ExchangeInventory, "stock.released", evt)
	return nil
}

func (s *Service) ListByDeposit(ctx context.Context, depositID string) ([]inventory.Item, error) {
	return s.repo.ListByDeposit(ctx, depositID)
}

func (s *Service) LowStockItems(ctx context.Context) ([]inventory.Item, error) {
	return s.repo.LowStockItems(ctx)
}

func (s *Service) ListDeposits(ctx context.Context) ([]inventory.Deposit, error) {
	return s.repo.ListDeposits(ctx)
}