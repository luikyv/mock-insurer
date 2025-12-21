package resource

import (
	"context"
	"fmt"

	"github.com/luikyv/mock-insurer/internal/page"
	"gorm.io/gorm"
)

type Storage interface {
	Resources(ctx context.Context, orgID string, filter *Filter, pag page.Pagination) (page.Page[*Resource], error)
}

type storage struct {
	db *gorm.DB
}

func (s storage) Resources(ctx context.Context, orgID string, filter *Filter, pag page.Pagination) (page.Page[*Resource], error) {
	query := s.db.WithContext(ctx).Where("org_id = ?", orgID).Order("created_at DESC")
	if filter == nil {
		filter = &Filter{}
	}
	if filter.OwnerID != "" {
		query = query.Where("owner_id = ?", filter.OwnerID)
	}
	if filter.ConsentID != "" {
		query = query.Where("consent_id = ?", filter.ConsentID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	rs, err := page.Paginate[*Resource](query, pag)
	if err != nil {
		return page.Page[*Resource]{}, fmt.Errorf("could not find consented resources: %w", err)
	}

	return rs, nil
}
