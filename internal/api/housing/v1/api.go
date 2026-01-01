//go:generate go tool oapi-codegen -config=./config.yml -package=v1 -o=./api_gen.go ./swagger.yml
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
	"github.com/luikyv/mock-insurer/internal/housing"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL        string
	service        housing.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(
	host string,
	service housing.Service,
	consentService consent.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:        host + "/open-insurance/insurance-housing/v1",
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	authCodeAuthMiddleware := middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, housing.Scope)
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

	handler = http.HandlerFunc(wrapper.GetInsuranceHousing)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleHousingRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-housing", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceHousingpolicyIDPolicyInfo)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleHousingPolicyInfoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-housing/{policyId}/policy-info", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceHousingpolicyIDPremium)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleHousingPremiumRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-housing/{policyId}/premium", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceHousingpolicyIDClaims)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleHousingClaimRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-housing/{policyId}/claim", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/insurance-housing/v1", handler), swaggerVersion
}

func (s Server) GetInsuranceHousing(ctx context.Context, req GetInsuranceHousingRequestObject) (GetInsuranceHousingResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	policies, err := s.service.ConsentedPolicies(ctx, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceHousing{
		Meta:  *api.NewPaginatedMeta(policies),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-housing", policies),
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

	return GetInsuranceHousing200JSONResponse{OKResponseInsuranceHousingJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceHousingpolicyIDPolicyInfo(ctx context.Context, req GetInsuranceHousingpolicyIDPolicyInfoRequestObject) (GetInsuranceHousingpolicyIDPolicyInfoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, string(req.PolicyID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	policyInfo := InsuranceHousingPolicyInfo{
		PolicyID:            policy.ID,
		DocumentType:        InsuranceHousingPolicyInfoDocumentType(policy.Data.DocumentType),
		IssuanceType:        InsuranceHousingPolicyInfoIssuanceType(policy.Data.IssuanceType),
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
					BirthDate: func() timeutil.BrazilDate {
						if ins.BirthDate != nil {
							return *ins.BirthDate
						}
						return timeutil.BrazilDate{}
					}(),
					Email:   ins.Email,
					City:    ins.City,
					State:   PersonalInfoState(ins.State),
					Country: PersonalInfoCountry(ins.Country),
					Address: ins.Address,
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
					Country: func() *IntermediaryCountry {
						if in.Country == nil {
							return nil
						}
						country := IntermediaryCountry(*in.Country)
						return &country
					}(),
					Address: in.Address,
				}
			}
			return &intermediaries
		}(),
		InsuredObjects: func() []InsuranceHousingInsuredObject {
			insuredObjects := make([]InsuranceHousingInsuredObject, len(policy.Data.InsuredObjects))
			for i, obj := range policy.Data.InsuredObjects {
				insuredObjects[i] = InsuranceHousingInsuredObject{
					Identification:     obj.Identification,
					Type:               InsuranceHousingInsuredObjectType(obj.Type),
					TypeAdditionalInfo: obj.TypeAdditionalInfo,
					Description:        obj.Description,
					Amount:             obj.Amount,
					Coverages: func() []InsuranceHousingInsuredObjectCoverage {
						coverages := make([]InsuranceHousingInsuredObjectCoverage, len(obj.Coverages))
						for j, cov := range obj.Coverages {
							coverages[j] = InsuranceHousingInsuredObjectCoverage{
								Branch:             cov.Branch,
								Code:               InsuranceHousingInsuredObjectCoverageCode(cov.Code),
								Description:        cov.Description,
								InternalCode:       cov.InternalCode,
								SusepProcessNumber: cov.SusepProcessNumber,
								LMI:                cov.LMI,
								TermStartDate:      cov.TermStartDate,
								TermEndDate:        cov.TermEndDate,
								IsMainCoverage:     cov.IsMainCoverage,
								Feature:            InsuranceHousingInsuredObjectCoverageFeature(cov.Feature),
								Type:               InsuranceHousingInsuredObjectCoverageType(cov.Type),
								GracePeriod:        cov.GracePeriod,
								GracePeriodicity: func() *InsuranceHousingInsuredObjectCoverageGracePeriodicity {
									if cov.GracePeriodicity == nil {
										return nil
									}
									periodicity := InsuranceHousingInsuredObjectCoverageGracePeriodicity(*cov.GracePeriodicity)
									return &periodicity
								}(),
								GracePeriodCountingMethod: func() *InsuranceHousingInsuredObjectCoverageGracePeriodCountingMethod {
									if cov.GracePeriodCountingMethod == nil {
										return nil
									}
									method := InsuranceHousingInsuredObjectCoverageGracePeriodCountingMethod(*cov.GracePeriodCountingMethod)
									return &method
								}(),
								GracePeriodStartDate:     cov.GracePeriodStartDate,
								GracePeriodEndDate:       cov.GracePeriodEndDate,
								IsLMISublimit:            cov.IsLMISublimit,
								PremiumPeriodicity:       InsuranceHousingInsuredObjectCoveragePremiumPeriodicity(cov.PremiumPeriodicity),
								PremiumPeriodicityOthers: cov.PremiumPeriodicityOthers,
							}
						}
						return coverages
					}(),
				}
			}
			return insuredObjects
		}(),
		Coverages: func() *[]InsuranceHousingCoverage {
			if policy.Data.Coverages == nil {
				return nil
			}
			coverages := make([]InsuranceHousingCoverage, len(*policy.Data.Coverages))
			for i, cov := range *policy.Data.Coverages {
				coverages[i] = InsuranceHousingCoverage{
					Branch:      cov.Branch,
					Code:        InsuranceHousingCoverageCode(cov.Code),
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
		BranchInfo: func() InsuranceHousingSpecificPolicyInfo {
			return InsuranceHousingSpecificPolicyInfo{
				InsuredObjects: func() []InsuranceHousingSpecificInsuredObject {
					insuredObjects := make([]InsuranceHousingSpecificInsuredObject, len(policy.Data.BranchInfo.InsuredObjects))
					for i, obj := range policy.Data.BranchInfo.InsuredObjects {
						insuredObjects[i] = InsuranceHousingSpecificInsuredObject{
							CostRate:       obj.CostRate,
							Identification: obj.Identification,
							InterestRate:   obj.InterestRate,
							Lenders: func() []struct {
								CnpjNumber  string `json:"cnpjNumber"`
								CompanyName string `json:"companyName"`
							} {
								lenders := make([]struct {
									CnpjNumber  string `json:"cnpjNumber"`
									CompanyName string `json:"companyName"`
								}, len(obj.Lenders))
								for j, l := range obj.Lenders {
									lenders[j] = struct {
										CnpjNumber  string `json:"cnpjNumber"`
										CompanyName string `json:"companyName"`
									}{
										CnpjNumber:  l.CnpjNumber,
										CompanyName: l.CompanyName,
									}
								}
								return lenders
							}(),
							PostCode:                   obj.PostCode,
							PropertyType:               InsuranceHousingSpecificInsuredObjectPropertyType(obj.PropertyType),
							PropertyTypeAdditionalInfo: obj.PropertyTypeAdditionalInfo,
							UpdateIndex:                InsuranceHousingSpecificInsuredObjectUpdateIndex(obj.UpdateIndex),
							UpdateIndexOthers:          obj.UpdateIndexOthers,
						}
					}
					return insuredObjects
				}(),
				Insureds: func() []InsuranceHousingSpecificInsured {
					insureds := make([]InsuranceHousingSpecificInsured, len(policy.Data.BranchInfo.Insureds))
					for i, ins := range policy.Data.BranchInfo.Insureds {
						insureds[i] = InsuranceHousingSpecificInsured{
							BirthDate:                ins.BirthDate,
							Identification:           ins.Identification,
							IdentificationType:       InsuranceHousingSpecificInsuredIdentificationType(ins.IdentificationType),
							IdentificationTypeOthers: ins.IdentificationTypeOthers,
						}
					}
					return insureds
				}(),
			}
		}(),
	}

	resp := ResponseInsuranceHousingPolicyInfo{
		Data:  policyInfo,
		Links: *api.NewLinks(s.baseURL + "/insurance-housing/" + string(req.PolicyID) + "/policy-info"),
		Meta:  *api.NewMeta(),
	}

	return GetInsuranceHousingpolicyIDPolicyInfo200JSONResponse{OKResponseInsuranceHousingPolicyInfoJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceHousingpolicyIDPremium(ctx context.Context, req GetInsuranceHousingpolicyIDPremiumRequestObject) (GetInsuranceHousingpolicyIDPremiumResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, string(req.PolicyID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	premium := InsuranceHousingPremium{
		PaymentsQuantity: policy.Data.Premium.PaymentsQuantity,
		Amount:           policy.Data.Premium.Amount,
		Coverages: func() []InsuranceHousingPremiumCoverage {
			coverages := make([]InsuranceHousingPremiumCoverage, len(policy.Data.Premium.Coverages))
			for i, cov := range policy.Data.Premium.Coverages {
				coverages[i] = InsuranceHousingPremiumCoverage{
					Branch:        cov.Branch,
					Code:          InsuranceHousingPremiumCoverageCode(cov.Code),
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

	resp := ResponseInsuranceHousingPremium{
		Data:  premium,
		Links: *api.NewLinks(s.baseURL + "/insurance-housing/" + string(req.PolicyID) + "/premium"),
		Meta:  *api.NewMeta(),
	}

	return GetInsuranceHousingpolicyIDPremium200JSONResponse{OKResponseInsuranceHousingPremiumJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceHousingpolicyIDClaims(ctx context.Context, req GetInsuranceHousingpolicyIDClaimsRequestObject) (GetInsuranceHousingpolicyIDClaimsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	claims, err := s.service.ConsentedClaims(ctx, string(req.PolicyID), consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceHousingClaims{
		Meta:  *api.NewPaginatedMeta(claims),
		Links: *api.NewPaginatedLinks(fmt.Sprintf("%s/insurance-housing/%s/claim", s.baseURL, req.PolicyID), claims),
		Data: func() []InsuranceHousingClaim {
			respClaims := make([]InsuranceHousingClaim, 0, len(claims.Records))
			for _, claim := range claims.Records {
				respClaims = append(respClaims, InsuranceHousingClaim{
					Identification:            claim.Data.Identification,
					DocumentationDeliveryDate: claim.Data.DocumentationDeliveryDate,
					Status:                    InsuranceHousingClaimStatus(claim.Data.Status),
					StatusAlterationDate:      claim.Data.StatusAlterationDate,
					OccurrenceDate:            claim.Data.OccurrenceDate,
					WarningDate:               claim.Data.WarningDate,
					ThirdPartyClaimDate:       claim.Data.ThirdPartyClaimDate,
					Amount:                    claim.Data.Amount,
					DenialJustification: func() *InsuranceHousingClaimDenialJustification {
						if claim.Data.DenialJustification == nil {
							return nil
						}
						denialJust := InsuranceHousingClaimDenialJustification(*claim.Data.DenialJustification)
						return &denialJust
					}(),
					DenialJustificationDescription: claim.Data.DenialJustificationDescription,
					Coverages: func() []InsuranceHousingClaimCoverage {
						coverages := make([]InsuranceHousingClaimCoverage, 0, len(claim.Data.Coverages))
						for _, cov := range claim.Data.Coverages {
							coverages = append(coverages, InsuranceHousingClaimCoverage{
								InsuredObjectID:     cov.InsuredObjectID,
								Branch:              cov.Branch,
								Code:                InsuranceHousingClaimCoverageCode(cov.Code),
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

	return GetInsuranceHousingpolicyIDClaims200JSONResponse{OKResponseInsuranceHousingClaimsJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
