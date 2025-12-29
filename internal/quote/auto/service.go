package auto

import (
	"context"

	"github.com/luikyv/mock-insurer/internal/quote"
	"gorm.io/gorm"
)

type Service struct {
	serviceLead quote.ServiceLead[*Lead]
	service     quote.Service[*Quote]
}

func NewService(db *gorm.DB) Service {
	return Service{
		serviceLead: quote.NewServiceLead[*Lead](db),
		service:     quote.NewService[*Quote](db),
	}
}

func (s Service) CreateLead(ctx context.Context, lead *Lead) error {
	return s.serviceLead.CreateLead(ctx, lead)
}

func (s Service) CancelLead(ctx context.Context, consentID, orgID string, data quote.PatchData) (*Lead, error) {
	return s.serviceLead.CancelLead(ctx, consentID, orgID, data)
}

func (s Service) CreateQuote(ctx context.Context, q *Quote) error {
	return s.service.CreateQuote(ctx, q)
}

func (s Service) Quote(ctx context.Context, consentID, orgID string) (*Quote, error) {
	return s.service.Quote(ctx, consentID, orgID)
}

func (s Service) Update(ctx context.Context, consentID, orgID string, patchData quote.PatchData) (*Quote, error) {
	return s.service.Update(ctx, consentID, orgID, patchData)
}
