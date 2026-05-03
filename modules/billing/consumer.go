package billing

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type orderDeliveredEvent struct {
	OrderID     string    `json:"order_id"`
	ClientID    string    `json:"client_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	DeliveredAt time.Time `json:"delivered_at"`
}

type mqSubscriber interface {
	Subscribe(exchange, queue, routingKey string) (<-chan amqp.Delivery, error)
}

type Consumer struct {
	svc         *Service
	mq          mqSubscriber
	pricePerUnit int64 // preço unitário padrão em centavos (simplificado)
}

func NewConsumer(svc *Service, mq mqSubscriber, pricePerUnit int64) *Consumer {
	return &Consumer{svc: svc, mq: mq, pricePerUnit: pricePerUnit}
}

func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.mq.Subscribe("orders", "billing.order_delivered", "order.delivered")
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
				c.handleOrderDelivered(ctx, msg)
			}
		}
	}()

	log.Info().Msg("billing.consumer: started")
	return nil
}

func (c *Consumer) handleOrderDelivered(ctx context.Context, msg amqp.Delivery) {
	var evt orderDeliveredEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		log.Error().Err(err).Msg("billing.consumer: unmarshal order_delivered")
		msg.Nack(false, false)
		return
	}

	amountCents := c.pricePerUnit * int64(evt.Quantity)
	_, err := c.svc.GenerateCharge(ctx, evt.OrderID, evt.ClientID, amountCents)
	if err != nil {
		log.Error().Err(err).Str("order_id", evt.OrderID).Msg("billing.consumer: generate charge failed")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
}