	package orders

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// Consumer escuta eventos relevantes para o módulo de orders.
// Por ora: reserva confirmada pelo inventory → atualiza status para 'approved'.
type Consumer struct {
	svc *Service
	mq  interface {
		Subscribe(exchange, queue, routingKey string) (<-chan amqp.Delivery, error)
	}
}

func NewConsumer(svc *Service, mq interface {
	Subscribe(exchange, queue, routingKey string) (<-chan amqp.Delivery, error)
}) *Consumer {
	return &Consumer{svc: svc, mq: mq}
}

type stockReservedEvent struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// Start inicia os consumers em goroutines.
func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.mq.Subscribe("inventory", "orders.stock_reserved", "stock.reserved")
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-deliveries:
				if !ok {
					return
				}
				c.handleStockReserved(ctx, msg)
			}
		}
	}()

	log.Info().Msg("orders.consumer: started")
	return nil
}

func (c *Consumer) handleStockReserved(ctx context.Context, msg amqp.Delivery) {
	var evt stockReservedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		log.Error().Err(err).Msg("orders.consumer: unmarshal stock_reserved")
		msg.Nack(false, false) // envia para DLQ
		return
	}

	_, err := c.svc.Transition(ctx, TransitionInput{
		OrderID:   evt.OrderID,
		ToStatus:  "approved",
		ChangedBy: "system",
		Reason:    "estoque reservado automaticamente",
	})
	if err != nil {
		log.Error().Err(err).Str("order_id", evt.OrderID).Msg("orders.consumer: auto-approve failed")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
	log.Info().Str("order_id", evt.OrderID).Msg("orders.consumer: order auto-approved via stock reservation")
}