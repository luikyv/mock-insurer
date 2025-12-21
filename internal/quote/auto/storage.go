package auto

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Storage interface {
	createLead(ctx context.Context, lead *Lead) error
	updateLead(ctx context.Context, lead *Lead) error
	lead(ctx context.Context, query LeadQuery, orgID string) (*Lead, error)
	create(ctx context.Context, quote *Quote) error
	update(ctx context.Context, quote *Quote) error
	quote(ctx context.Context, query Query, orgID string) (*Quote, error)
}

type storage struct {
	db *gorm.DB
}

func (s storage) createLead(ctx context.Context, lead *Lead) error {
	if err := s.db.WithContext(ctx).Create(lead).Error; err != nil {
		return fmt.Errorf("could not create lead: %w", err)
	}
	return nil
}

func (s storage) updateLead(ctx context.Context, lead *Lead) error {
	err := s.db.WithContext(ctx).
		Model(&Lead{}).
		Omit("CreatedAt").
		Where("id = ? AND org_id = ?", lead.ID, lead.OrgID).
		Updates(lead).Error
	if err != nil {
		return fmt.Errorf("could not update lead: %w", err)
	}

	return nil
}

func (s storage) lead(ctx context.Context, opts LeadQuery, orgID string) (*Lead, error) {
	query := s.db.WithContext(ctx).Model(&Lead{}).Where("org_id = ?", orgID)
	if opts.ID != "" {
		query = query.Where("id = ?", opts.ID)
	}

	if opts.ConsentID != "" {
		query = query.Where("consent_id = ?", opts.ConsentID)
	}

	lead := &Lead{}
	if err := query.First(lead).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLeadNotFound
		}
		return nil, fmt.Errorf("could not fetch lead: %w", err)
	}
	return lead, nil
}

func (s storage) create(ctx context.Context, quote *Quote) error {
	if err := s.db.WithContext(ctx).Create(quote).Error; err != nil {
		return fmt.Errorf("could not create quote: %w", err)
	}
	return nil
}

func (s storage) update(ctx context.Context, quote *Quote) error {
	err := s.db.WithContext(ctx).
		Model(&Quote{}).
		Omit("CreatedAt").
		Where("id = ? AND org_id = ?", quote.ID, quote.OrgID).
		Updates(quote).Error
	if err != nil {
		return fmt.Errorf("could not update quote: %w", err)
	}
	return nil
}

func (s storage) quote(ctx context.Context, opts Query, orgID string) (*Quote, error) {
	query := s.db.WithContext(ctx).Model(&Quote{}).Where("org_id = ?", orgID)
	if opts.ID != "" {
		query = query.Where("id = ?", opts.ID)
	}

	if opts.ConsentID != "" {
		query = query.Where("consent_id = ?", opts.ConsentID)
	}

	quote := &Quote{}
	if err := query.First(quote).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch quote: %w", err)
	}
	return quote, nil
}
