package consent

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"github.com/luikyv/mock-insurer/internal/user"
	"gorm.io/gorm"
)

type Service struct {
	storage     Storage
	userService user.Service
}

func NewService(db *gorm.DB, userService user.Service) Service {
	return Service{
		storage:     storage{db: db},
		userService: userService,
	}
}

func (s Service) Create(ctx context.Context, c *Consent) error {

	if err := validatePermissions(c.Permissions); err != nil {
		return err
	}

	now := timeutil.DateTimeNow()
	if c.ExpiresAt.After(now.AddDate(1, 0, 0)) || c.ExpiresAt.Before(now) {
		return ErrInvalidExpiration
	}

	if u, err := s.userService.User(ctx, user.Query{CPF: c.UserIdentification}, c.OrgID); err == nil {
		c.OwnerID = &u.ID
	}

	if c.BusinessIdentification != nil {
		if u, err := s.userService.User(ctx, user.Query{CNPJ: *c.BusinessIdentification}, c.OrgID); err == nil {
			c.OwnerID = &u.ID
		}
	}

	c.Status = StatusAwaitingAuthorization
	c.StatusUpdatedAt = now
	c.CreatedAt = now
	c.UpdatedAt = now
	return s.storage.create(ctx, c)
}

func (s Service) Authorize(ctx context.Context, c *Consent) error {
	if c.Status != StatusAwaitingAuthorization {
		return errorutil.New("consent is not in the awaiting authorization status")
	}

	return s.updateWithStatus(ctx, c, StatusAuthorized)
}

func (s Service) Consent(ctx context.Context, id, orgID string) (*Consent, error) {
	id = strings.TrimPrefix(id, URNPrefix)
	c, err := s.storage.consent(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if ctx.Value(api.CtxKeyClientID) != nil && ctx.Value(api.CtxKeyClientID) != c.ClientID {
		return nil, ErrAccessNotAllowed
	}

	return c, s.runAutomations(ctx, c)
}

func (s Service) Consents(ctx context.Context, ownerID uuid.UUID, orgID string, pag page.Pagination) (page.Page[*Consent], error) {
	consents, err := s.storage.consents(ctx, orgID, &Filter{OwnerID: ownerID.String()}, pag)
	if err != nil {
		return page.Page[*Consent]{}, err
	}

	for _, c := range consents.Records {
		if err := s.runAutomations(ctx, c); err != nil {
			return page.Page[*Consent]{}, err
		}
	}

	return consents, nil
}

func (s Service) Reject(ctx context.Context, id, orgID string, rejection Rejection) error {
	c, err := s.Consent(ctx, id, orgID)
	if err != nil {
		return err
	}

	return s.reject(ctx, c, rejection)
}

func (s Service) Delete(ctx context.Context, id, orgID string) error {
	c, err := s.Consent(ctx, id, orgID)
	if err != nil {
		return err
	}

	rejectedBy := RejectedByUser
	rejectionReason := RejectionReasonCodeCustomerManuallyRejected
	additionalInfo := "customer manually rejected consent"
	if c.Status == StatusAuthorized {
		rejectionReason = RejectionReasonCodeCustomerManuallyRevoked
		additionalInfo = "customer manually revoked consent after authorization"
	}

	return s.Reject(ctx, id, orgID, Rejection{
		By:                   rejectedBy,
		ReasonCode:           rejectionReason,
		ReasonAdditionalInfo: &additionalInfo,
	})
}

func (s Service) runAutomations(ctx context.Context, c *Consent) error {
	switch c.Status {
	case StatusAwaitingAuthorization:
		if timeutil.DateTimeNow().After(c.CreatedAt.Add(3600 * time.Second)) {
			slog.DebugContext(ctx, "consent awaiting authorization for too long, moving to rejected")
			reasonAdditionalInfo := "consent awaiting authorization for too long"
			return s.reject(ctx, c, Rejection{
				By:                   RejectedByUser,
				ReasonCode:           RejectionReasonCodeConsentExpired,
				ReasonAdditionalInfo: &reasonAdditionalInfo,
			})
		}
	case StatusAuthorized:
		if timeutil.DateTimeNow().After(c.ExpiresAt) {
			slog.DebugContext(ctx, "consent reached expiration, moving to rejected")
			reasonAdditionalInfo := "consent reached expiration"
			return s.reject(ctx, c, Rejection{
				By:                   RejectedByASPSP,
				ReasonCode:           RejectionReasonCodeConsentMaxDateReached,
				ReasonAdditionalInfo: &reasonAdditionalInfo,
			})
		}
	}

	return nil
}

func (s Service) reject(ctx context.Context, c *Consent, rejection Rejection) error {
	if c.Status == StatusRejected {
		return ErrAlreadyRejected
	}

	c.Rejection = &rejection
	return s.updateWithStatus(ctx, c, StatusRejected)
}

func (s Service) updateWithStatus(ctx context.Context, c *Consent, status Status) error {
	c.Status = status
	c.StatusUpdatedAt = timeutil.DateTimeNow()
	return s.update(ctx, c)
}

func (s Service) update(ctx context.Context, c *Consent) error {
	c.UpdatedAt = timeutil.DateTimeNow()
	return s.storage.update(ctx, c)
}
