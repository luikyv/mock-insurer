//go:generate oapi-codegen -config=./config.yml -package=v1 -o=./api_gen.go ./swagger.yml
package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/acceptancebranchesabroad"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL        string
	service        acceptancebranchesabroad.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(
	host string,
	service acceptancebranchesabroad.Service,
	consentService consent.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:        host + "/open-insurance/insurance-acceptance-and-branches-abroad/v1",
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	authCodeAuthMiddleware := middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, acceptancebranchesabroad.Scope)
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

	handler = http.HandlerFunc(wrapper.GetInsuranceAcceptanceAndBranchesAbroad)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-acceptance-and-branches-abroad", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPolicyInfo)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadPolicyInfoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-acceptance-and-branches-abroad/{policyId}/policy-info", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPremium)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadPremiumRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-acceptance-and-branches-abroad/{policyId}/premium", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceAcceptanceAndBranchesAbroadpolicyIDClaims)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadClaimRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-acceptance-and-branches-abroad/{policyId}/claim", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/insurance-acceptance-and-branches-abroad/v1", handler), swaggerVersion
}

func (s Server) GetInsuranceAcceptanceAndBranchesAbroad(ctx context.Context, req GetInsuranceAcceptanceAndBranchesAbroadRequestObject) (GetInsuranceAcceptanceAndBranchesAbroadResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	policies, err := s.service.ConsentedPolicies(ctx, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceAcceptanceAndBranchesAbroad{
		Meta:  *api.NewPaginatedMeta(policies),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-acceptance-and-branches-abroad", policies),
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
			}{{
				Brand: insurer.Brand,
				Companies: []struct {
					CnpjNumber  string `json:"cnpjNumber"`
					CompanyName string `json:"companyName"`
					Policies    []struct {
						PolicyID    string `json:"policyId"`
						ProductName string `json:"productName"`
					} `json:"policies"`
				}{{
					CnpjNumber:  insurer.CNPJ,
					CompanyName: insurer.Brand,
					Policies:    respPolicies,
				}},
			}}
		}(),
	}

	return GetInsuranceAcceptanceAndBranchesAbroad200JSONResponse{OKResponseInsuranceAcceptanceAndBranchesAbroadJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPolicyInfo(ctx context.Context, request GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPolicyInfoRequestObject) (GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPolicyInfoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, request.PolicyID, consentID, orgID)
	if err != nil {
		return nil, err
	}

	policyInfo := InsuranceAcceptanceAndBranchesAbroadPolicyInfo{
		PolicyID:                      policy.ID,
		DocumentType:                  InsuranceAcceptanceAndBranchesAbroadPolicyInfoDocumentType(policy.Data.DocumentType),
		IssuanceType:                  InsuranceAcceptanceAndBranchesAbroadPolicyInfoIssuanceType(policy.Data.IssuanceType),
		IssuanceDate:                  policy.Data.IssuanceDate,
		TermStartDate:                 policy.Data.TermStartDate,
		TermEndDate:                   policy.Data.TermEndDate,
		ProposalID:                    policy.Data.ProposalID,
		MaxLMG:                        policy.Data.MaxLMG,
		SusepProcessNumber:            policy.Data.SusepProcessNumber,
		GroupCertificateID:            policy.Data.GroupCertificateID,
		LeadInsurerCode:               policy.Data.LeadInsurerCode,
		LeadInsurerPolicyID:           policy.Data.LeadInsurerPolicyID,
		CoinsuranceRetainedPercentage: policy.Data.CoinsuranceRetainedPercentage,
		BranchInfo: InsuranceAcceptanceAndBranchesAbroadSpecificPolicyInfo{
			RiskCountry:      InsuranceAcceptanceAndBranchesAbroadSpecificPolicyInfoRiskCountry(policy.Data.BranchInfo.RiskCountry),
			HasForum:         policy.Data.BranchInfo.HasForum,
			ForumDescription: policy.Data.BranchInfo.ForumDescription,
			TransferorID:     policy.Data.BranchInfo.TransferorID,
			TransferorName:   policy.Data.BranchInfo.TransferorName,
			GroupBranches:    policy.Data.BranchInfo.GroupBranches,
		},
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
					AdressAditionalInfo:      p.AddressAdditionalInfo,
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
					Type:                     IntermediaryType(in.Type),
					TypeOthers:               in.TypeOthers,
					Identification:           in.Identification,
					BrokerID:                 in.BrokerID,
					IdentificationTypeOthers: in.IdentificationTypeOthers,
					Name:                     in.Name,
					PostCode:                 in.PostCode,
					City:                     in.City,
					Address:                  in.Address,
					IdentificationType: func() *IntermediaryIdentificationType {
						if in.IdentificationType == nil {
							return nil
						}
						idType := IntermediaryIdentificationType(*in.IdentificationType)
						return &idType
					}(),
					State: func() *IntermediaryState {
						if in.State == nil {
							return nil
						}
						state := IntermediaryState(*in.State)
						return &state
					}(),
					Country: func() *IntermediaryCountry {
						if in.Country == nil {
							return nil
						}
						country := IntermediaryCountry(*in.Country)
						return &country
					}(),
				}
			}
			return &intermediaries
		}(),
		Insureds: func() []PersonalInfo {
			insureds := make([]PersonalInfo, len(policy.Data.Insureds))
			for i, ins := range policy.Data.Insureds {
				insureds[i] = PersonalInfo{
					Identification:           ins.Identification,
					IdentificationType:       PersonalInfoIdentificationType(ins.IdentificationType),
					IdentificationTypeOthers: ins.IdentificationTypeOthers,
					Name:                     ins.Name,
					BirthDate:                ins.BirthDate,
					PostCode:                 ins.PostCode,
					Email:                    ins.Email,
					City:                     ins.City,
					State:                    PersonalInfoState(ins.State),
					Country:                  PersonalInfoCountry(ins.Country),
					Address:                  ins.Address,
					AdressAditionalInfo:      ins.AddressAdditionalInfo,
				}
			}
			return insureds
		}(),
		InsuredObjects: func() []InsuranceAcceptanceAndBranchesAbroadInsuredObject {
			insuredObjects := make([]InsuranceAcceptanceAndBranchesAbroadInsuredObject, len(policy.Data.InsuredObjects))
			for i, obj := range policy.Data.InsuredObjects {
				insuredObjects[i] = InsuranceAcceptanceAndBranchesAbroadInsuredObject{
					Identification:     obj.Identification,
					Type:               InsuranceAcceptanceAndBranchesAbroadInsuredObjectType(obj.Type),
					TypeAdditionalInfo: obj.TypeAdditionalInfo,
					Description:        obj.Description,
					Amount:             obj.Amount,
					Coverages: func() []InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverage {
						coverages := make([]InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverage, len(obj.Coverages))
						for j, cov := range obj.Coverages {
							coverages[j] = InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverage{
								Branch:             cov.Branch,
								Code:               InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverageCode(cov.Code),
								Description:        cov.Description,
								InternalCode:       cov.InternalCode,
								SusepProcessNumber: cov.SusepProcessNumber,
								LMI:                cov.LMI,
								TermStartDate:      cov.TermStartDate,
								TermEndDate:        cov.TermEndDate,
								IsMainCoverage:     cov.IsMainCoverage,
								Feature:            InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverageFeature(cov.Feature),
								Type:               InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverageType(cov.Type),
								GracePeriod:        cov.GracePeriod,
								GracePeriodicity: func() *InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverageGracePeriodicity {
									if cov.GracePeriodicity == nil {
										return nil
									}
									periodicity := InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverageGracePeriodicity(*cov.GracePeriodicity)
									return &periodicity
								}(),
								GracePeriodCountingMethod: func() *InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverageGracePeriodCountingMethod {
									if cov.GracePeriodCountingMethod == nil {
										return nil
									}
									method := InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoverageGracePeriodCountingMethod(*cov.GracePeriodCountingMethod)
									return &method
								}(),
								GracePeriodStartDate:     cov.GracePeriodStartDate,
								GracePeriodEndDate:       cov.GracePeriodEndDate,
								PremiumPeriodicity:       InsuranceAcceptanceAndBranchesAbroadInsuredObjectCoveragePremiumPeriodicity(cov.PremiumPeriodicity),
								PremiumPeriodicityOthers: cov.PremiumPeriodicityOthers,
							}
						}
						return coverages
					}(),
				}
			}
			return insuredObjects
		}(),
		Coverages: func() *[]InsuranceAcceptanceAndBranchesAbroadCoverage {
			if policy.Data.Coverages == nil {
				return nil
			}
			coverages := make([]InsuranceAcceptanceAndBranchesAbroadCoverage, len(*policy.Data.Coverages))
			for i, cov := range *policy.Data.Coverages {
				coverages[i] = InsuranceAcceptanceAndBranchesAbroadCoverage{
					Branch:      cov.Branch,
					Code:        InsuranceAcceptanceAndBranchesAbroadCoverageCode(cov.Code),
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
							Description:        cov.Deductible.Description,
							PeriodStartDate:    cov.Deductible.PeriodStartDate,
							PeriodEndDate:      cov.Deductible.PeriodEndDate,
							Periodicity:        DeductiblePeriodicity(cov.Deductible.Periodicity),
							PeriodCountingMethod: func() *DeductiblePeriodCountingMethod {
								if cov.Deductible.PeriodCountingMethod == nil {
									return nil
								}
								method := DeductiblePeriodCountingMethod(*cov.Deductible.PeriodCountingMethod)
								return &method
							}(),
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
	}

	resp := ResponseInsuranceAcceptanceAndBranchesAbroadPolicyInfo{
		Data:  policyInfo,
		Links: *api.NewLinks(s.baseURL + "/insurance-acceptance-and-branches-abroad/" + request.PolicyID + "/policy-info"),
		Meta:  *api.NewMeta(),
	}

	return GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPolicyInfo200JSONResponse{OKResponseInsuranceAcceptanceAndBranchesAbroadPolicyInfoJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPremium(ctx context.Context, request GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPremiumRequestObject) (GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPremiumResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, request.PolicyID, consentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceAcceptanceAndBranchesAbroadPremium{
		Meta:  *api.NewMeta(),
		Links: *api.NewLinks(s.baseURL + "/insurance-acceptance-and-branches-abroad/" + request.PolicyID + "/premium"),
		Data: InsuranceAcceptanceAndBranchesAbroadPremium{
			PaymentsQuantity: policy.Data.Premium.PaymentsQuantity,
			Amount:           policy.Data.Premium.Amount,
			Coverages: func() []InsuranceAcceptanceAndBranchesAbroadPremiumCoverage {
				coverages := make([]InsuranceAcceptanceAndBranchesAbroadPremiumCoverage, 0, len(policy.Data.Premium.Coverages))
				for _, cov := range policy.Data.Premium.Coverages {
					coverages = append(coverages, InsuranceAcceptanceAndBranchesAbroadPremiumCoverage{
						Branch:        cov.Branch,
						Code:          InsuranceAcceptanceAndBranchesAbroadPremiumCoverageCode(cov.Code),
						Description:   cov.Description,
						PremiumAmount: cov.PremiumAmount,
					})
				}
				return coverages
			}(),
			Payments: func() []Payment {
				payments := make([]Payment, 0, len(policy.Data.Premium.Payments))
				for _, pay := range policy.Data.Premium.Payments {
					payments = append(payments, Payment{
						MovementDate:             pay.MovementDate,
						MovementType:             PaymentMovementType(pay.MovementType),
						MovementPaymentsNumber:   pay.MovementPaymentsNumber,
						Amount:                   pay.Amount,
						MaturityDate:             pay.MaturityDate,
						TellerID:                 pay.TellerID,
						TellerIDOthers:           pay.TellerIDOthers,
						TellerName:               pay.TellerName,
						FinancialInstitutionCode: pay.FinancialInstitutionCode,
						PaymentTypeOthers:        pay.PaymentTypeOthers,
						MovementOrigin: func() *PaymentMovementOrigin {
							if pay.MovementOrigin == nil {
								return nil
							}
							origin := PaymentMovementOrigin(*pay.MovementOrigin)
							return &origin
						}(),
						TellerIDType: func() *PaymentTellerIDType {
							if pay.TellerIDType == nil {
								return nil
							}
							tellerIDType := PaymentTellerIDType(*pay.TellerIDType)
							return &tellerIDType
						}(),
						PaymentType: func() *PaymentPaymentType {
							if pay.PaymentType == nil {
								return nil
							}
							paymentType := PaymentPaymentType(*pay.PaymentType)
							return &paymentType
						}(),
					})
				}
				return payments
			}(),
		},
	}

	return GetInsuranceAcceptanceAndBranchesAbroadpolicyIDPremium200JSONResponse{OKResponseInsuranceAcceptanceAndBranchesAbroadPremiumJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceAcceptanceAndBranchesAbroadpolicyIDClaims(ctx context.Context, request GetInsuranceAcceptanceAndBranchesAbroadpolicyIDClaimsRequestObject) (GetInsuranceAcceptanceAndBranchesAbroadpolicyIDClaimsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(request.Params.Page, request.Params.PageSize)
	claims, err := s.service.ConsentedClaims(ctx, request.PolicyID, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceAcceptanceAndBranchesAbroadClaims{
		Meta:  *api.NewPaginatedMeta(claims),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-acceptance-and-branches-abroad/"+request.PolicyID+"/claim", claims),
		Data: func() []InsuranceAcceptanceAndBranchesAbroadClaim {
			respClaims := make([]InsuranceAcceptanceAndBranchesAbroadClaim, 0, len(claims.Records))
			for _, claim := range claims.Records {
				respClaims = append(respClaims, InsuranceAcceptanceAndBranchesAbroadClaim{
					Identification:                 claim.Data.Identification,
					DocumentationDeliveryDate:      claim.Data.DocumentDeliveryDate,
					Status:                         InsuranceAcceptanceAndBranchesAbroadClaimStatus(claim.Data.Status),
					OccurrenceDate:                 claim.Data.OccurrenceDate,
					WarningDate:                    claim.Data.WarningDate,
					ThirdPartyClaimDate:            claim.Data.ThirdPartyClaimDate,
					Amount:                         claim.Data.Amount,
					DenialJustificationDescription: claim.Data.DenialJustificationDescription,
					StatusAlterationDate: func() timeutil.BrazilDate {
						if claim.Data.StatusAlterationDate != nil {
							return *claim.Data.StatusAlterationDate
						}
						return timeutil.BrazilDate{}
					}(),
					DenialJustification: func() *InsuranceAcceptanceAndBranchesAbroadClaimDenialJustification {
						if claim.Data.DenialJustification == nil {
							return nil
						}
						denialJust := InsuranceAcceptanceAndBranchesAbroadClaimDenialJustification(*claim.Data.DenialJustification)
						return &denialJust
					}(),
					Coverages: func() []InsuranceAcceptanceAndBranchesAbroadClaimCoverage {
						coverages := make([]InsuranceAcceptanceAndBranchesAbroadClaimCoverage, 0, len(claim.Data.Coverages))
						for _, cov := range claim.Data.Coverages {
							coverages = append(coverages, InsuranceAcceptanceAndBranchesAbroadClaimCoverage{
								InsuredObjectID:     cov.InsuredObjectID,
								Branch:              cov.Branch,
								Code:                InsuranceAcceptanceAndBranchesAbroadClaimCoverageCode(cov.Code),
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

	return GetInsuranceAcceptanceAndBranchesAbroadpolicyIDClaims200JSONResponse{OKResponseInsuranceAcceptanceAndBranchesAbroadClaimsJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
