package quote

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type StorageLead[L Lead] interface {
	create(context.Context, L) error
	update(context.Context, L) error
	lead(ctx context.Context, query LeadQuery, orgID string) (L, error)
}

type storageLead[L Lead] struct {
	db *gorm.DB
}

//nolint:unused
func (s storageLead[L]) create(ctx context.Context, lead L) error {
	if err := s.db.WithContext(ctx).Create(lead).Error; err != nil {
		return fmt.Errorf("could not create lead: %w", err)
	}
	return nil
}

//nolint:unused
func (s storageLead[L]) update(ctx context.Context, lead L) error {
	err := s.db.WithContext(ctx).
		Model(new(L)).
		Omit("CreatedAt").
		Where("id = ? AND org_id = ?", lead.GetID(), lead.GetOrgID()).
		Updates(lead).Error
	if err != nil {
		return fmt.Errorf("could not update lead: %w", err)
	}

	return nil
}

//nolint:unused
func (s storageLead[L]) lead(ctx context.Context, opts LeadQuery, orgID string) (L, error) {
	var zero L
	query := s.db.WithContext(ctx).Model(new(L)).Where("org_id = ?", orgID)
	if opts.ID != "" {
		query = query.Where("id = ?", opts.ID)
	}

	if opts.ConsentID != "" {
		query = query.Where("consent_id = ?", opts.ConsentID)
	}

	lead := new(L)
	if err := query.First(lead).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return zero, ErrNotFound
		}
		return zero, fmt.Errorf("could not fetch lead: %w", err)
	}
	return *lead, nil
}

type Storage[Q Quote] interface {
	create(context.Context, Q) error
	update(context.Context, Q) error
	quote(ctx context.Context, query Query, orgID string) (Q, error)
}

type storage[Q Quote] struct {
	db *gorm.DB
}

//nolint:unused
func (s storage[Q]) create(ctx context.Context, q Q) error {
	if err := s.db.WithContext(ctx).Create(q).Error; err != nil {
		return fmt.Errorf("could not create quote: %w", err)
	}
	return nil
}

//nolint:unused
func (s storage[Q]) update(ctx context.Context, q Q) error {
	err := s.db.WithContext(ctx).
		Model(new(Q)).
		Omit("CreatedAt").
		Where("id = ? AND org_id = ?", q.GetID(), q.GetOrgID()).
		Updates(q).Error
	if err != nil {
		return fmt.Errorf("could not update quote: %w", err)
	}
	return nil
}

//nolint:unused
func (s storage[Q]) quote(ctx context.Context, opts Query, orgID string) (Q, error) {
	var zero Q
	query := s.db.WithContext(ctx).Model(new(Q)).Where("org_id = ?", orgID)
	if opts.ID != "" {
		query = query.Where("id = ?", opts.ID)
	}

	if opts.ConsentID != "" {
		query = query.Where("consent_id = ?", opts.ConsentID)
	}

	quote := new(Q)
	if err := query.First(quote).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return zero, ErrNotFound
		}
		return zero, fmt.Errorf("could not fetch quote: %w", err)
	}
	return *quote, nil
}
