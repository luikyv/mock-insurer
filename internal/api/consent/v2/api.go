//go:generate oapi-codegen -config=./config.yml -package=v2 -o=./api_gen.go ./swagger.yml
package v2

import (
	"context"
	"errors"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/idempotency"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

type Server struct {
	baseURL            string
	service            consent.Service
	op                 *provider.Provider
	idempotencyService idempotency.Service
}

func NewServer(host string, service consent.Service, op *provider.Provider, idempotencyService idempotency.Service) Server {
	return Server{
		baseURL:            host + "/open-insurance/consents/v2",
		service:            service,
		op:                 op,
		idempotencyService: idempotencyService,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	clientCredentialsAuthMiddleware := middleware.Auth(s.op, goidc.GrantClientCredentials, consent.Scope)
	swaggerMiddleware, swaggerVersion := middleware.Swagger(GetSwagger, func(err error) api.Error {
		return api.NewError("INVALID_REQUEST", http.StatusBadRequest, err.Error())
	})

	wrapper := ServerInterfaceWrapper{
		Handler: NewStrictHandlerWithOptions(s, nil, StrictHTTPServerOptions{
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				writeResponseError(w, r, err)
			},
		}),
		HandlerMiddlewares: []MiddlewareFunc{swaggerMiddleware},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusBadRequest, err.Error()))
		},
	}

	var handler http.Handler

	handler = http.HandlerFunc(wrapper.ConsentsPostConsents)
	handler = middleware.Idempotency(s.idempotencyService)(handler)
	handler = clientCredentialsAuthMiddleware(handler)
	mux.Handle("POST /consents", handler)

	handler = http.HandlerFunc(wrapper.ConsentsDeleteConsentsConsentID)
	handler = clientCredentialsAuthMiddleware(handler)
	mux.Handle("DELETE /consents/{consentId}", handler)

	handler = http.HandlerFunc(wrapper.ConsentsGetConsentsConsentID)
	handler = clientCredentialsAuthMiddleware(handler)
	mux.Handle("GET /consents/{consentId}", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/consents/v2", handler), swaggerVersion
}

func (s Server) ConsentsPostConsents(ctx context.Context, req ConsentsPostConsentsRequestObject) (ConsentsPostConsentsResponseObject, error) {
	var perms []consent.Permission
	for _, p := range req.Body.Data.Permissions {
		perms = append(perms, consent.Permission(p))
	}
	c := &consent.Consent{
		Status:             consent.StatusAwaitingAuthorization,
		UserIdentification: req.Body.Data.LoggedUser.Document.Identification,
		UserRel:            consent.Relation(req.Body.Data.LoggedUser.Document.Rel),
		Permissions:        perms,
		ExpiresAt:          req.Body.Data.ExpirationDateTime,
		ClientID:           ctx.Value(api.CtxKeyClientID).(string),
		OrgID:              ctx.Value(api.CtxKeyOrgID).(string),
	}

	if business := req.Body.Data.BusinessEntity; business != nil {
		rel := consent.Relation(business.Document.Rel)
		c.BusinessIdentification = &business.Document.Identification
		c.BusinessRel = &rel
	}

	if err := s.service.Create(ctx, c); err != nil {
		return nil, err
	}

	var respPerms []ResponseConsentDataPermissions
	for _, p := range c.Permissions {
		respPerms = append(respPerms, ResponseConsentDataPermissions(p))
	}
	resp := ResponseConsent{
		Data: struct {
			ClaimNotificationInformation *struct {
				DocumentType          ResponseConsentDataClaimNotificationInformationDocumentType `json:"documentType"`
				GroupCertificateID    *string                                                     `json:"groupCertificateId,omitempty"`
				InsuredObjectID       []string                                                    `json:"insuredObjectId"`
				OccurrenceDate        timeutil.BrazilDate                                         `json:"occurrenceDate"`
				OccurrenceDescription string                                                      `json:"occurrenceDescription"`
				OccurrenceTime        *string                                                     `json:"occurrenceTime,omitempty"`
				PolicyID              string                                                      `json:"policyId"`
				ProposalID            *string                                                     `json:"proposalId,omitempty"`
			} `json:"claimNotificationInformation,omitempty"`
			ConsentID              string            `json:"consentId"`
			CreationDateTime       timeutil.DateTime `json:"creationDateTime"`
			EndorsementInformation *struct {
				EndorsementType    ResponseConsentDataEndorsementInformationEndorsementType `json:"endorsementType"`
				InsuredObjectID    []string                                                 `json:"insuredObjectId"`
				PolicyID           string                                                   `json:"policyId"`
				ProposalID         *string                                                  `json:"proposalId,omitempty"`
				RequestDescription string                                                   `json:"requestDescription"`
			} `json:"endorsementInformation,omitempty"`
			ExpirationDateTime                  timeutil.DateTime                `json:"expirationDateTime"`
			Permissions                         []ResponseConsentDataPermissions `json:"permissions"`
			RaffleCaptalizationTitleInformation *struct {
				ContactType ResponseConsentDataRaffleCaptalizationTitleInformationContactType `json:"contactType"`
				Email       *string                                                           `json:"email,omitempty"`
				Phone       *string                                                           `json:"phone,omitempty"`
			} `json:"raffleCaptalizationTitleInformation,omitempty"`
			Rejection *struct {
				Reason     RejectedReason `json:"reason"`
				RejectedBy EnumRejectedBy `json:"rejectedBy"`
			} `json:"rejection,omitempty"`
			Status                             ResponseConsentDataStatus `json:"status"`
			StatusUpdateDateTime               timeutil.DateTime         `json:"statusUpdateDateTime"`
			WithdrawalCaptalizationInformation *struct {
				CapitalizationTitleName string                                                                `json:"capitalizationTitleName"`
				PlanID                  string                                                                `json:"planId"`
				SeriesID                string                                                                `json:"seriesId"`
				TermEndDate             timeutil.BrazilDate                                                   `json:"termEndDate"`
				TitleID                 string                                                                `json:"titleId"`
				WithdrawalReason        ResponseConsentDataWithdrawalCaptalizationInformationWithdrawalReason `json:"withdrawalReason"`
				WithdrawalReasonOthers  *string                                                               `json:"withdrawalReasonOthers,omitempty"`
				WithdrawalTotalAmount   AmountDetails                                                         `json:"withdrawalTotalAmount"`
			} `json:"withdrawalCaptalizationInformation,omitempty"`
			WithdrawalLifePensionInformation *struct {
				CertificateID          string                                                              `json:"certificateId"`
				DesiredTotalAmount     *AmountDetails                                                      `json:"desiredTotalAmount,omitempty"`
				PmbacAmount            AmountDetails                                                       `json:"pmbacAmount"`
				ProductName            string                                                              `json:"productName"`
				WithdrawalReason       ResponseConsentDataWithdrawalLifePensionInformationWithdrawalReason `json:"withdrawalReason"`
				WithdrawalReasonOthers *string                                                             `json:"withdrawalReasonOthers,omitempty"`
				WithdrawalType         ResponseConsentDataWithdrawalLifePensionInformationWithdrawalType   `json:"withdrawalType"`
			} `json:"withdrawalLifePensionInformation,omitempty"`
		}{
			ConsentID:            c.URN(),
			CreationDateTime:     c.CreatedAt,
			ExpirationDateTime:   c.ExpiresAt,
			Permissions:          respPerms,
			Status:               ResponseConsentDataStatus(c.Status),
			StatusUpdateDateTime: c.StatusUpdatedAt,
		},
		Links: api.NewLinks(s.baseURL + "/consents/" + c.URN()),
		Meta:  api.NewMeta(),
	}

	return ConsentsPostConsents201JSONResponse{N201ConsentsCreatedJSONResponse(resp)}, nil
}

