package client

import (
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

type Client struct {
	ID          string       `gorm:"primaryKey"`
	Data        goidc.Client `gorm:"serializer:json"`
	WebhookURIs []string     `gorm:"serializer:json"`
	Name        string

	OrgID     string
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (Client) TableName() string {
	return "oauth_clients"
}
