package capitalizationtitle

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/resource"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

type Service struct {
	storage Storage
}

func NewService(db *gorm.DB) Service {
	return Service{storage: storage{db: db}}
}

func (s Service) Plans(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*Plan], error) {
	return s.storage.plans(ctx, ownerID, orgID, pag)
}

func (s Service) Authorize(ctx context.Context, ids []string, ownerID, consentID, orgID string) error {
	return s.transaction(ctx, func(txService Service) error {
		for _, id := range ids {
			if err := txService.storage.createConsentPlan(ctx, &ConsentPlan{
				ConsentID: uuid.MustParse(consentID),
				PlanID:    uuid.MustParse(id),
				OwnerID:   uuid.MustParse(ownerID),
				Status:    resource.StatusAvailable,
				OrgID:     orgID,
				CreatedAt: timeutil.DateTimeNow(),
				UpdatedAt: timeutil.DateTimeNow(),
			}); err != nil {
				return fmt.Errorf("could not create consent plan: %w", err)
			}
		}

		return nil
	})
}

func (s Service) ConsentedPlan(ctx context.Context, id, consentID, orgID string) (*Plan, error) {
	consentPlan, err := s.storage.consentPlan(ctx, id, consentID, orgID)
	if err != nil {
		return nil, err
	}

	if consentPlan.Status != resource.StatusAvailable {
		return nil, ErrNotAvailable
	}

	return consentPlan.Plan, nil
}

func (s Service) ConsentedPlans(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*Plan], error) {
	consentPlans, err := s.storage.consentPlans(ctx, consentID, orgID, pag)
	if err != nil {
		return page.Page[*Plan]{}, err
	}

	var plans []*Plan
	for _, consentPlan := range consentPlans.Records {
		plans = append(plans, consentPlan.Plan)
	}
	return page.New(plans, pag, consentPlans.TotalRecords), nil
}

func (s Service) ConsentedEvents(ctx context.Context, planID, consentID, orgID string, pag page.Pagination) (page.Page[*Event], error) {
	if _, err := s.ConsentedPlan(ctx, planID, consentID, orgID); err != nil {
		return page.Page[*Event]{}, err
	}

	return s.storage.events(ctx, planID, orgID, pag)
}

func (s Service) ConsentedSettlements(ctx context.Context, planID, consentID, orgID string, pag page.Pagination) (page.Page[*Settlement], error) {
	if _, err := s.ConsentedPlan(ctx, planID, consentID, orgID); err != nil {
		return page.Page[*Settlement]{}, err
	}

	return s.storage.settlements(ctx, planID, orgID, pag)
}

func (s Service) transaction(ctx context.Context, fn func(Service) error) error {
	return s.storage.transaction(ctx, func(txStorage Storage) error {
		return fn(Service{storage: txStorage})
	})
}