func (s Server) ConsentsGetConsentsConsentID(ctx context.Context, req ConsentsGetConsentsConsentIDRequestObject) (ConsentsGetConsentsConsentIDResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	c, err := s.service.Consent(ctx, req.ConsentID, orgID)
	if err != nil {
		return nil, err
	}

	var respPerms []ResponseConsentDataPermissions
	for _, p := range c.Permissions {
		respPerms = append(respPerms, ResponseConsentDataPermissions(p))
	}
	resp := ResponseConsent{
		Data: struct {
			ClaimNotificationInformation *struct {
				DocumentType          ResponseConsentDataClaimNotificationInformationDocumentType `json:"documentType"`
				GroupCertificateID    *string                                                     `json:"groupCertificateId,omitempty"`
				InsuredObjectID       []string                                                    `json:"insuredObjectId"`
				OccurrenceDate        timeutil.BrazilDate                                         `json:"occurrenceDate"`
				OccurrenceDescription string                                                      `json:"occurrenceDescription"`
				OccurrenceTime        *string                                                     `json:"occurrenceTime,omitempty"`
				PolicyID              string                                                      `json:"policyId"`
				ProposalID            *string                                                     `json:"proposalId,omitempty"`
			} `json:"claimNotificationInformation,omitempty"`
			ConsentID              string            `json:"consentId"`
			CreationDateTime       timeutil.DateTime `json:"creationDateTime"`
			EndorsementInformation *struct {
				EndorsementType    ResponseConsentDataEndorsementInformationEndorsementType `json:"endorsementType"`
				InsuredObjectID    []string                                                 `json:"insuredObjectId"`
				PolicyID           string                                                   `json:"policyId"`
				ProposalID         *string                                                  `json:"proposalId,omitempty"`
				RequestDescription string                                                   `json:"requestDescription"`
			} `json:"endorsementInformation,omitempty"`
			ExpirationDateTime                  timeutil.DateTime                `json:"expirationDateTime"`
			Permissions                         []ResponseConsentDataPermissions `json:"permissions"`
			RaffleCaptalizationTitleInformation *struct {
				ContactType ResponseConsentDataRaffleCaptalizationTitleInformationContactType `json:"contactType"`
				Email       *string                                                           `json:"email,omitempty"`
				Phone       *string                                                           `json:"phone,omitempty"`
			} `json:"raffleCaptalizationTitleInformation,omitempty"`
			Rejection *struct {
				Reason     RejectedReason `json:"reason"`
				RejectedBy EnumRejectedBy `json:"rejectedBy"`
			} `json:"rejection,omitempty"`
			Status                             ResponseConsentDataStatus `json:"status"`
			StatusUpdateDateTime               timeutil.DateTime         `json:"statusUpdateDateTime"`
			WithdrawalCaptalizationInformation *struct {
				CapitalizationTitleName string                                                                `json:"capitalizationTitleName"`
				PlanID                  string                                                                `json:"planId"`
				SeriesID                string                                                                `json:"seriesId"`
				TermEndDate             timeutil.BrazilDate                                                   `json:"termEndDate"`
				TitleID                 string                                                                `json:"titleId"`
				WithdrawalReason        ResponseConsentDataWithdrawalCaptalizationInformationWithdrawalReason `json:"withdrawalReason"`
				WithdrawalReasonOthers  *string                                                               `json:"withdrawalReasonOthers,omitempty"`
				WithdrawalTotalAmount   AmountDetails                                                         `json:"withdrawalTotalAmount"`
			} `json:"withdrawalCaptalizationInformation,omitempty"`
			WithdrawalLifePensionInformation *struct {
				CertificateID          string                                                              `json:"certificateId"`
				DesiredTotalAmount     *AmountDetails                                                      `json:"desiredTotalAmount,omitempty"`
				PmbacAmount            AmountDetails                                                       `json:"pmbacAmount"`
				ProductName            string                                                              `json:"productName"`
				WithdrawalReason       ResponseConsentDataWithdrawalLifePensionInformationWithdrawalReason `json:"withdrawalReason"`
				WithdrawalReasonOthers *string                                                             `json:"withdrawalReasonOthers,omitempty"`
				WithdrawalType         ResponseConsentDataWithdrawalLifePensionInformationWithdrawalType   `json:"withdrawalType"`
			} `json:"withdrawalLifePensionInformation,omitempty"`
		}{
			ConsentID:            c.URN(),
			CreationDateTime:     c.CreatedAt,
			ExpirationDateTime:   c.ExpiresAt,
			Permissions:          respPerms,
			Status:               ResponseConsentDataStatus(c.Status),
			StatusUpdateDateTime: c.StatusUpdatedAt,
		},
		Links: api.NewLinks(s.baseURL + "/consents/" + c.URN()),
		Meta:  api.NewMeta(),
	}

	if c.Rejection != nil {
		resp.Data.Rejection = &struct {
			// Reason Define a razão pela qual o consentimento foi rejeitado.
			Reason RejectedReason `json:"reason"`

			// RejectedBy Informar usuário responsável pela rejeição.
			// 1. USER usuário
			// 2. ASPSP instituição transmissora
			// 3. TPP instituição receptora
			RejectedBy EnumRejectedBy `json:"rejectedBy"`
		}{
			Reason: RejectedReason{
				Code:                  EnumReasonCode(c.Rejection.ReasonCode),
				AdditionalInformation: c.Rejection.ReasonAdditionalInfo,
			},
			RejectedBy: EnumRejectedBy(c.Rejection.By),
		}
	}

	return ConsentsGetConsentsConsentID200JSONResponse{N200ConsentsConsentIDReadJSONResponse(resp)}, nil
}

