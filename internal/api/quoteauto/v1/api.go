//go:generate oapi-codegen -config=./config.yml -package=v1 -o=./api_gen.go ./swagger.yml
package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/customer"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/idempotency"
	"github.com/luikyv/mock-insurer/internal/quote"
	"github.com/luikyv/mock-insurer/internal/quote/auto"
)

type Server struct {
	baseURL            string
	service            auto.Service
	idempotencyService idempotency.Service
	op                 *provider.Provider
}

func NewServer(
	host string,
	service auto.Service,
	idempotencyService idempotency.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:            host + "/open-insurance/quote-auto/v1",
		service:            service,
		idempotencyService: idempotencyService,
		op:                 op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	clientCredentialsMiddleware := middleware.Auth(s.op, goidc.GrantClientCredentials, auto.Scope)
	clientCredentialsLeadMiddleware := middleware.Auth(s.op, goidc.GrantClientCredentials, auto.ScopeLead)
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

	handler = http.HandlerFunc(wrapper.PostQuoteAutoLead)
	handler = middleware.Idempotency(s.idempotencyService)(handler)
	handler = clientCredentialsLeadMiddleware(handler)
	mux.Handle("POST /lead/request", handler)

	handler = http.HandlerFunc(wrapper.PostQuoteAutoLead)
	handler = clientCredentialsLeadMiddleware(handler)
	mux.Handle("PATCH /lead/request/{consentId}", handler)

	handler = http.HandlerFunc(wrapper.PostQuoteAuto)
	handler = middleware.Idempotency(s.idempotencyService)(handler)
	handler = clientCredentialsMiddleware(handler)
	mux.Handle("POST /request", handler)

	handler = http.HandlerFunc(wrapper.GetQuoteAuto)
	handler = clientCredentialsMiddleware(handler)
	mux.Handle("GET /request/{consentId}/quote-status", handler)

	handler = http.HandlerFunc(wrapper.PatchQuoteAuto)
	handler = clientCredentialsMiddleware(handler)
	mux.Handle("PATCH /request/{consentId}", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/quote-auto/v1", handler), swaggerVersion
}

func (s Server) PostQuoteAutoLead(ctx context.Context, req PostQuoteAutoLeadRequestObject) (PostQuoteAutoLeadResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	lead := auto.Lead{
		ConsentID: req.Body.Data.ConsentID,
		OrgID:     orgID,
	}

	if identificationData, err := req.Body.Data.QuoteCustomer.IdentificationData.AsPersonalIdentificationData(); err == nil {
		lead.Data.Customer.Personal = &quote.PersonalData{
			Identification: &customer.PersonalIdentificationData{
				UpdateDateTime:          identificationData.UpdateDateTime,
				PersonalID:              identificationData.PersonalID,
				BrandName:               identificationData.BrandName,
				CivilName:               identificationData.CivilName,
				SocialName:              identificationData.SocialName,
				CPF:                     identificationData.CpfNumber,
				HasBrazilianNationality: identificationData.HasBrazilianNationality,
				CompanyInfo: customer.CompanyInfo{
					CNPJ: identificationData.CompanyInfo.CnpjNumber,
					Name: identificationData.CompanyInfo.Name,
				},
			},
		}
	}

	if _, err := req.Body.Data.QuoteCustomer.QualificationData.AsPersonalQualificationData(); err == nil {
		lead.Data.Customer.Personal.Qualification = &customer.PersonalQualificationData{}
	}

	if _, err := req.Body.Data.QuoteCustomer.ComplimentaryInformationData.AsPersonalComplimentaryInformationData(); err == nil {
		lead.Data.Customer.Personal.ComplimentaryInfo = &customer.PersonalComplimentaryInformationData{}
	}

	if _, err := req.Body.Data.QuoteCustomer.IdentificationData.AsBusinessIdentificationData(); err == nil {
		lead.Data.Customer.Business = &quote.BusinessData{
			Identification: &customer.BusinessIdentificationData{},
		}
	}

	if _, err := req.Body.Data.QuoteCustomer.QualificationData.AsBusinessQualificationData(); err == nil {
		lead.Data.Customer.Business.Qualification = &customer.BusinessQualificationData{}
	}

	if _, err := req.Body.Data.QuoteCustomer.ComplimentaryInformationData.AsBusinessComplimentaryInformationData(); err == nil {
		lead.Data.Customer.Business.ComplimentaryInfo = &customer.BusinessComplimentaryInformationData{}
	}

	err := s.service.CreateLead(ctx, &lead)
	if err != nil {
		return nil, err
	}

	resp := ResponseQuoteAutoLead{
		Data: QuoteStatus{
			Status:               QuoteStatusStatus(lead.Status),
			StatusUpdateDateTime: lead.StatusUpdatedAt,
		},
		Meta:  *api.NewMeta(),
		Links: *api.NewLinks(s.baseURL + "/lead/request/" + lead.ConsentID + "/quote-status"),
	}
	return PostQuoteAutoLead201JSONResponse{CreatedResponseQuoteRequestAutoLeadJSONResponse(resp)}, nil
}

func (s Server) PatchQuoteAutoLead(ctx context.Context, request PatchQuoteAutoLeadRequestObject) (PatchQuoteAutoLeadResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	lead, err := s.service.UpdateLead(ctx, request.ConsentID, orgID, quote.PatchData{})
	if err != nil {
		return nil, err
	}

	resp := ResponseRevokePatch{
		Data: struct {
			Status ResponseRevokePatchDataStatus `json:"status"`
		}{
			Status: ResponseRevokePatchDataStatus(lead.Status),
		},
	}
	return PatchQuoteAutoLead200JSONResponse{N200UpdatedQuoteAutoLeadJSONResponse(resp)}, nil
}

func (s Server) PostQuoteAuto(ctx context.Context, request PostQuoteAutoRequestObject) (PostQuoteAutoResponseObject, error) {
	return nil, nil
}

func (s Server) PatchQuoteAuto(ctx context.Context, request PatchQuoteAutoRequestObject) (PatchQuoteAutoResponseObject, error) {
	return nil, nil
}

func (s Server) GetQuoteAuto(ctx context.Context, request GetQuoteAutoRequestObject) (GetQuoteAutoResponseObject, error) {
	return nil, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
