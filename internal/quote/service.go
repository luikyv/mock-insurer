package quote

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

type ServiceLead[L Lead] struct {
	storage StorageLead[L]
}

func NewServiceLead[L Lead](db *gorm.DB) ServiceLead[L] {
	return ServiceLead[L]{storage: storageLead[L]{db: db}}
}

func (s ServiceLead[L]) CreateLead(ctx context.Context, lead L) error {
	lead.SetStatus(StatusReceived)
	lead.SetStatusUpdatedAt(timeutil.DateTimeNow())
	lead.SetCreatedAt(timeutil.DateTimeNow())
	lead.SetUpdatedAt(timeutil.DateTimeNow())
	return s.storage.create(ctx, lead)
}

func (s ServiceLead[L]) CancelLead(ctx context.Context, consentID, orgID string, data PatchData) (L, error) {
	var zero L
	lead, err := s.lead(ctx, consentID, orgID)
	if err != nil {
		return zero, err
	}
	return lead, s.updateLeadWithStatus(ctx, lead, StatusCancelled)
}

func (s ServiceLead[L]) lead(ctx context.Context, consentID, orgID string) (L, error) {
	return s.storage.lead(ctx, LeadQuery{ConsentID: consentID}, orgID)
}

func (s ServiceLead[L]) updateLeadWithStatus(ctx context.Context, lead L, status Status) error {
	lead.SetStatus(status)
	lead.SetStatusUpdatedAt(timeutil.DateTimeNow())
	return s.storage.update(ctx, lead)
}

type Service[Q Quote] struct {
	storage Storage[Q]
}

func NewService[Q Quote](db *gorm.DB) Service[Q] {
	return Service[Q]{storage: storage[Q]{db: db}}
}

func (s Service[Q]) CreateQuote(ctx context.Context, q Q) error {
	q.SetStatus(StatusReceived)
	q.SetStatusUpdatedAt(timeutil.DateTimeNow())
	q.SetCreatedAt(timeutil.DateTimeNow())
	q.SetUpdatedAt(timeutil.DateTimeNow())
	if err := s.storage.create(ctx, q); err != nil {
		return err
	}

	go func() {
		run := func(ctx context.Context, q Q) error {
			switch q.GetStatus() {
			case StatusReceived:
				if q.GetTermStartDate().After(q.GetTermEndDate()) {
					return s.rejectQuote(ctx, q, "term start date is after term end date")
				}
				return s.updateQuoteWithStatus(ctx, q, StatusEvaluated)
			case StatusEvaluated:
				q.CreateOffers()
				return s.updateQuoteWithStatus(ctx, q, StatusAccepted)
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
					slog.ErrorContext(ctx, "error running quote automations for quote", "quote_id", q.GetID(), "error", err)
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

func (s Service[Q]) Quote(ctx context.Context, consentID, orgID string) (Q, error) {
	return s.storage.quote(ctx, Query{ConsentID: consentID}, orgID)
}

func (s Service[Q]) Update(ctx context.Context, consentID, orgID string, patchData PatchData) (Q, error) {
	var zero Q
	q, err := s.Quote(ctx, consentID, orgID)
	if err != nil {
		return zero, err
	}

	switch patchData.AuthorIdentificationType {
	case insurer.IdentificationTypeCPF:
		personalIdentification := q.GetPersonalIdentification()
		if personalIdentification == nil {
			return zero, errorutil.New("personal identification not defined")
		}
		if *personalIdentification != patchData.AuthorIdentificationNumber {
			return zero, errorutil.New("identification number mismatch")
		}
	case insurer.IdentificationTypeCNPJ:
		businessIdentification := q.GetBusinessIdentification()
		if businessIdentification == nil {
			return zero, errorutil.New("business identification not defined")
		}
		if *businessIdentification != patchData.AuthorIdentificationNumber {
			return zero, errorutil.New("business identification number mismatch")
		}
	default:
		return zero, errorutil.New("invalid identification type")
	}

	if patchData.Status == StatusAcknowledged {
		if q.GetStatus() != StatusAccepted {
			return zero, errorutil.New("quote not accepted")
		}

		if patchData.InsurerQuoteID == nil {
			return zero, errorutil.New("insurer quote id not informed")
		}

		if !slices.Contains(q.GetOfferIDs(), *patchData.InsurerQuoteID) {
			return zero, errorutil.New("insurer quote id not found")
		}

		q.SetInsurerQuoteID(*patchData.InsurerQuoteID)
		q.SetProtocolDateTime(timeutil.DateTimeNow())
		q.SetProtocolNumber(uuid.New().String())
		q.SetRedirectLink("https://www.raidiam.com")
	}
	if patchData.Status == StatusCancelled {
		if !slices.Contains([]Status{StatusReceived, StatusEvaluated, StatusAccepted}, q.GetStatus()) {
			return zero, errorutil.New("quote not in a status to be cancelled")
		}
	}

	return q, s.updateQuoteWithStatus(ctx, q, patchData.Status)
}

func (s Service[Q]) rejectQuote(ctx context.Context, q Q, reason string) error {
	q.SetRejectionReason(reason)
	return s.updateQuoteWithStatus(ctx, q, StatusRejected)
}

func (s Service[Q]) updateQuoteWithStatus(ctx context.Context, q Q, status Status) error {
	q.SetStatus(status)
	q.SetStatusUpdatedAt(timeutil.DateTimeNow())
	return s.updateQuote(ctx, q)
}

func (s Service[Q]) updateQuote(ctx context.Context, q Q) error {
	q.SetUpdatedAt(timeutil.DateTimeNow())
	return s.storage.update(ctx, q)
}
