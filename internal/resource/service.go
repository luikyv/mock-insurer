package resource

import (
	"context"

	"github.com/luikyv/mock-insurer/internal/page"
	"gorm.io/gorm"
)

type Service struct {
	storage Storage
}

func NewService(db *gorm.DB) Service {
	return Service{
		storage: storage{db: db},
	}
}

func (s Service) Resources(ctx context.Context, orgID string, filter *Filter, pag page.Pagination) (page.Page[*Resource], error) {
	return s.storage.Resources(ctx, orgID, filter, pag)
}