func (s Server) ConsentsDeleteConsentsConsentID(ctx context.Context, req ConsentsDeleteConsentsConsentIDRequestObject) (ConsentsDeleteConsentsConsentIDResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	if err := s.service.Delete(ctx, req.ConsentID, orgID); err != nil {
		return nil, err
	}

	return ConsentsDeleteConsentsConsentID204Response{}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, consent.ErrAccessNotAllowed) {
		api.WriteError(w, r, api.NewError("FORBIDDEN", http.StatusForbidden, err.Error()))
		return
	}

	// if errors.Is(err, consent.ErrInvalidPermissionGroup) {
	// 	api.WriteError(w, r, api.NewError("COMBINACAO_PERMISSOES_INCORRETA", http.StatusUnprocessableEntity, consent.ErrInvalidPermissionGroup.Error()))
	// 	return
	// }

	if errors.Is(err, consent.ErrPersonalAndBusinessPermissionsTogether) {
		api.WriteError(w, r, api.NewError("PERMISSAO_PF_PJ_EM_CONJUNTO", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	if errors.Is(err, consent.ErrInvalidExpiration) {
		api.WriteError(w, r, api.NewError("DATA_EXPIRACAO_INVALIDA", http.StatusBadRequest, err.Error()))
		return
	}

	if errors.Is(err, consent.ErrAlreadyRejected) {
		api.WriteError(w, r, api.NewError("CONSENTIMENTO_EM_STATUS_REJEITADO", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	if errors.Is(err, consent.ErrInvalidPermissions) {
		api.WriteError(w, r, api.NewError("NAO_INFORMADO", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	if errors.Is(err, consent.ErrPermissionResourcesReadAlone) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusBadRequest, err.Error()))
		return
	}

	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("NAO_INFORMADO", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
