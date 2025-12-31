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
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/financialassistance"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/page"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL        string
	service        financialassistance.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(
	host string,
	service financialassistance.Service,
	consentService consent.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:        host + "/open-insurance/insurance-financial-assistance/v1",
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	authCodeAuthMiddleware := middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, financialassistance.Scope)
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

	handler = http.HandlerFunc(wrapper.GetInsuranceFinancialAssistance)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionFinancialAssistanceRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-financial-assistance/contracts", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceFinancialAssistanceContractInfo)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionFinancialAssistanceContractInfoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-financial-assistance/{contractId}/contract-info", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceFinancialAssistanceMovements)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionFinancialAssistanceMovementsRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-financial-assistance/{contractId}/movements", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/insurance-financial-assistance/v1", handler), swaggerVersion
}

func (s Server) GetInsuranceFinancialAssistance(ctx context.Context, req GetInsuranceFinancialAssistanceRequestObject) (GetInsuranceFinancialAssistanceResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	contracts, err := s.service.ConsentedContracts(ctx, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceFinancialAssistance{
		Meta:  *api.NewPaginatedMeta(contracts),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-financial-assistance/contracts", contracts),
		Data: func() []struct {
			Brand struct {
				Companies []struct {
					CnpjNumber  string `json:"cnpjNumber"`
					CompanyName string `json:"companyName"`
					Contracts   *[]struct {
						ContractID string `json:"contractId"`
					} `json:"contracts,omitempty"`
				} `json:"companies"`
				Name string `json:"name"`
			} `json:"brand"`
		} {
			respContracts := func() *[]struct {
				ContractID string `json:"contractId"`
			} {
				contractsList := make([]struct {
					ContractID string `json:"contractId"`
				}, 0, len(contracts.Records))
				for _, contract := range contracts.Records {
					contractsList = append(contractsList, struct {
						ContractID string `json:"contractId"`
					}{
						ContractID: contract.ID,
					})
				}
				return &contractsList
			}()

			return []struct {
				Brand struct {
					Companies []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Contracts   *[]struct {
							ContractID string `json:"contractId"`
						} `json:"contracts,omitempty"`
					} `json:"companies"`
					Name string `json:"name"`
				} `json:"brand"`
			}{{
				Brand: struct {
					Companies []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Contracts   *[]struct {
							ContractID string `json:"contractId"`
						} `json:"contracts,omitempty"`
					} `json:"companies"`
					Name string `json:"name"`
				}{
					Name: insurer.Brand,
					Companies: []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Contracts   *[]struct {
							ContractID string `json:"contractId"`
						} `json:"contracts,omitempty"`
					}{{
						CnpjNumber:  insurer.CNPJ,
						CompanyName: insurer.Brand,
						Contracts:   respContracts,
					}},
				},
			}}
		}(),
	}

	return GetInsuranceFinancialAssistance200JSONResponse{OKResponseInsuranceFinancialAssistanceJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceFinancialAssistanceContractInfo(ctx context.Context, req GetInsuranceFinancialAssistanceContractInfoRequestObject) (GetInsuranceFinancialAssistanceContractInfoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	contract, err := s.service.ConsentedContract(ctx, string(req.ContractID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceFinancialAssistanceContractInfo{
		Data: InsuranceFinancialAssistanceContractInfo{
			ContractID:              contract.ID,
			CertificateID:           contract.Data.CertificateID,
			GroupContractID:         contract.Data.GroupContractID,
			SusepProcessNumber:      contract.Data.SusepProcessNumber,
			ConceivedCreditValue:    contract.Data.ConceivedCreditValue,
			CreditedLiquidValue:     contract.Data.CreditedLiquidValue,
			InterestRate:            contract.Data.InterestRate,
			EffectiveCostRate:       contract.Data.EffectiveCostRate,
			AmortizationPeriod:      contract.Data.AmortizationPeriod,
			AcquittanceValue:        contract.Data.AcquittanceValue,
			AcquittanceDate:         contract.Data.AcquittanceDate,
			TaxesValue:              contract.Data.TaxesValue,
			ExpensesValue:           contract.Data.ExpensesValue,
			FinesValue:              contract.Data.FinesValue,
			MonetaryUpdatesValue:    contract.Data.MonetaryUpdatesValue,
			AdministrativeFeesValue: contract.Data.AdministrativeFeesValue,
			InterestValue:           contract.Data.InterestValue,
			Insureds: func() []InsuranceFinancialAssistanceInsured {
				insureds := make([]InsuranceFinancialAssistanceInsured, len(contract.Data.Insureds))
				for i, ins := range contract.Data.Insureds {
					insureds[i] = InsuranceFinancialAssistanceInsured{
						DocumentType:       InsuranceFinancialAssistanceInsuredDocumentType(ins.DocumentType),
						DocumentTypeOthers: ins.DocumentTypeOthers,
						DocumentNumber:     ins.DocumentNumber,
						Name:               ins.Name,
					}
				}
				return insureds
			}(),
			CounterInstallments: func() InsuranceFinancialAssistanceCounterInstallments {
				ci := contract.Data.CounterInstallments
				return InsuranceFinancialAssistanceCounterInstallments{
					FirstDate:   ci.FirstDate,
					LastDate:    ci.LastDate,
					Periodicity: InsuranceFinancialAssistanceCounterInstallmentsPeriodicity(ci.Periodicity),
					Quantity:    ci.Quantity,
					Value:       ci.Value,
				}
			}(),
		},
		Links: *api.NewLinks(s.baseURL + "/insurance-financial-assistance/" + string(req.ContractID) + "/contract-info"),
		Meta:  *api.NewMeta(),
	}
	return GetInsuranceFinancialAssistanceContractInfo200JSONResponse{OKResponseInsuranceFinancialAssistanceContractInfoJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceFinancialAssistanceMovements(ctx context.Context, req GetInsuranceFinancialAssistanceMovementsRequestObject) (GetInsuranceFinancialAssistanceMovementsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	movements, err := s.service.ConsentedMovements(ctx, string(req.ContractID), consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceFinancialAssistanceMovements{
		Data: func() []InsuranceFinancialAssistanceMovements {
			respMovements := make([]InsuranceFinancialAssistanceMovements, 0, len(movements.Records))
			for _, movement := range movements.Records {
				respMovements = append(respMovements, InsuranceFinancialAssistanceMovements{
					UpdatedDebitAmount:                         movement.MovementData.UpdatedDebitAmount,
					RemainingCounterInstallmentsQuantity:       movement.MovementData.RemainingCounterInstallmentsQuantity,
					RemainingUnpaidCounterInstallmentsQuantity: movement.MovementData.RemainingUnpaidCounterInstallmentsQuantity,
					LifePensionPmBacAmount:                     movement.MovementData.LifePensionPMBACAmount,
					PensionPlanPmBacAmount:                     movement.MovementData.PensionPlanPMBACAmount,
				})
			}
			return respMovements
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-financial-assistance/"+string(req.ContractID)+"/movements", movements),
		Meta:  *api.NewPaginatedMeta(movements),
	}

	return GetInsuranceFinancialAssistanceMovements200JSONResponse{OKResponseInsuranceFinancialAssistanceMovementsJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
