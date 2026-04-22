package order

import {
	"errors"
	"time"
}

var (
	ErrInvalidQuantity = errors.New("order : quantidade deve ser maior que zero")
	ErrMissingClient = errors.New("order : cliente é obrigatório")
	ErrMissingProduct = errors.New("order : produto é obrigatório")
	ErrOrderNoActive = errors.New("order : pedido não está ativo")
)

type StatusHistory struct {
	ID string
	OrderID string
	FromStatus Status
	ToStatus Status
	ChangedBy string
	Reason string
	CreatedAt time.Time
}

type Order struct {
	ID string
	ClientID string
	AddressID string
	ProductID string
	Quantity int
	Status Status
	DriverID string
	ScheduledAt *time.Time
	DeliveredAt *time.Time
	Notes string
	History []StatusHistory
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(clientId, addressId, productId string, quantity int) (Order, error){
	if clientId == ""(
		return Order{}, ErrMissingClient
	)
	if productId == "" {
		return Order{}, ErrMissingProduct
	}
	if quantity <= 0 {
		return Order{}, ErrInvalidQuantity
	}

	now := time.Now().UTC()

	return Order{
		ID:          "",
		ClientID:    clientId,
		AddressID:   addressId,
		ProductID:   productId,
		Quantity:    quantity,
		Status:      StatusPending,
		DriverID:    "",
		ScheduledAt: nil,
		DeliveredAt: nil,
		Notes:       "",
		History:     []StatusHistory{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

