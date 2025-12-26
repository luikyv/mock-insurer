package customer

import (
	"context"

	"github.com/luikyv/mock-insurer/internal/page"
	"gorm.io/gorm"
)

type Service struct {
	storage Storage
}

func NewService(db *gorm.DB) Service {
	return Service{storage: storage{db: db}}
}

func (s Service) PersonalIdentifications(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*PersonalIdentification], error) {
	return s.storage.personalIdentifications(ctx, orgID, &Filter{OwnerID: ownerID}, pag)
}

func (s Service) PersonalQualifications(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*PersonalQualification], error) {
	return s.storage.personalQualifications(ctx, orgID, &Filter{OwnerID: ownerID}, pag)
}

func (s Service) PersonalComplimentaryInformations(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*PersonalComplimentaryInformation], error) {
	return s.storage.personalComplimentaryInformations(ctx, orgID, &Filter{OwnerID: ownerID}, pag)
}

func (s Service) BusinessIdentifications(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*BusinessIdentification], error) {
	return s.storage.businessIdentifications(ctx, orgID, &Filter{OwnerID: ownerID}, pag)
}

func (s Service) BusinessQualifications(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*BusinessQualification], error) {
	return s.storage.businessQualifications(ctx, orgID, &Filter{OwnerID: ownerID}, pag)
}

func (s Service) BusinessComplimentaryInformations(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*BusinessComplimentaryInformation], error) {
	return s.storage.businessComplimentaryInformations(ctx, orgID, &Filter{OwnerID: ownerID}, pag)
}
