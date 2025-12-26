package auto

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/quote"
	"github.com/luikyv/mock-insurer/internal/strutil"
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
	lead.Status = quote.StatusReceived
	lead.StatusUpdatedAt = timeutil.DateTimeNow()
	lead.CreatedAt = timeutil.DateTimeNow()
	lead.UpdatedAt = timeutil.DateTimeNow()
	return s.storage.createLead(ctx, lead)
}

func (s Service) CancelLead(ctx context.Context, consentID, orgID string, data quote.PatchData) (*Lead, error) {
	lead, err := s.Lead(ctx, consentID, orgID)
	if err != nil {
		return nil, err
	}
	return lead, s.updateLeadWithStatus(ctx, lead, quote.StatusCancelled)
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
	q.StatusUpdatedAt = timeutil.DateTimeNow()
	q.CreatedAt = timeutil.DateTimeNow()
	q.UpdatedAt = timeutil.DateTimeNow()
	if err := s.storage.create(ctx, q); err != nil {
		return err
	}

	go func() {
		run := func(ctx context.Context, q *Quote) error {
			switch q.Status {
			case quote.StatusReceived:
				if q.Data.TermStartDate.After(q.Data.TermEndDate) {
					return s.rejectQuote(ctx, q, "term start date is after term end date")
				}
				return s.updateQuoteWithStatus(ctx, q, quote.StatusEvaluated)
			case quote.StatusEvaluated:
				q.Data.Quotes = &[]Offer{
					{
						InsurerQuoteID:      uuid.New().String(),
						SusepProcessNumbers: []string{strutil.Random(50)},
						Premium: Premium{
							PaymentsQuantity: "1",
							TotalAmount: insurer.AmountDetails{
								Amount:   "100.00",
								UnitType: insurer.UnitTypeMonetary,
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.UnitDescriptionBRL,
								},
							},
							TotalNetAmount: insurer.AmountDetails{
								Amount:   "100.00",
								UnitType: insurer.UnitTypeMonetary,
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.UnitDescriptionBRL,
								},
							},
							IOF: insurer.AmountDetails{
								Amount:   "100.00",
								UnitType: insurer.UnitTypeMonetary,
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.UnitDescriptionBRL,
								},
							},
						},
						Coverages:   []OfferCoverage{},
						Assistances: []Assistance{},
					},
				}
				return s.updateQuoteWithStatus(ctx, q, quote.StatusAccepted)
			default:
				return nil
			}
		}

		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 1*time.Minute)
		defer cancel()

		ticker := time.NewTicker(15 * time.Second)
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

func (s Service) Update(ctx context.Context, consentID, orgID string, patchData quote.PatchData) (*Quote, error) {
	q, err := s.Quote(ctx, consentID, orgID)
	if err != nil {
		return nil, err
	}

	switch patchData.AuthorIdentificationType {
	case insurer.IdentificationTypeCPF:
		if q.Data.Customer.Personal == nil {
			return nil, errorutil.New("personal customer not found")
		}
		if q.Data.Customer.Personal.Identification == nil {
			return nil, errorutil.New("personal identification not found")
		}
		if patchData.AuthorIdentificationNumber != q.Data.Customer.Personal.Identification.CPF {
			return nil, errorutil.New("identification number mismatch")
		}
	case insurer.IdentificationTypeCNPJ:
		if q.Data.Customer.Business == nil {
			return nil, errorutil.New("business customer not found")
		}
		if q.Data.Customer.Business.Identification == nil {
			return nil, errorutil.New("business identification not found")
		}
		if patchData.AuthorIdentificationNumber != q.Data.Customer.Business.Identification.CompanyInfo.CNPJ {
			return nil, errorutil.New("identification number mismatch")
		}
	default:
		return nil, errorutil.New("invalid identification type")
	}

	if patchData.Status == quote.StatusAcknowledged {
		if q.Status != quote.StatusAccepted {
			return nil, errorutil.New("quote not accepted")
		}

		if patchData.InsurerQuoteID == nil {
			return nil, errorutil.New("insurer quote id not informed")
		}

		quoteIDs := make([]string, len(*q.Data.Quotes))
		for i, q := range *q.Data.Quotes {
			quoteIDs[i] = q.InsurerQuoteID
		}

		if !slices.Contains(quoteIDs, *patchData.InsurerQuoteID) {
			return nil, errorutil.New("insurer quote id not found")
		}

		q.Data.InsurerQuoteID = patchData.InsurerQuoteID
		now := timeutil.DateTimeNow()
		q.Data.ProtocolDateTime = &now
		protocolNumber := uuid.New().String()
		q.Data.ProtocolNumber = &protocolNumber
		redirectLink := "https://www.raidiam.com"
		q.Data.RedirectLink = &redirectLink
	}
	if patchData.Status == quote.StatusCancelled {
		if !slices.Contains([]quote.Status{quote.StatusReceived, quote.StatusEvaluated, quote.StatusAccepted}, q.Status) {
			return nil, errorutil.New("quote not in a status to be cancelled")
		}
	}

	return q, s.updateQuoteWithStatus(ctx, q, patchData.Status)
}

func (s Service) rejectQuote(ctx context.Context, q *Quote, reason string) error {
	q.Data.RejectionReason = &reason
	return s.updateQuoteWithStatus(ctx, q, quote.StatusRejected)
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
