package clients

import (
	"context"
	"fmt"

	"gasflow/internal/domain/client"
	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateClientInput struct {
	Name     string `json:"name"`
	Document string `json:"document"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type AddAddressInput struct {
	ClientID  string `json:"-"`
	Street    string `json:"street"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zipcode   string `json:"zipcode"`
	Region    string `json:"region"`
	IsPrimary bool   `json:"is_primary"`
}

func (s *Service) Create(ctx context.Context, in CreateClientInput) (client.Client, error) {
	// Verifica duplicidade de documento.
	_, err := s.repo.GetByDocument(ctx, in.Document)
	if err == nil {
		return client.Client{}, fmt.Errorf("clients: documento já cadastrado")
	}

	c, err := client.New(uuid.NewString(), in.Name, in.Document, in.Phone, in.Email)
	if err != nil {
		return client.Client{}, err
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return client.Client{}, err
	}
	return c, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (client.Client, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return client.Client{}, err
	}
	addrs, _ := s.repo.ListAddresses(ctx, id)
	c.Addresses = addrs
	return c, nil
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]client.Client, int, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) Block(ctx context.Context, id string) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	c.Block()
	return s.repo.Update(ctx, c)
}

func (s *Service) Activate(ctx context.Context, id string) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	c.Activate()
	return s.repo.Update(ctx, c)
}

func (s *Service) AddAddress(ctx context.Context, in AddAddressInput) (client.Address, error) {
	addr := client.Address{
		ID:        uuid.NewString(),
		ClientID:  in.ClientID,
		Street:    in.Street,
		City:      in.City,
		Zipcode:   in.Zipcode,
		Region:    in.Region,
		IsPrimary: in.IsPrimary,
	}
	if err := s.repo.CreateAddress(ctx, addr); err != nil {
		return client.Address{}, err
	}
	return addr, nil
}