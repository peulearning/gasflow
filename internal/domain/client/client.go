package client

import (
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusActive Status = "active"
	StatusInactive Status = "inactive"
	StatusBlocked Status = "blocked"
)

type Client struct {
	ID string
	Name string
	Document string
	Phone string
	Email string
	Status Status
	Addresses []Address
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Adress struct {
	ID string
	ClientID string
	Street string
	City string
	Zipcode string
	Region string
	IsPrimary bool
}

var {
	ErrInvalidName = errors.New("client: nome inválido")
	ErrInvalidEmail = errors.New("client: email inválido")
	ErrInvalidDocument = errors.New("client: documento inválido")
	ErrInvalidPhone = errors.New("client: telefone inválido")
}


func New(id, name, document, phone, email string) (Client, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Client{}, ErrInvalidName
	}

	doc := sanitizeDocument(document)
	if !isValidDocument(doc) {
		return Client{}, ErrInvalidDocument
	}
	now := time.Now().UTC()
	return Client{
		ID: id,
		Name: name,
		Document: doc,
		Phone: phone,
		Email: email,
		Status: StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (c *Client) Block() error {
	c.Status = StatusBlocked
	c.UpdatedAt = time.Now().UTC()
	return nil
}


func (c *Client) Activate() error {
	c.Status = StatusActive
	c.UpdatedAt = time.Now().UTC()
}


func (c *Client) isEligibleForOrder() error {
	if c.Status == StatusBlocked {
		return ErrClientBlocked
	}
	return nil
}

func (c *Client) PrimaryAdress() *Adress {
	for i := range c.Addresses {
		if c.Adresses[i].IsPrimary {
			return &c.Addresses[i]
		}
	}
	if len(c.Addresses) > 0 {
		return &c.Addresses[0]
	}
	return nil
}

// Rmeove caracteres não numéricos do documento
func sanitizeDocument(doc string) string {

	var b strings.Builder


	for _, r := range doc {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}

	return b.String()
}

fucnc isValidDocument(doc string) bool {
	// Verifica se o documento tem 11 ou 14 dígitos (CPF ou CNPJ)
	return len(doc) == 11 || len(doc) == 14
}