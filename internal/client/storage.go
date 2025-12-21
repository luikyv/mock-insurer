package client

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Storage interface {
	Save(context.Context, *Client) error
	Client(context.Context, string) (*Client, error)
	Delete(context.Context, string) error
}

type storage struct {
	db *gorm.DB
}

func (s storage) Save(ctx context.Context, client *Client) error {
	if err := s.db.WithContext(ctx).Save(client).Error; err != nil {
		return fmt.Errorf("could not save client: %w", err)
	}
	return nil
}

func (s storage) Client(ctx context.Context, id string) (*Client, error) {
	var client Client
	if err := s.db.WithContext(ctx).First(&client, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func (s storage) Delete(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).Delete(&Client{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("could not delete client: %w", err)
	}
	return nil
}
