package inventory

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type orderCreatedEvent struct {
	OrderID   string    `json:"order_id"`
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	OccurredAt time.Time `json:"occurred_at"`
}

type orderCancelledEvent struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type orderDeliveredEvent struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type mqSubscriber interface {
	Subscribe(exchange, queue, routingKey string) (<-chan amqp.Delivery, error)
}

type Consumer struct {
	svc *Service
	mq  mqSubscriber
}

func NewConsumer(svc *Service, mq mqSubscriber) *Consumer {
	return &Consumer{svc: svc, mq: mq}
}

func (c *Consumer) Start(ctx context.Context) error {
	// Reserva quando pedido é criado.
	created, err := c.mq.Subscribe("orders", "inventory.order_created", "order.created")
	if err != nil {
		return err
	}

	// Libera reserva quando pedido é cancelado.
	cancelled, err := c.mq.Subscribe("orders", "inventory.order_cancelled", "order.cancelled")
	if err != nil {
		return err
	}

	// Baixa estoque físico quando pedido é entregue.
	delivered, err := c.mq.Subscribe("orders", "inventory.order_delivered", "order.delivered")
	if err != nil {
		return err
	}

	go c.loop(ctx, created, c.handleOrderCreated)
	go c.loop(ctx, cancelled, c.handleOrderCancelled)
	go c.loop(ctx, delivered, c.handleOrderDelivered)

	log.Info().Msg("inventory.consumer: started")
	return nil
}

func (c *Consumer) loop(ctx context.Context, ch <-chan amqp.Delivery, fn func(context.Context, amqp.Delivery)) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fn(ctx, msg)
		}
	}
}

func (c *Consumer) handleOrderCreated(ctx context.Context, msg amqp.Delivery) {
	var evt orderCreatedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		log.Error().Err(err).Msg("inventory.consumer: unmarshal order_created")
		msg.Nack(false, false)
		return
	}

	if err := c.svc.Reserve(ctx, evt.OrderID, evt.ProductID, evt.Quantity); err != nil {
		log.Error().Err(err).Str("order_id", evt.OrderID).Msg("inventory.consumer: reserve failed")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
	log.Info().Str("order_id", evt.OrderID).Int("qty", evt.Quantity).Msg("inventory.consumer: stock reserved")
}

func (c *Consumer) handleOrderCancelled(ctx context.Context, msg amqp.Delivery) {
	var evt orderCancelledEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		msg.Nack(false, false)
		return
	}

	if err := c.svc.Release(ctx, evt.OrderID, evt.ProductID, evt.Quantity); err != nil {
		log.Error().Err(err).Str("order_id", evt.OrderID).Msg("inventory.consumer: release failed")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
	log.Info().Str("order_id", evt.OrderID).Msg("inventory.consumer: stock released (cancelled)")
}

func (c *Consumer) handleOrderDelivered(ctx context.Context, msg amqp.Delivery) {
	var evt orderDeliveredEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		msg.Nack(false, false)
		return
	}

	if err := c.svc.Consume(ctx, evt.ProductID, evt.Quantity); err != nil {
		log.Error().Err(err).Str("order_id", evt.OrderID).Msg("inventory.consumer: consume failed")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
	log.Info().Str("order_id", evt.OrderID).Msg("inventory.consumer: stock consumed (delivered)")
}