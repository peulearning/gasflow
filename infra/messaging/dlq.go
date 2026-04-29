package messaging

import (
	"context"
	"fmt"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	maxRetries    = 5
	retryHeader   = "x-retry-count"
	baseDelayMS   = 1000 // 1s base para backoff exponencial
)

// DLQHandler processa mensagens que falharam e as recoloca na fila original
// com backoff exponencial, ou as descarta após maxRetries.
type DLQHandler struct {
	client *Client
}

func NewDLQHandler(client *Client) *DLQHandler {
	return &DLQHandler{client: client}
}

// SetupDLX declara o dead-letter exchange e a fila DLQ para um exchange dado.
func (d *DLQHandler) SetupDLX(exchange, queueName string) error {
	ch := d.client.channel
	dlxName := exchange + ".dlx"
	dlqName := queueName + ".dlq"

	// Declara o DLX como fanout (todas as mensagens mortas chegam aqui).
	if err := ch.ExchangeDeclare(dlxName, "fanout", true, false, false, false, nil); err != nil {
		return fmt.Errorf("dlq: declare dlx %s: %w", dlxName, err)
	}

	// Declara a DLQ.
	q, err := ch.QueueDeclare(dlqName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("dlq: declare dlq %s: %w", dlqName, err)
	}

	if err := ch.QueueBind(q.Name, "#", dlxName, false, nil); err != nil {
		return fmt.Errorf("dlq: bind dlq: %w", err)
	}

	log.Info().Str("dlx", dlxName).Str("dlq", dlqName).Msg("dlq: setup complete")
	return nil
}

// StartReprocessor inicia uma goroutine que consome a DLQ e retenta mensagens.
func (d *DLQHandler) StartReprocessor(ctx context.Context, dlqName, targetExchange, targetRoutingKey string) {
	go func() {
		deliveries, err := d.client.channel.Consume(dlqName, "dlq-reprocessor-"+dlqName, false, false, false, false, nil)
		if err != nil {
			log.Error().Err(err).Str("dlq", dlqName).Msg("dlq: failed to consume")
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-deliveries:
				if !ok {
					return
				}
				d.handleDLQMessage(ctx, msg, targetExchange, targetRoutingKey)
			}
		}
	}()
}

func (d *DLQHandler) handleDLQMessage(ctx context.Context, msg amqp.Delivery, exchange, routingKey string) {
	retryCount := getRetryCount(msg)

	if retryCount >= maxRetries {
		log.Error().
			Str("exchange", exchange).
			Str("routing_key", routingKey).
			Int("retry_count", retryCount).
			Str("body", string(msg.Body)).
			Msg("dlq: max retries reached — discarding message")
		msg.Ack(false)
		return
	}

	// Backoff exponencial: 1s, 2s, 4s, 8s, 16s.
	delay := time.Duration(math.Pow(2, float64(retryCount))) * time.Duration(baseDelayMS) * time.Millisecond
	log.Warn().
		Int("retry_count", retryCount+1).
		Dur("delay", delay).
		Str("exchange", exchange).
		Msg("dlq: retrying message")

	time.Sleep(delay)

	headers := amqp.Table{
		retryHeader: int32(retryCount + 1),
	}

	pub := amqp.Publishing{
		ContentType:  msg.ContentType,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Headers:      headers,
		Body:         msg.Body,
	}

	if err := d.client.channel.PublishWithContext(ctx, exchange, routingKey, false, false, pub); err != nil {
		log.Error().Err(err).Msg("dlq: failed to republish")
		msg.Nack(false, true) // recoloca na DLQ
		return
	}

	msg.Ack(false)
}

func getRetryCount(msg amqp.Delivery) int {
	if msg.Headers == nil {
		return 0
	}
	v, ok := msg.Headers[retryHeader]
	if !ok {
		return 0
	}
	count, _ := v.(int32)
	return int(count)
}