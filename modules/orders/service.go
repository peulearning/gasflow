package orders

import (
	"context"
	"fmt"
	"time"

	"gasflow/internal/domain/order"
	"gasflow/internal/infra/messaging"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Service struct {
	repo *Repository
	mq   *messaging.Client
}

func NewService(repo *Repository, mq *messaging.Client) *Service {
	return &Service{repo: repo, mq: mq}
}

type CreateOrderInput struct {
	ClientID    string     `json:"client_id"`
	AddressID   string     `json:"address_id"`
	ProductID   string     `json:"product_id"`
	Quantity    int        `json:"quantity"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	Notes       string     `json:"notes,omitempty"`
}

type TransitionInput struct {
	OrderID   string `json:"-"`
	ToStatus  string `json:"status"`
	ChangedBy string `json:"-"` // preenchido pelo middleware de auth
	Reason    string `json:"reason,omitempty"`
	DriverID  string `json:"driver_id,omitempty"`
}

func (s *Service) Create(ctx context.Context, in CreateOrderInput) (order.Order, error) {
	o, err := order.New(uuid.NewString(), in.ClientID, in.AddressID, in.ProductID, in.Quantity)
	if err != nil {
		return order.Order{}, fmt.Errorf("orders.service: %w", err)
	}

	if in.ScheduledAt != nil {
		o.Schedule(*in.ScheduledAt)
	}
	o.Notes = in.Notes

	if err := s.repo.Create(ctx, o); err != nil {
		return order.Order{}, err
	}

	// Publica evento assíncrono para reserva de estoque.
	evt := order.EventOrderCreated{
		OrderID:    o.ID,
		ClientID:   o.ClientID,
		ProductID:  o.ProductID,
		Quantity:   o.Quantity,
		OccurredAt: time.Now().UTC(),
	}
	if err := s.mq.Publish(ctx, messaging.ExchangeOrders, order.RoutingKeyOrderCreated, evt); err != nil {
		log.Error().Err(err).Str("order_id", o.ID).Msg("orders.service: falha ao publicar OrderCreated")
	}

	log.Info().Str("order_id", o.ID).Str("client_id", o.ClientID).Msg("orders.service: pedido criado")
	return o, nil
}

func (s *Service) Transition(ctx context.Context, in TransitionInput) (order.Order, error) {
	o, err := s.repo.GetByID(ctx, in.OrderID)
	if err != nil {
		return order.Order{}, err
	}

	if in.DriverID != "" {
		o.AssignDriver(in.DriverID)
	}

	toStatus := order.Status(in.ToStatus)
	history, err := o.Transition(toStatus, in.ChangedBy, in.Reason)
	if err != nil {
		return order.Order{}, fmt.Errorf("orders.service: %w", err)
	}

	history.ID = uuid.NewString()

	if err := s.repo.UpdateStatus(ctx, o); err != nil {
		return order.Order{}, err
	}
	if err := s.repo.SaveStatusHistory(ctx, history); err != nil {
		log.Error().Err(err).Msg("orders.service: falha ao salvar histórico")
	}

	// Publica evento de mudança de status.
	evtChanged := order.EventOrderStatusChanged{
		OrderID:    o.ID,
		ClientID:   o.ClientID,
		FromStatus: string(history.FromStatus),
		ToStatus:   string(history.ToStatus),
		ChangedBy:  in.ChangedBy,
		OccurredAt: time.Now().UTC(),
	}
	s.mq.Publish(ctx, messaging.ExchangeOrders, order.RoutingKeyOrderStatusChanged, evtChanged)

	// Publica evento especializado de entrega.
	if toStatus == order.StatusDelivered {
		evtDelivered := order.EventOrderDelivered{
			OrderID:     o.ID,
			ClientID:    o.ClientID,
			ProductID:   o.ProductID,
			Quantity:    o.Quantity,
			DriverID:    o.DriverID,
			DeliveredAt: *o.DeliveredAt,
			OccurredAt:  time.Now().UTC(),
		}
		s.mq.Publish(ctx, messaging.ExchangeOrders, order.RoutingKeyOrderDelivered, evtDelivered)
	}

	// Publica evento de cancelamento com liberação de estoque.
	if toStatus == order.StatusCancelled {
		evtCancelled := order.EventOrderCancelled{
			OrderID:    o.ID,
			ClientID:   o.ClientID,
			ProductID:  o.ProductID,
			Quantity:   o.Quantity,
			Reason:     in.Reason,
			OccurredAt: time.Now().UTC(),
		}
		s.mq.Publish(ctx, messaging.ExchangeOrders, order.RoutingKeyOrderCancelled, evtCancelled)
	}

	log.Info().
		Str("order_id", o.ID).
		Str("from", string(history.FromStatus)).
		Str("to", string(history.ToStatus)).
		Msg("orders.service: status atualizado")

	return o, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (order.Order, error) {
	o, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}
	o.History, _ = s.repo.GetHistory(ctx, id)
	return o, nil
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]order.Order, int, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) GetHistory(ctx context.Context, orderID string) ([]order.StatusHistory, error) {
	return s.repo.GetHistory(ctx, orderID)
}