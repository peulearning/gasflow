package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// Config agrupa as configurações de conexão com o RabbitMQ.
type Config struct {
	URL string // amqp://user:pass@host:5672/
}

// Exchanges declaradas no sistema.
const (
	ExchangeOrders    = "orders"
	ExchangeInventory = "inventory"
	ExchangeBilling   = "billing"
)

// Client encapsula a conexão e o canal AMQP.
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// Connect abre a conexão e o canal, e declara os exchanges.
func Connect(cfg Config) (*Client, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq: channel: %w", err)
	}

	c := &Client{conn: conn, channel: ch}

	for _, exchange := range []string{ExchangeOrders, ExchangeInventory, ExchangeBilling} {
		if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
			c.Close()
			return nil, fmt.Errorf("rabbitmq: declare exchange %s: %w", exchange, err)
		}
	}

	log.Info().Str("url", cfg.URL).Msg("rabbitmq: connected")
	return c, nil
}

// MustConnect conecta ou encerra o processo.
func MustConnect(cfg Config) *Client {
	var c *Client
	var err error

	// Retry simples para aguardar o container subir.
	for attempts := 0; attempts < 10; attempts++ {
		c, err = Connect(cfg)
		if err == nil {
			return c
		}
		log.Warn().Err(err).Int("attempt", attempts+1).Msg("rabbitmq: retrying...")
		time.Sleep(3 * time.Second)
	}
	log.Fatal().Err(err).Msg("rabbitmq: failed to connect")
	return nil
}

// Publish serializa e publica um evento em um exchange com a routing key fornecida.
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rabbitmq: marshal: %w", err)
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	}

	if err := c.channel.PublishWithContext(ctx, exchange, routingKey, false, false, msg); err != nil {
		return fmt.Errorf("rabbitmq: publish %s/%s: %w", exchange, routingKey, err)
	}

	log.Debug().
		Str("exchange", exchange).
		Str("routing_key", routingKey).
		Msg("rabbitmq: message published")

	return nil
}

// Subscribe declara uma fila, vincula ao exchange e retorna o canal de delivery.
func (c *Client) Subscribe(exchange, queueName, routingKey string) (<-chan amqp.Delivery, error) {
	// Declara a fila principal com suporte a DLQ.
	args := amqp.Table{
		"x-dead-letter-exchange": exchange + ".dlx",
		"x-message-ttl":          int32(86400000), // 24h
	}

	q, err := c.channel.QueueDeclare(queueName, true, false, false, false, args)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: declare queue %s: %w", queueName, err)
	}

	if err := c.channel.QueueBind(q.Name, routingKey, exchange, false, nil); err != nil {
		return nil, fmt.Errorf("rabbitmq: bind queue %s: %w", queueName, err)
	}

	// Prefetch de 1 para processamento sequencial por consumer.
	if err := c.channel.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("rabbitmq: qos: %w", err)
	}

	deliveries, err := c.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: consume %s: %w", queueName, err)
	}

	return deliveries, nil
}

// Close fecha o canal e a conexão.
func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// IsConnected verifica se a conexão está ativa.
func (c *Client) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}