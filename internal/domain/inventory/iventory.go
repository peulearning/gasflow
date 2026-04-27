package inventory

import (
	"errors"
	"time"
)

var (
	ErrInsufficientStock  = errors.New("inventory: estoque insuficiente")
	ErrInvalidReservation = errors.New("inventory: reserva inválida")
	ErrNegativeQuantity   = errors.New("inventory: quantidade não pode ser negativa")
)

// Deposit representa um depósito/centro de distribuição.
type Deposit struct {
	ID        string
	Name      string
	City      string
	CreatedAt time.Time
}

func NewDeposit(id, name, city string) Deposit {
	return Deposit{
		ID:        id,
		Name:      name,
		City:      city,
		CreatedAt: time.Now().UTC(),
	}
}