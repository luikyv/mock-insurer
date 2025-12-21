package auto

import (
	"context"
	"log/slog"
	"time"

	"github.com/luikyv/mock-insurer/internal/quote"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

type Service struct {
	storage Storage
}

func NewService(db *gorm.DB) Service {
	return Service{storage: storage{db: db}}
}

func (s Service) CreateLead(ctx context.Context, lead *Lead) error {
	return s.storage.createLead(ctx, lead)
}

func (s Service) UpdateLead(ctx context.Context, consentID, orgID string, data quote.PatchData) (*Lead, error) {
	lead, err := s.Lead(ctx, consentID, orgID)
	if err != nil {
		return nil, err
	}
	return lead, s.updateLeadWithStatus(ctx, lead, data.Status)
}

func (s Service) Lead(ctx context.Context, consentID, orgID string) (*Lead, error) {
	return s.storage.lead(ctx, LeadQuery{ConsentID: consentID}, orgID)
}

func (s Service) updateLeadWithStatus(ctx context.Context, lead *Lead, status quote.Status) error {
	lead.Status = status
	lead.StatusUpdatedAt = timeutil.DateTimeNow()
	return s.updateLead(ctx, lead)
}

func (s Service) updateLead(ctx context.Context, lead *Lead) error {
	lead.UpdatedAt = timeutil.DateTimeNow()
	return s.storage.updateLead(ctx, lead)
}

func (s Service) CreateQuote(ctx context.Context, q *Quote) error {
	q.Status = quote.StatusReceived
	q.CreatedAt = timeutil.DateTimeNow()
	q.UpdatedAt = timeutil.DateTimeNow()
	if err := s.storage.create(ctx, q); err != nil {
		return err
	}

	go func() {
		run := func(ctx context.Context, q *Quote) error {
			switch q.Status {
			case quote.StatusReceived:
				return s.updateQuoteWithStatus(ctx, q, quote.StatusEvaluated)
			default:
				return nil
			}
		}

		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
		defer cancel()

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				slog.DebugContext(ctx, "evaluating quote automations")
				if err := run(ctx, q); err != nil {
					slog.ErrorContext(ctx, "error running quote automations for quote", "quote_id", q.ID, "error", err)
					return
				}
			case <-ctx.Done():
				slog.DebugContext(ctx, "quote automation deadline reached, stopping ticker")
				return
			}
		}
	}()

	return nil
}

func (s Service) Quote(ctx context.Context, consentID, orgID string) (*Quote, error) {
	return s.storage.quote(ctx, Query{ConsentID: consentID}, orgID)
}

func (s Service) Update(ctx context.Context, consentID, orgID string, data quote.PatchData) error {
	q, err := s.Quote(ctx, consentID, orgID)
	if err != nil {
		return err
	}

	return s.updateQuoteWithStatus(ctx, q, data.Status)
}

func (s Service) updateQuoteWithStatus(ctx context.Context, q *Quote, status quote.Status) error {
	q.Status = status
	q.StatusUpdatedAt = timeutil.DateTimeNow()
	return s.updateQuote(ctx, q)
}

func (s Service) updateQuote(ctx context.Context, quote *Quote) error {
	quote.UpdatedAt = timeutil.DateTimeNow()
	return s.storage.update(ctx, quote)
}
