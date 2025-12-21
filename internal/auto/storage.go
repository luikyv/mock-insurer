package auto

import (
	"context"
	"errors"
	"fmt"

	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/resource"
	"gorm.io/gorm"
)

type Storage interface {
	policies(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*Policy], error)
	createConsentPolicy(ctx context.Context, c *ConsentPolicy) error
	consentPolicy(ctx context.Context, id, consentID, orgID string) (*ConsentPolicy, error)
	consentPolicies(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*ConsentPolicy], error)
	claims(ctx context.Context, policyID, orgID string, pag page.Pagination) (page.Page[*Claim], error)
	transaction(ctx context.Context, fn func(Storage) error) error
}

type storage struct {
	db *gorm.DB
}

func (s storage) policies(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*Policy], error) {
	query := s.db.WithContext(ctx).Where("org_id = ? OR cross_org = true", orgID).Order("created_at DESC")
	if opts.OwnerID != "" {
		query = query.Where("owner_id = ?", opts.OwnerID)
	}

	policies, err := page.Paginate[*Policy](query, pag)
	if err != nil {
		return page.Page[*Policy]{}, err
	}
	return policies, nil
}

func (s storage) createConsentPolicy(ctx context.Context, consentPolicy *ConsentPolicy) error {
	if err := s.db.WithContext(ctx).Create(consentPolicy).Error; err != nil {
		return fmt.Errorf("could not create consent policy: %w", err)
	}
	return nil
}

func (s storage) consentPolicy(ctx context.Context, policyID, consentID, orgID string) (*ConsentPolicy, error) {
	consentPolicy := &ConsentPolicy{}
	if err := s.db.WithContext(ctx).
		Preload("Policy").
		Where(`policy_id = ? AND consent_id = ? AND org_id = ?`, policyID, consentID, orgID).
		First(consentPolicy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch consent account: %w", err)
	}
	return consentPolicy, nil
}

func (s storage) consentPolicies(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*ConsentPolicy], error) {
	query := s.db.WithContext(ctx).
		Model(&ConsentPolicy{}).
		Preload("Policy").
		Where(`consent_id = ? AND org_id = ?`, consentID, orgID).
		Where("status = ?", resource.StatusAvailable).
		Order("created_at DESC")

	consentPolicies, err := page.Paginate[*ConsentPolicy](query, pag)
	if err != nil {
		return page.Page[*ConsentPolicy]{}, fmt.Errorf("failed to find consented policies: %w", err)
	}

	return consentPolicies, nil
}

func (s storage) claims(ctx context.Context, policyID, orgID string, pag page.Pagination) (page.Page[*Claim], error) {
	query := s.db.WithContext(ctx).
		Where("org_id = ? OR cross_org = true", orgID).
		Where("policy_id = ?", policyID).
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
