package financialassistance

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

func (s Service) Contracts(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*Contract], error) {
	return s.storage.contracts(ctx, ownerID, orgID, pag)
}

func (s Service) Authorize(ctx context.Context, ids []string, ownerID, consentID, orgID string) error {
	return s.transaction(ctx, func(txService Service) error {
		for _, id := range ids {
			if err := txService.storage.createConsentContract(ctx, &ConsentContract{
				ConsentID:  uuid.MustParse(consentID),
				ContractID: id,
				OwnerID:    uuid.MustParse(ownerID),
				Status:     resource.StatusAvailable,
				OrgID:      orgID,
				CreatedAt:  timeutil.DateTimeNow(),
				UpdatedAt:  timeutil.DateTimeNow(),
			}); err != nil {
				return fmt.Errorf("could not create consent contract: %w", err)
			}
		}

		return nil
	})
}

func (s Service) ConsentedContract(ctx context.Context, id, consentID, orgID string) (*Contract, error) {
	consentContract, err := s.storage.consentContract(ctx, id, consentID, orgID)
	if err != nil {
		return nil, err
	}

	if consentContract.Status != resource.StatusAvailable {
		return nil, ErrNotAvailable
	}

	return consentContract.Contract, nil
}

func (s Service) ConsentedContracts(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*Contract], error) {
	consentContracts, err := s.storage.consentContracts(ctx, consentID, orgID, pag)
	if err != nil {
		return page.Page[*Contract]{}, err
	}

	var contracts []*Contract
	for _, consentContract := range consentContracts.Records {
		contracts = append(contracts, consentContract.Contract)
	}
	return page.New(contracts, pag, consentContracts.TotalRecords), nil
}

func (s Service) ConsentedMovements(ctx context.Context, contractID, consentID, orgID string, pag page.Pagination) (page.Page[*Movement], error) {
	if _, err := s.ConsentedContract(ctx, contractID, consentID, orgID); err != nil {
		return page.Page[*Movement]{}, err
	}

	return s.storage.movements(ctx, contractID, orgID, pag)
}

func (s Service) transaction(ctx context.Context, fn func(Service) error) error {
	return s.storage.transaction(ctx, func(txStorage Storage) error {
		return fn(Service{storage: txStorage})
	})
}
