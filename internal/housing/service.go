package housing

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

func (s Service) Policies(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*Policy], error) {
	return s.storage.policies(ctx, ownerID, orgID, pag)
}

func (s Service) Authorize(ctx context.Context, ids []string, ownerID, consentID, orgID string) error {
	return s.transaction(ctx, func(txService Service) error {
		for _, id := range ids {
			if err := txService.storage.createConsentPolicy(ctx, &ConsentPolicy{
				ConsentID: uuid.MustParse(consentID),
				PolicyID:  id,
				OwnerID:   uuid.MustParse(ownerID),
				Status:    resource.StatusAvailable,
				OrgID:     orgID,
				CreatedAt: timeutil.DateTimeNow(),
				UpdatedAt: timeutil.DateTimeNow(),
			}); err != nil {
				return fmt.Errorf("could not create consent policy: %w", err)
			}
		}

		return nil
	})
}

func (s Service) ConsentedPolicy(ctx context.Context, id, consentID, orgID string) (*Policy, error) {
	consentPolicy, err := s.storage.consentPolicy(ctx, id, consentID, orgID)
	if err != nil {
		return nil, err
	}

	if consentPolicy.Status != resource.StatusAvailable {
		return nil, ErrNotAvailable
	}

	return consentPolicy.Policy, nil
}

func (s Service) ConsentedPolicies(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*Policy], error) {
	consentPolicies, err := s.storage.consentPolicies(ctx, consentID, orgID, pag)
	if err != nil {
		return page.Page[*Policy]{}, err
	}

	var policies []*Policy
	for _, consentPolicy := range consentPolicies.Records {
		policies = append(policies, consentPolicy.Policy)
	}
	return page.New(policies, pag, consentPolicies.TotalRecords), nil
}

func (s Service) ConsentedClaims(ctx context.Context, policyID, consentID, orgID string, pag page.Pagination) (page.Page[*Claim], error) {
	if _, err := s.ConsentedPolicy(ctx, policyID, consentID, orgID); err != nil {
		return page.Page[*Claim]{}, err
	}

	return s.storage.claims(ctx, policyID, orgID, pag)
}

func (s Service) transaction(ctx context.Context, fn func(Service) error) error {
	return s.storage.transaction(ctx, func(txStorage Storage) error {
		return fn(Service{storage: txStorage})
	})
}

