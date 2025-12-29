package capitalizationtitle

import (
	"context"
	"errors"
	"fmt"

	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/resource"
	"gorm.io/gorm"
)

type Storage interface {
	plans(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*Plan], error)
	createConsentPlan(ctx context.Context, c *ConsentPlan) error
	consentPlan(ctx context.Context, id, consentID, orgID string) (*ConsentPlan, error)
	consentPlans(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*ConsentPlan], error)
	events(ctx context.Context, planID, orgID string, pag page.Pagination) (page.Page[*Event], error)
	settlements(ctx context.Context, planID, orgID string, pag page.Pagination) (page.Page[*Settlement], error)
	transaction(ctx context.Context, fn func(Storage) error) error
}

type storage struct {
	db *gorm.DB
}

func (s storage) plans(ctx context.Context, ownerID, orgID string, pag page.Pagination) (page.Page[*Plan], error) {
	query := s.db.WithContext(ctx).Where("org_id = ? OR cross_org = true", orgID).Order("created_at DESC")
	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}

	plans, err := page.Paginate[*Plan](query, pag)
	if err != nil {
		return page.Page[*Plan]{}, fmt.Errorf("failed to find plans: %w", err)
	}
	return plans, nil
}

func (s storage) createConsentPlan(ctx context.Context, c *ConsentPlan) error {
	if err := s.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("could not create consent plan: %w", err)
	}
	return nil
}

func (s storage) consentPlan(ctx context.Context, id, consentID, orgID string) (*ConsentPlan, error) {
	consentPlan := &ConsentPlan{}
	if err := s.db.WithContext(ctx).
		Preload("Plan").
		Where(`plan_id = ? AND consent_id = ? AND org_id = ?`, id, consentID, orgID).
		First(consentPlan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch consent plan: %w", err)
	}
	return consentPlan, nil
}

func (s storage) consentPlans(ctx context.Context, consentID, orgID string, pag page.Pagination) (page.Page[*ConsentPlan], error) {
	query := s.db.WithContext(ctx).
		Preload("Plan").
		Where("consent_id = ? AND org_id = ?", consentID, orgID).
		Where("status = ?", resource.StatusAvailable).
		Order("created_at DESC")

	consentPlans, err := page.Paginate[*ConsentPlan](query, pag)
	if err != nil {
		return page.Page[*ConsentPlan]{}, fmt.Errorf("failed to find consented plans: %w", err)
	}
	return consentPlans, nil
}

func (s storage) events(ctx context.Context, planID, orgID string, pag page.Pagination) (page.Page[*Event], error) {
	query := s.db.WithContext(ctx).
		Where("org_id = ? OR cross_org = true", orgID).
		Where("plan_id = ?", planID).
		Order("created_at DESC")

	events, err := page.Paginate[*Event](query, pag)
	if err != nil {
		return page.Page[*Event]{}, fmt.Errorf("failed to find events: %w", err)
	}

	return events, nil
}

func (s storage) settlements(ctx context.Context, planID, orgID string, pag page.Pagination) (page.Page[*Settlement], error) {
	query := s.db.WithContext(ctx).
		Where("org_id = ? OR cross_org = true", orgID).
		Where("plan_id = ?", planID).
		Order("created_at DESC")

	settlements, err := page.Paginate[*Settlement](query, pag)
	if err != nil {
		return page.Page[*Settlement]{}, fmt.Errorf("failed to find settlements: %w", err)
	}

	return settlements, nil
}

func (s storage) transaction(ctx context.Context, fn func(Storage) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txStorage := storage{db: tx.WithContext(ctx)}
		return fn(txStorage)
	})
}
