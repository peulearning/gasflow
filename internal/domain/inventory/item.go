package inventory

import "time"

// Item representa o saldo de um produto em um depósito.
type Item struct {
	ID        string
	DepositID string
	ProductID string
	Quantity  int // total físico
	Reserved  int // reservado (pedidos aprovados ainda não entregues)
	UpdatedAt time.Time
}

// Available retorna o saldo disponível para novos pedidos.
func (i *Item) Available() int {
	return i.Quantity - i.Reserved
}

// Reserve tenta reservar `qty` unidades. Retorna erro se não houver saldo.
func (i *Item) Reserve(qty int) error {
	if qty <= 0 {
		return ErrNegativeQuantity
	}
	if i.Available() < qty {
		return ErrInsufficientStock
	}
	i.Reserved += qty
	i.UpdatedAt = time.Now().UTC()
	return nil
}

// Release libera `qty` unidades da reserva (cancelamento ou entrega).
func (i *Item) Release(qty int) error {
	if qty <= 0 {
		return ErrNegativeQuantity
	}
	if i.Reserved < qty {
		return ErrInvalidReservation
	}
	i.Reserved -= qty
	i.UpdatedAt = time.Now().UTC()
	return nil
}

// Consume confirma a saída física do estoque após entrega.
// Deve ser chamado junto com Release.
func (i *Item) Consume(qty int) error {
	if err := i.Release(qty); err != nil {
		return err
	}
	if i.Quantity < qty {
		return ErrInsufficientStock
	}
	i.Quantity -= qty
	return nil
}

// Receive registra entrada de estoque.
func (i *Item) Receive(qty int) error {
	if qty <= 0 {
		return ErrNegativeQuantity
	}
	i.Quantity += qty
	i.UpdatedAt = time.Now().UTC()
	return nil
}

// LowStockThreshold define o limite para alerta de estoque baixo.
const LowStockThreshold = 10

func (i *Item) IsLowStock() bool {
	return i.Available() < LowStockThreshold
}