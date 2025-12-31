//go:generate oapi-codegen -config=./config.yml -package=v1 -o=./api_gen.go ./swagger.yml
package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/financialrisk"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/page"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL        string
	service        financialrisk.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(
	host string,
	service financialrisk.Service,
	consentService consent.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:        host + "/open-insurance/insurance-financial-risk/v1",
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	authCodeAuthMiddleware := middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, financialrisk.Scope)
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

	handler = http.HandlerFunc(wrapper.GetInsuranceFinancialRisk)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleFinancialRisksRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-financial-risk", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceFinancialRiskpolicyIDPolicyInfo)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleFinancialRisksPolicyInfoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-financial-risk/{policyId}/policy-info", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceFinancialRiskpolicyIDPremium)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleFinancialRisksPremiumRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-financial-risk/{policyId}/premium", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceFinancialRiskpolicyIDClaims)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleFinancialRisksClaimRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-financial-risk/{policyId}/claim", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/insurance-financial-risk/v1", handler), swaggerVersion
}

func (s Server) GetInsuranceFinancialRisk(ctx context.Context, req GetInsuranceFinancialRiskRequestObject) (GetInsuranceFinancialRiskResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	policies, err := s.service.ConsentedPolicies(ctx, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceFinancialRisk{
		Meta:  *api.NewPaginatedMeta(policies),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-financial-risk", policies),
		Data: func() []struct {
			Brand     string `json:"brand"`
			Companies []struct {
				CnpjNumber  string `json:"cnpjNumber"`
				CompanyName string `json:"companyName"`
				Policies    []struct {
					PolicyID    string `json:"policyId"`
					ProductName string `json:"productName"`
				} `json:"policies"`
			} `json:"companies"`
		} {
			respPolicies := make([]struct {
				PolicyID    string `json:"policyId"`
				ProductName string `json:"productName"`
			}, 0, len(policies.Records))
			for _, policy := range policies.Records {
				respPolicies = append(respPolicies, struct {
					PolicyID    string `json:"policyId"`
					ProductName string `json:"productName"`
				}{
					PolicyID:    policy.ID,
					ProductName: policy.Data.ProductName,
				})
			}
			return []struct {
				Brand     string `json:"brand"`
				Companies []struct {
					CnpjNumber  string `json:"cnpjNumber"`
					CompanyName string `json:"companyName"`
					Policies    []struct {
						PolicyID    string `json:"policyId"`
						ProductName string `json:"productName"`
					} `json:"policies"`
				} `json:"companies"`
			}{
				{
					Brand: insurer.Brand,
					Companies: []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Policies    []struct {
							PolicyID    string `json:"policyId"`
							ProductName string `json:"productName"`
						} `json:"policies"`
					}{
						{
							CnpjNumber:  insurer.CNPJ,
							CompanyName: insurer.Brand,
							Policies:    respPolicies,
						},
					},
				},
			}
		}(),
	}

	return GetInsuranceFinancialRisk200JSONResponse{OKResponseInsuranceFinancialRiskJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceFinancialRiskpolicyIDPolicyInfo(ctx context.Context, req GetInsuranceFinancialRiskpolicyIDPolicyInfoRequestObject) (GetInsuranceFinancialRiskpolicyIDPolicyInfoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, string(req.PolicyID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	policyInfo := InsuranceFinancialRiskPolicyInfo{
		PolicyID:            policy.ID,
		DocumentType:        InsuranceFinancialRiskPolicyInfoDocumentType(policy.Data.DocumentType),
		IssuanceType:        InsuranceFinancialRiskPolicyInfoIssuanceType(policy.Data.IssuanceType),
		IssuanceDate:        policy.Data.IssuanceDate,
		TermStartDate:       policy.Data.TermStartDate,
		TermEndDate:         policy.Data.TermEndDate,
		ProposalID:          policy.Data.ProposalID,
		SusepProcessNumber:  policy.Data.SusepProcessNumber,
		GroupCertificateID:  policy.Data.GroupCertificateID,
		LeadInsurerCode:     policy.Data.LeadInsurerCode,
		LeadInsurerPolicyID: policy.Data.LeadInsurerPolicyID,
		MaxLMG:              policy.Data.MaxLMG,
		Insureds: func() []PersonalInfo {
			insureds := make([]PersonalInfo, len(policy.Data.Insureds))
			for i, ins := range policy.Data.Insureds {
				insureds[i] = PersonalInfo{
					Identification:           ins.Identification,
					IdentificationType:       PersonalInfoIdentificationType(ins.IdentificationType),
					IdentificationTypeOthers: ins.IdentificationTypeOthers,
					Name:                     ins.Name,
					PostCode:                 ins.PostCode,
					BirthDate:                ins.BirthDate,
					Email:                    ins.Email,
					City:                     ins.City,
					State:                    PersonalInfoState(ins.State),
					Country:                  PersonalInfoCountry(ins.Country),
					Address:                  ins.Address,
				}
			}
			return insureds
		}(),
		Beneficiaries: func() *[]BeneficiaryInfo {
			if policy.Data.Beneficiaries == nil {
				return nil
			}
			beneficiaries := make([]BeneficiaryInfo, len(*policy.Data.Beneficiaries))
			for i, b := range *policy.Data.Beneficiaries {
				beneficiaries[i] = BeneficiaryInfo{
					Identification:           b.Identification,
					IdentificationType:       BeneficiaryInfoIdentificationType(b.IdentificationType),
					IdentificationTypeOthers: b.IdentificationTypeOthers,
					Name:                     b.Name,
				}
			}
			return &beneficiaries
		}(),
		Principals: func() *[]PrincipalInfo {
			if policy.Data.Principals == nil {
				return nil
			}
			principals := make([]PrincipalInfo, len(*policy.Data.Principals))
			for i, p := range *policy.Data.Principals {
				principals[i] = PrincipalInfo{
					Identification:           p.Identification,
					IdentificationType:       PrincipalInfoIdentificationType(p.IdentificationType),
					IdentificationTypeOthers: p.IdentificationTypeOthers,
					Name:                     p.Name,
					PostCode:                 p.PostCode,
					Email:                    p.Email,
					City:                     p.City,
					State:                    PrincipalInfoState(p.State),
					Country:                  PrincipalInfoCountry(p.Country),
					Address:                  p.Address,
					AddressAdditionalInfo:    p.AddressAdditionalInfo,
				}
			}
			return &principals
		}(),
		Intermediaries: func() *[]Intermediary {
			if policy.Data.Intermediaries == nil {
				return nil
			}
			intermediaries := make([]Intermediary, len(*policy.Data.Intermediaries))
			for i, in := range *policy.Data.Intermediaries {
				intermediaries[i] = Intermediary{
					Type:           IntermediaryType(in.Type),
					TypeOthers:     in.TypeOthers,
					Identification: in.Identification,
					BrokerID:       in.BrokerID,
					IdentificationType: func() *IntermediaryIdentificationType {
						if in.IdentificationType == nil {
							return nil
						}
						idType := IntermediaryIdentificationType(*in.IdentificationType)
						return &idType
					}(),
					IdentificationTypeOthers: in.IdentificationTypeOthers,
					Name:                     in.Name,
					PostCode:                 in.PostCode,
					City:                     in.City,
					State:                    in.State,
					Country:                  in.Country,
					Address:                  in.Address,
				}
			}
			return &intermediaries
		}(),
		InsuredObjects: func() []InsuranceFinancialRiskInsuredObject {
			insuredObjects := make([]InsuranceFinancialRiskInsuredObject, len(policy.Data.InsuredObjects))
			for i, obj := range policy.Data.InsuredObjects {
				insuredObjects[i] = InsuranceFinancialRiskInsuredObject{
					Identification:     obj.Identification,
					Type:               InsuranceFinancialRiskInsuredObjectType(obj.Type),
					TypeAdditionalInfo: obj.TypeAdditionalInfo,
					Description:        obj.Description,
					Amount:             obj.Amount,
					Coverages: func() []InsuranceFinancialRiskInsuredObjectCoverage {
						coverages := make([]InsuranceFinancialRiskInsuredObjectCoverage, len(obj.Coverages))
						for j, cov := range obj.Coverages {
							coverages[j] = InsuranceFinancialRiskInsuredObjectCoverage{
								Branch:             cov.Branch,
								Code:               InsuranceFinancialRiskInsuredObjectCoverageCode(cov.Code),
								Description:        cov.Description,
								InternalCode:       cov.InternalCode,
								SusepProcessNumber: cov.SusepProcessNumber,
								LMI:                cov.LMI,
								TermStartDate:      cov.TermStartDate,
								TermEndDate:        cov.TermEndDate,
								IsMainCoverage:     cov.IsMainCoverage,
								Feature:            InsuranceFinancialRiskInsuredObjectCoverageFeature(cov.Feature),
								Type:               InsuranceFinancialRiskInsuredObjectCoverageType(cov.Type),
								GracePeriod:        cov.GracePeriod,
								GracePeriodicity: func() *InsuranceFinancialRiskInsuredObjectCoverageGracePeriodicity {
									if cov.GracePeriodicity == nil {
										return nil
									}
									periodicity := InsuranceFinancialRiskInsuredObjectCoverageGracePeriodicity(*cov.GracePeriodicity)
									return &periodicity
								}(),
								GracePeriodCountingMethod: func() *InsuranceFinancialRiskInsuredObjectCoverageGracePeriodCountingMethod {
									if cov.GracePeriodCountingMethod == nil {
										return nil
									}
									method := InsuranceFinancialRiskInsuredObjectCoverageGracePeriodCountingMethod(*cov.GracePeriodCountingMethod)
									return &method
								}(),
								GracePeriodStartDate:     cov.GracePeriodStartDate,
								GracePeriodEndDate:       cov.GracePeriodEndDate,
								IsLMISublimit:            cov.IsLMISublimit,
								PremiumPeriodicity:       InsuranceFinancialRiskInsuredObjectCoveragePremiumPeriodicity(cov.PremiumPeriodicity),
								PremiumPeriodicityOthers: cov.PremiumPeriodicityOthers,
							}
						}
						return coverages
					}(),
				}
			}
			return insuredObjects
		}(),
		Coverages: func() *[]InsuranceFinancialRiskCoverage {
			if policy.Data.Coverages == nil {
				return nil
			}
			coverages := make([]InsuranceFinancialRiskCoverage, len(*policy.Data.Coverages))
			for i, cov := range *policy.Data.Coverages {
				coverages[i] = InsuranceFinancialRiskCoverage{
					Branch:      cov.Branch,
					Code:        InsuranceFinancialRiskCoverageCode(cov.Code),
					Description: cov.Description,
					Deductible: func() *Deductible {
						if cov.Deductible == nil {
							return nil
						}
						return &Deductible{
							Type:               DeductibleType(cov.Deductible.Type),
							TypeAdditionalInfo: cov.Deductible.TypeAdditionalInfo,
							Amount:             cov.Deductible.Amount,
							Period:             cov.Deductible.Period,
							Periodicity:        DeductiblePeriodicity(cov.Deductible.Periodicity),
							PeriodCountingMethod: func() *DeductiblePeriodCountingMethod {
								if cov.Deductible.PeriodCountingMethod == nil {
									return nil
								}
								method := DeductiblePeriodCountingMethod(*cov.Deductible.PeriodCountingMethod)
								return &method
							}(),
							PeriodStartDate: cov.Deductible.PeriodStartDate,
							PeriodEndDate:   cov.Deductible.PeriodEndDate,
							Description:     cov.Deductible.Description,
						}
					}(),
					POS: func() *POS {
						if cov.POS == nil {
							return nil
						}
						return &POS{
							ApplicationType: POSApplicationType(cov.POS.ApplicationType),
							Description:     cov.POS.Description,
							MinValue:        cov.POS.MinValue,
							MaxValue:        cov.POS.MaxValue,
							Percentage:      cov.POS.Percentage,
							ValueOthers:     cov.POS.ValueOthers,
						}
					}(),
				}
			}
			return &coverages
		}(),
		CoinsuranceRetainedPercentage: policy.Data.CoinsuranceRetainedPercentage,
		Coinsurers: func() *[]Coinsurer {
			if policy.Data.Coinsurers == nil {
				return nil
			}
			coinsurers := make([]Coinsurer, len(*policy.Data.Coinsurers))
			for i, c := range *policy.Data.Coinsurers {
				coinsurers[i] = Coinsurer{
					Identification:  c.Identification,
					CededPercentage: c.CededPercentage,
				}
			}
			return &coinsurers
		}(),
		BranchInfo: func() *InsuranceStopLossSpecificPolicyInfo {
			if policy.Data.BranchInfo == nil {
				return nil
			}
			return &InsuranceStopLossSpecificPolicyInfo{
				Identification:   policy.Data.BranchInfo.Identification,
				TechnicalSurplus: policy.Data.BranchInfo.TechnicalSurplus,
				UserGroup:        policy.Data.BranchInfo.UserGroup,
			}
		}(),
	}

	resp := ResponseInsuranceFinancialRiskPolicyInfo{
		Data:  policyInfo,
		Links: *api.NewLinks(s.baseURL + "/insurance-financial-risk/" + string(req.PolicyID) + "/policy-info"),
		Meta:  *api.NewMeta(),
	}

	return GetInsuranceFinancialRiskpolicyIDPolicyInfo200JSONResponse{OKResponseInsuranceFinancialRiskPolicyInfoJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceFinancialRiskpolicyIDPremium(ctx context.Context, req GetInsuranceFinancialRiskpolicyIDPremiumRequestObject) (GetInsuranceFinancialRiskpolicyIDPremiumResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, string(req.PolicyID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	premium := InsuranceFinancialRiskPremium{
		PaymentsQuantity: policy.Data.Premium.PaymentsQuantity,
		Amount:           policy.Data.Premium.Amount,
		Coverages: func() []InsuranceFinancialRiskPremiumCoverage {
			coverages := make([]InsuranceFinancialRiskPremiumCoverage, len(policy.Data.Premium.Coverages))
			for i, cov := range policy.Data.Premium.Coverages {
				coverages[i] = InsuranceFinancialRiskPremiumCoverage{
					Branch:        cov.Branch,
					Code:          InsuranceFinancialRiskPremiumCoverageCode(cov.Code),
					Description:   cov.Description,
					PremiumAmount: cov.PremiumAmount,
				}
			}
			return coverages
		}(),
		Payments: func() []Payment {
			payments := make([]Payment, len(policy.Data.Premium.Payments))
			for i, pay := range policy.Data.Premium.Payments {
				payments[i] = Payment{
					MovementDate: pay.MovementDate,
					MovementType: PaymentMovementType(pay.MovementType),
					MovementOrigin: func() *PaymentMovementOrigin {
						if pay.MovementOrigin == nil {
							return nil
						}
						origin := PaymentMovementOrigin(*pay.MovementOrigin)
						return &origin
					}(),
					MovementPaymentsNumber: pay.MovementPaymentsNumber,
					Amount:                 pay.Amount,
					MaturityDate:           pay.MaturityDate,
					TellerID:               pay.TellerID,
					TellerIDType: func() *PaymentTellerIDType {
						if pay.TellerIDType == nil {
							return nil
						}
						tellerIDType := PaymentTellerIDType(*pay.TellerIDType)
						return &tellerIDType
					}(),
					TellerIDOthers:           pay.TellerIDOthers,
					TellerName:               pay.TellerName,
					FinancialInstitutionCode: pay.FinancialInstitutionCode,
					PaymentType: func() *PaymentPaymentType {
						if pay.PaymentType == nil {
							return nil
						}
						paymentType := PaymentPaymentType(*pay.PaymentType)
						return &paymentType
					}(),
					PaymentTypeOthers: pay.PaymentTypeOthers,
				}
			}
			return payments
		}(),
	}

	resp := ResponseInsuranceFinancialRiskPremium{
		Data:  premium,
		Links: *api.NewLinks(s.baseURL + "/insurance-financial-risk/" + string(req.PolicyID) + "/premium"),
		Meta:  *api.NewMeta(),
	}

	return GetInsuranceFinancialRiskpolicyIDPremium200JSONResponse{OKResponseInsuranceFinancialRiskPremiumJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceFinancialRiskpolicyIDClaims(ctx context.Context, req GetInsuranceFinancialRiskpolicyIDClaimsRequestObject) (GetInsuranceFinancialRiskpolicyIDClaimsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	claims, err := s.service.ConsentedClaims(ctx, string(req.PolicyID), consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceFinancialRiskClaims{
		Meta:  *api.NewPaginatedMeta(claims),
		Links: *api.NewPaginatedLinks(fmt.Sprintf("%s/insurance-financial-risk/%s/claim", s.baseURL, req.PolicyID), claims),
		Data: func() []InsuranceFinancialRiskClaim {
			respClaims := make([]InsuranceFinancialRiskClaim, 0, len(claims.Records))
			for _, claim := range claims.Records {
				respClaims = append(respClaims, InsuranceFinancialRiskClaim{
					Identification:            claim.Data.Identification,
					DocumentationDeliveryDate: claim.Data.DocumentationDeliveryDate,
					Status:                    InsuranceFinancialRiskClaimStatus(claim.Data.Status),
					StatusAlterationDate:      claim.Data.StatusAlterationDate,
					OccurrenceDate:            claim.Data.OccurrenceDate,
					WarningDate:               claim.Data.WarningDate,
					ThirdPartyClaimDate:       claim.Data.ThirdPartyClaimDate,
					Amount:                    claim.Data.Amount,
					DenialJustification: func() *InsuranceFinancialRiskClaimDenialJustification {
						if claim.Data.DenialJustification == nil {
							return nil
						}
						denialJust := InsuranceFinancialRiskClaimDenialJustification(*claim.Data.DenialJustification)
						return &denialJust
					}(),
					DenialJustificationDescription: claim.Data.DenialJustificationDescription,
					Coverages: func() []InsuranceFinancialRiskClaimCoverage {
						coverages := make([]InsuranceFinancialRiskClaimCoverage, 0, len(claim.Data.Coverages))
						for _, cov := range claim.Data.Coverages {
							coverages = append(coverages, InsuranceFinancialRiskClaimCoverage{
								InsuredObjectID:     cov.InsuredObjectID,
								Branch:              cov.Branch,
								Code:                InsuranceFinancialRiskClaimCoverageCode(cov.Code),
								Description:         cov.Description,
								WarningDate:         cov.WarningDate,
								ThirdPartyClaimDate: cov.ThirdPartyClaimDate,
							})
						}
						return coverages
					}(),
				})
			}
			return respClaims
		}(),
	}

	return GetInsuranceFinancialRiskpolicyIDClaims200JSONResponse{OKResponseInsuranceFinancialRiskClaimsJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
