package user

import (
	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username    string
	Name        string
	CPF         string
	CNPJ        *string
	Description *string

	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (User) TableName() string {
	return "mock_users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type Query struct {
	ID       string
	Username string
	CPF      string
	CNPJ     string
}

type UserBusiness struct {
	UserID         uuid.UUID
	BusinessUserID uuid.UUID

	OrgID     string
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (UserBusiness) TableName() string {
	return "mock_user_business"
}
