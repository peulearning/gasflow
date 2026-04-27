package order

import (
	"errors"
	"time"
)

var (
	ErrInvalidQuantity = errors.New("order: quantidade deve ser maior que zero")
	ErrMissingClient   = errors.New("order: cliente é obrigatório")
	ErrMissingProduct  = errors.New("order: produto é obrigatório")
	ErrOrderNotActive  = errors.New("order: pedido está em estado terminal")
)

// StatusHistory registra cada mudança de estado do pedido.
type StatusHistory struct {
	ID         string
	OrderID    string
	FromStatus Status
	ToStatus   Status
	ChangedBy  string
	Reason     string
	CreatedAt  time.Time
}

// Order é a entidade raiz do agregado de pedidos.
type Order struct {
	ID          string
	ClientID    string
	AddressID   string
	ProductID   string
	Quantity    int
	Status      Status
	DriverID    string
	ScheduledAt *time.Time
	DeliveredAt *time.Time
	Notes       string
	History     []StatusHistory
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func New(id, clientID, addressID, productID string, quantity int) (Order, error) {
	if clientID == "" {
		return Order{}, ErrMissingClient
	}
	if productID == "" {
		return Order{}, ErrMissingProduct
	}
	if quantity <= 0 {
		return Order{}, ErrInvalidQuantity
	}
	now := time.Now().UTC()
	return Order{
		ID:        id,
		ClientID:  clientID,
		AddressID: addressID,
		ProductID: productID,
		Quantity:  quantity,
		Status:    StatusReceived,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Transition realiza a transição de status, validando a FSM.
// Retorna o novo StatusHistory para ser persistido.
func (o *Order) Transition(to Status, changedBy, reason string) (StatusHistory, error) {
	if IsTerminal(o.Status) {
		return StatusHistory{}, ErrOrderNotActive
	}
	if err := CanTransitionTo(o.Status, to); err != nil {
		return StatusHistory{}, err
	}

	entry := StatusHistory{
		OrderID:    o.ID,
		FromStatus: o.Status,
		ToStatus:   to,
		ChangedBy:  changedBy,
		Reason:     reason,
		CreatedAt:  time.Now().UTC(),
	}

	o.Status = to
	o.UpdatedAt = time.Now().UTC()

	if to == StatusDelivered {
		now := time.Now().UTC()
		o.DeliveredAt = &now
	}

	return entry, nil
}

func (o *Order) AssignDriver(driverID string) {
	o.DriverID = driverID
	o.UpdatedAt = time.Now().UTC()
}

func (o *Order) Schedule(at time.Time) {
	o.ScheduledAt = &at
	o.UpdatedAt = time.Now().UTC()
}

// IsLate retorna true se o pedido está em rota mas passou do horário agendado.
func (o *Order) IsLate() bool {
	if o.ScheduledAt == nil {
		return false
	}
	return o.Status == StatusInRoute && time.Now().UTC().After(*o.ScheduledAt)
}