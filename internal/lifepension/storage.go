package lifepension

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
	portabilities(ctx context.Context, contractID, orgID string, pag page.Pagination) (page.Page[*Portability], error)
	withdrawals(ctx context.Context, contractID, orgID string, pag page.Pagination) (page.Page[*Withdrawal], error)
	claims(ctx context.Context, contractID, orgID string, pag page.Pagination) (page.Page[*Claim], error)
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

func (s storage) portabilities(ctx context.Context, contractID, orgID string, pag page.Pagination) (page.Page[*Portability], error) {
	query := s.db.WithContext(ctx).
		Where("org_id = ? OR cross_org = true", orgID).
		Where("contract_id = ?", contractID).
		Order("created_at DESC")

	portabilities, err := page.Paginate[*Portability](query, pag)
	if err != nil {
		return page.Page[*Portability]{}, fmt.Errorf("failed to find portabilities: %w", err)
	}

	return portabilities, nil
}

func (s storage) withdrawals(ctx context.Context, contractID, orgID string, pag page.Pagination) (page.Page[*Withdrawal], error) {
	query := s.db.WithContext(ctx).
		Where("org_id = ? OR cross_org = true", orgID).
		Where("contract_id = ?", contractID).
		Order("created_at DESC")

	withdrawals, err := page.Paginate[*Withdrawal](query, pag)
	if err != nil {
		return page.Page[*Withdrawal]{}, fmt.Errorf("failed to find withdrawals: %w", err)
	}

	return withdrawals, nil
}

func (s storage) claims(ctx context.Context, contractID, orgID string, pag page.Pagination) (page.Page[*Claim], error) {
	query := s.db.WithContext(ctx).
		Where("org_id = ? OR cross_org = true", orgID).
		Where("contract_id = ?", contractID).
		Order("created_at DESC")

	claims, err := page.Paginate[*Claim](query, pag)
	if err != nil {
		return page.Page[*Claim]{}, fmt.Errorf("failed to find claims: %w", err)
	}

	return claims, nil
}

func (s storage) transaction(ctx context.Context, fn func(Storage) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txStorage := storage{db: tx.WithContext(ctx)}
		return fn(txStorage)
	})
}
