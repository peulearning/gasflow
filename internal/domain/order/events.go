package order

import "time"

// Eventos de domínio publicados no RabbitMQ.
// Cada evento é imutável e contém apenas dados serializáveis.

type EventOrderCreated struct {
	OrderID   string    `json:"order_id"`
	ClientID  string    `json:"client_id"`
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	OccurredAt time.Time `json:"occurred_at"`
}

type EventOrderStatusChanged struct {
	OrderID    string    `json:"order_id"`
	ClientID   string    `json:"client_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	ChangedBy  string    `json:"changed_by"`
	OccurredAt time.Time `json:"occurred_at"`
}

type EventOrderDelivered struct {
	OrderID     string    `json:"order_id"`
	ClientID    string    `json:"client_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	DriverID    string    `json:"driver_id"`
	DeliveredAt time.Time `json:"delivered_at"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type EventOrderCancelled struct {
	OrderID    string    `json:"order_id"`
	ClientID   string    `json:"client_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

// Routing keys para o RabbitMQ.
const (
	RoutingKeyOrderCreated       = "order.created"
	RoutingKeyOrderStatusChanged = "order.status_changed"
	RoutingKeyOrderDelivered     = "order.delivered"
	RoutingKeyOrderCancelled     = "order.cancelled"
)