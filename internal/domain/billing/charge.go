package billing

import (
	"errors"
	"time"

	"github.com/gasflow/internal/domain/shared"
)

type ChargeStatus string

const (
	ChargePending   ChargeStatus = "pending"
	ChargePaid      ChargeStatus = "paid"
	ChargeOverdue   ChargeStatus = "overdue"
	ChargeCancelled ChargeStatus = "cancelled"
)

var (
	ErrChargeAlreadyPaid   = errors.New("billing: cobrança já foi paga")
	ErrInvalidChargeStatus = errors.New("billing: transição de status inválida")
)

// Charge representa uma cobrança gerada após entrega.
type Charge struct {
	ID        string
	OrderID   string
	ClientID  string
	Amount    shared.Money
	Status    ChargeStatus
	DueDate   time.Time
	PaidAt    *time.Time
	CreatedAt time.Time
}

// OverdueDays define após quantos dias a cobrança se torna inadimplente.
const OverdueDays = 30

func New(id, orderID, clientID string, amountCents int64, dueDate time.Time) (Charge, error) {
	amount, err := shared.NewMoney(amountCents)
	if err != nil {
		return Charge{}, err
	}
	return Charge{
		ID:        id,
		OrderID:   orderID,
		ClientID:  clientID,
		Amount:    amount,
		Status:    ChargePending,
		DueDate:   dueDate,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (c *Charge) MarkPaid() error {
	if c.Status == ChargePaid {
		return ErrChargeAlreadyPaid
	}
	now := time.Now().UTC()
	c.Status = ChargePaid
	c.PaidAt = &now
	return nil
}

func (c *Charge) MarkOverdue() error {
	if c.Status != ChargePending {
		return ErrInvalidChargeStatus
	}
	c.Status = ChargeOverdue
	return nil
}

func (c *Charge) Cancel() {
	c.Status = ChargeCancelled
}

// ShouldBeOverdue verifica se a cobrança deveria ser marcada como inadimplente.
func (c *Charge) ShouldBeOverdue() bool {
	return c.Status == ChargePending && time.Now().UTC().After(c.DueDate)
}

// EventChargeGenerated é publicado no RabbitMQ quando uma cobrança é criada.
type EventChargeGenerated struct {
	ChargeID    string    `json:"charge_id"`
	OrderID     string    `json:"order_id"`
	ClientID    string    `json:"client_id"`
	AmountCents int64     `json:"amount_cents"`
	DueDate     time.Time `json:"due_date"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// EventChargeOverdue é publicado quando a cobrança vence.
type EventChargeOverdue struct {
	ChargeID   string    `json:"charge_id"`
	ClientID   string    `json:"client_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

const (
	RoutingKeyChargeGenerated = "charge.generated"
	RoutingKeyChargeOverdue   = "charge.overdue"
)