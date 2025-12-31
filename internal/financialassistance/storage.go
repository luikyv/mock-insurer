package financialassistance

import (
	"context"
	"errors"
	"fmt"

	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/resource"
	"gorm.io/gorm"
)

type Storage interface {
	contracts(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*Contract], error)
	createConsentContract(ctx context.Context, c *ConsentContract) error
	consentContract(ctx context.Context, id, consentID, orgID string) (*ConsentContract, error)
	consentContracts(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*ConsentContract], error)
	movements(ctx context.Context, contractID, orgID string, pag page.Pagination) (page.Page[*Movement], error)
	transaction(ctx context.Context, fn func(Storage) error) error
}

type storage struct {
	db *gorm.DB
}

func (s storage) contracts(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*Contract], error) {
	query := s.db.WithContext(ctx).
		Where("org_id = ? OR cross_org = true", orgID).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC")
	contracts, err := page.Paginate[*Contract](query, pag)
	if err != nil {
		return page.Page[*Contract]{}, fmt.Errorf("failed to find contracts: %w", err)
	}
	return contracts, nil
}

func (s storage) createConsentContract(ctx context.Context, consentContract *ConsentContract) error {
	if err := s.db.WithContext(ctx).Create(consentContract).Error; err != nil {
		return fmt.Errorf("could not create consent contract: %w", err)
	}
	return nil
}

func (s storage) consentContract(ctx context.Context, contractID, consentID, orgID string) (*ConsentContract, error) {
	consentContract := &ConsentContract{}
	if err := s.db.WithContext(ctx).
		Preload("Contract").
		Where(`contract_id = ? AND consent_id = ? AND org_id = ?`, contractID, consentID, orgID).
		First(consentContract).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch consent contract: %w", err)
	}
	return consentContract, nil
}

func (s storage) consentContracts(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*ConsentContract], error) {
	query := s.db.WithContext(ctx).
		Model(&ConsentContract{}).
		Preload("Contract").
		Where(`consent_id = ? AND org_id = ?`, consentID, orgID).
		Where("status = ?", resource.StatusAvailable).
		Order("created_at DESC")

	consentContracts, err := page.Paginate[*ConsentContract](query, pag)
	if err != nil {
		return page.Page[*ConsentContract]{}, fmt.Errorf("failed to find consented contracts: %w", err)
	}

	return consentContracts, nil
}

func (s storage) movements(ctx context.Context, contractID, orgID string, pag page.Pagination) (page.Page[*Movement], error) {
	query := s.db.WithContext(ctx).
		Where("org_id = ? OR cross_org = true", orgID).
		Where("contract_id = ?", contractID).
		Order("created_at DESC")

	movements, err := page.Paginate[*Movement](query, pag)
	if err != nil {
		return page.Page[*Movement]{}, fmt.Errorf("failed to find movements: %w", err)
	}

	return movements, nil
}

func (s storage) transaction(ctx context.Context, fn func(Storage) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txStorage := storage{db: tx.WithContext(ctx)}
		return fn(txStorage)
	})
}
