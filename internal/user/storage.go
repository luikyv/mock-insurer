package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/luikyv/mock-insurer/internal/page"
	"gorm.io/gorm"
)

type Storage interface {
	create(ctx context.Context, c *User) error
	update(ctx context.Context, c *User) error
	user(ctx context.Context, query Query, orgID string) (*User, error)
	users(ctx context.Context, orgID string, pag page.Pagination) (page.Page[*User], error)
	business(ctx context.Context, userID, businessID uuid.UUID, orgID string) (*UserBusiness, error)
	delete(ctx context.Context, id string, orgID string) error
}

type storage struct {
	db *gorm.DB
}

func (s storage) create(ctx context.Context, u *User) error {
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s storage) update(ctx context.Context, u *User) error {
	tx := s.db.WithContext(ctx).
		Model(&User{}).
		Omit("ID", "CreatedAt", "OrgID").
		Where("id = ?", u.ID).
		Updates(u)
	if err := tx.Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s storage) user(ctx context.Context, query Query, orgID string) (*User, error) {
	u := &User{}
	dbQuery := s.db.WithContext(ctx).Where("org_id = ? OR cross_org = true", orgID)
	if query.ID != "" {
		dbQuery = dbQuery.Where("id = ?", query.ID)
	}
	if query.CPF != "" {
		dbQuery = dbQuery.Where("cpf = ?", query.CPF)
	}
	if query.Username != "" {
		dbQuery = dbQuery.Where("username = ?", query.Username)
	}
	if query.CNPJ != "" {
		dbQuery = dbQuery.Where("cnpj = ?", query.CNPJ)
	}

	if err := dbQuery.First(u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s storage) users(ctx context.Context, orgID string, pag page.Pagination) (page.Page[*User], error) {
	query := s.db.WithContext(ctx).
		Model(&User{}).
		Where("org_id = ? OR cross_org = true", orgID).
		Order("created_at DESC")

	users, err := page.Paginate[*User](query, pag)
	if err != nil {
		return page.Page[*User]{}, err
	}

	return users, nil
}

func (s storage) delete(ctx context.Context, id, orgID string) error {
	if err := s.db.WithContext(ctx).Where("id = ? AND org_id = ?", id, orgID).Delete(&User{}).Error; err != nil {
		return fmt.Errorf("could not delete user: %w", err)
	}
	return nil
}
