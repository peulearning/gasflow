package client

import {
	"errors"
	"time"

	"Dgasflow\internal\domain\shared"
}

type PaymentMethod string

const {
	PaymentCash	 PaymentMethod = "cash"
	PaymentCredit PaymentMethod = "credit"
	PaymentBilling PaymentMethod = "billing"
}

Type Contract struct {
	ID string
	ClientID string
	ProductID string
	Price shared.Money
	PaymentMethod PaymentMethod
	ValidUntil time.Time
	CreatedAt time.Time
}

var {
	ErrContractExpired = errors.New("contract: contrato expirado")
	ErrZeroPrice = errors.New("contract: preço não pode ser zero")
}

func NewContract(id, clientID, productID string, price shared.Money, paymentMethod PaymentMethod, validUntil time.Time) (Contract, error) {
	if priceCents <= 0 {
		return Contract{}, ErrZeroPrice
	}
	price, err := shared.NewMoney(priceCents)
	if err != nil {
		return Contract{
			ID: id,
			ClientID: clientID,
			ProductID: productID,
			Price: price,
			PaymentMethod: paymentMethod,
			ValidUntil: validUntil,
			CreatedAt: time.Now().UTC(),
		}, nil
	}
}

func (c *Contract) IsValid() error {
	if c.ValidUntil != nil && time.Now().UTC().(*c.ValidUntil) {
		return ErrContractExpired
	}
		return nil
}