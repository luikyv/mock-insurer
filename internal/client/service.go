package client

import (
	"context"

	"gorm.io/gorm"
)

type Service struct {
	storage Storage
}

func NewService(db *gorm.DB) Service {
	return Service{storage: storage{db: db}}
}

func (s Service) Save(ctx context.Context, client *Client) error {
	return s.storage.Save(ctx, client)
}

func (s Service) Client(ctx context.Context, id string) (*Client, error) {
	return s.storage.Client(ctx, id)
}

func (s Service) Delete(ctx context.Context, id string) error {
	return s.storage.Delete(ctx, id)
}
