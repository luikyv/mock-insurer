package customer

import (
	"context"
	"fmt"

	"github.com/luikyv/mock-insurer/internal/page"
	"gorm.io/gorm"
)

type Storage interface {
	personalIdentifications(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*PersonalIdentification], error)
	personalQualifications(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*PersonalQualification], error)
	personalComplimentaryInformations(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*PersonalComplimentaryInformation], error)
	businessIdentifications(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*BusinessIdentification], error)
	businessQualifications(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*BusinessQualification], error)
	businessComplimentaryInformations(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*BusinessComplimentaryInformation], error)
}

type storage struct {
	db *gorm.DB
}

func (s storage) personalIdentifications(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*PersonalIdentification], error) {
	query := s.db.WithContext(ctx).Where("org_id = ?", orgID).Order("created_at DESC")
	if opts == nil {
		opts = &Filter{}
	}
	if opts.OwnerID != "" {
		query = query.Where("owner_id = ?", opts.OwnerID)
	}

	personalIdentifications, err := page.Paginate[*PersonalIdentification](query, pag)
	if err != nil {
		return page.Page[*PersonalIdentification]{}, fmt.Errorf("could not find personal identifications: %w", err)
	}
	return personalIdentifications, nil
}

func (s storage) personalQualifications(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*PersonalQualification], error) {
	query := s.db.WithContext(ctx).Where("org_id = ?", orgID).Order("created_at DESC")
	if opts == nil {
		opts = &Filter{}
	}
	if opts.OwnerID != "" {
		query = query.Where("owner_id = ?", opts.OwnerID)
	}

	personalQualifications, err := page.Paginate[*PersonalQualification](query, pag)
	if err != nil {
		return page.Page[*PersonalQualification]{}, fmt.Errorf("could not find personal qualifications: %w", err)
	}
	return personalQualifications, nil
}

func (s storage) personalComplimentaryInformations(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*PersonalComplimentaryInformation], error) {
	query := s.db.WithContext(ctx).Where("org_id = ?", orgID).Order("created_at DESC")
	if opts == nil {
		opts = &Filter{}
	}
	if opts.OwnerID != "" {
		query = query.Where("owner_id = ?", opts.OwnerID)
	}

	personalComplimentaryInformations, err := page.Paginate[*PersonalComplimentaryInformation](query, pag)
	if err != nil {
		return page.Page[*PersonalComplimentaryInformation]{}, fmt.Errorf("could not find personal complimentary informations: %w", err)
	}
	return personalComplimentaryInformations, nil
}

func (s storage) businessIdentifications(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*BusinessIdentification], error) {
	query := s.db.WithContext(ctx).Where("org_id = ?", orgID).Order("created_at DESC")
	if opts == nil {
		opts = &Filter{}
	}
	if opts.OwnerID != "" {
		query = query.Where("owner_id = ?", opts.OwnerID)
	}

	businessIdentifications, err := page.Paginate[*BusinessIdentification](query, pag)
	if err != nil {
		return page.Page[*BusinessIdentification]{}, fmt.Errorf("could not find business identifications: %w", err)
	}
	return businessIdentifications, nil
}

func (s storage) businessQualifications(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*BusinessQualification], error) {
	query := s.db.WithContext(ctx).Where("org_id = ?", orgID).Order("created_at DESC")
	if opts == nil {
		opts = &Filter{}
	}
	if opts.OwnerID != "" {
		query = query.Where("owner_id = ?", opts.OwnerID)
	}

	businessQualifications, err := page.Paginate[*BusinessQualification](query, pag)
	if err != nil {
		return page.Page[*BusinessQualification]{}, fmt.Errorf("could not find business qualifications: %w", err)
	}
	return businessQualifications, nil
}

func (s storage) businessComplimentaryInformations(ctx context.Context, orgID string, opts *Filter, pag page.Pagination) (page.Page[*BusinessComplimentaryInformation], error) {
	query := s.db.WithContext(ctx).Where("org_id = ?", orgID).Order("created_at DESC")
	if opts == nil {
		opts = &Filter{}
	}
	if opts.OwnerID != "" {
		query = query.Where("owner_id = ?", opts.OwnerID)
	}

	businessComplimentaryInformations, err := page.Paginate[*BusinessComplimentaryInformation](query, pag)
	if err != nil {
		return page.Page[*BusinessComplimentaryInformation]{}, fmt.Errorf("could not find business complimentary informations: %w", err)
	}
	return businessComplimentaryInformations, nil
}
