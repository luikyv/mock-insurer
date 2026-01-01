//go:generate go tool oapi-codegen -config=./config.yml -package=v1 -o=./api_gen.go ./swagger.yml
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
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/patrimonial"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL        string
	service        patrimonial.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(
	host string,
	service patrimonial.Service,
	consentService consent.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:        host + "/open-insurance/insurance-patrimonial/v1",
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	authCodeAuthMiddleware := middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, patrimonial.Scope)
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

	handler = http.HandlerFunc(wrapper.GetInsurancePatrimonial)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeoplePatrimonialRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-patrimonial", handler)

	handler = http.HandlerFunc(wrapper.GetInsurancePatrimonialpolicyIDPolicyInfo)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeoplePatrimonialPolicyInfoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-patrimonial/{policyId}/policy-info", handler)

	handler = http.HandlerFunc(wrapper.GetInsurancePatrimonialpolicyIDPremium)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeoplePatrimonialPremiumRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-patrimonial/{policyId}/premium", handler)

	handler = http.HandlerFunc(wrapper.GetInsurancePatrimonialpolicyIDClaims)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeoplePatrimonialClaimRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-patrimonial/{policyId}/claim", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/insurance-patrimonial/v1", handler), swaggerVersion
}

func (s Server) GetInsurancePatrimonial(ctx context.Context, req GetInsurancePatrimonialRequestObject) (GetInsurancePatrimonialResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	policies, err := s.service.ConsentedPolicies(ctx, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsurancePatrimonial{
		Meta:  *api.NewPaginatedMeta(policies),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-patrimonial", policies),
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
			respPolicies := func() []struct {
				PolicyID    string `json:"policyId"`
				ProductName string `json:"productName"`
			} {
				policiesList := make([]struct {
					PolicyID    string `json:"policyId"`
					ProductName string `json:"productName"`
				}, 0, len(policies.Records))
				for _, policy := range policies.Records {
					policiesList = append(policiesList, struct {
						PolicyID    string `json:"policyId"`
						ProductName string `json:"productName"`
					}{
						PolicyID:    policy.ID,
						ProductName: policy.Data.ProductName,
					})
				}
				return policiesList
			}()

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

	return GetInsurancePatrimonial200JSONResponse{OKResponseInsurancePatrimonialJSONResponse(resp)}, nil
}

func (s Server) GetInsurancePatrimonialpolicyIDPolicyInfo(ctx context.Context, req GetInsurancePatrimonialpolicyIDPolicyInfoRequestObject) (GetInsurancePatrimonialpolicyIDPolicyInfoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, string(req.PolicyID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsurancePatrimonialPolicyInfo{
		Data: func() InsurancePatrimonialPolicyInfo {
			return InsurancePatrimonialPolicyInfo{
				PolicyID:            policy.ID,
				DocumentType:        InsurancePatrimonialPolicyInfoDocumentType(policy.Data.DocumentType),
				SusepProcessNumber:  &policy.Data.SusepProcessNumber,
				GroupCertificateID:  &policy.Data.GroupCertificateID,
				IssuanceType:        InsurancePatrimonialPolicyInfoIssuanceType(policy.Data.IssuanceType),
				IssuanceDate:        policy.Data.IssuanceDate,
				TermStartDate:       policy.Data.TermStartDate,
				TermEndDate:         policy.Data.TermEndDate,
				LeadInsurerCode:     policy.Data.LeadInsurerCode,
				LeadInsurerPolicyID: policy.Data.LeadInsurerPolicyID,
				MaxLMG:              policy.Data.MaxLMG,
				ProposalID:          policy.Data.ProposalID,
				Insureds: func() []PersonalInfo {
					insureds := make([]PersonalInfo, len(policy.Data.Insureds))
					for i, insured := range policy.Data.Insureds {
						insureds[i] = PersonalInfo{
							Identification:           insured.Identification,
							IdentificationType:       PersonalInfoIdentificationType(insured.IdentificationType),
							IdentificationTypeOthers: insured.IdentificationTypeOthers,
							Name:                     insured.Name,
							BirthDate:                insured.BirthDate,
							PostCode:                 insured.PostCode,
							Email:                    insured.Email,
							City:                     insured.City,
							State:                    PersonalInfoState(insured.State),
							Country:                  PersonalInfoCountry(insured.Country),
							Address:                  insured.Address,
							AddressAdditionalInfo:    insured.AddressAdditionalInfo,
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
					for i, inter := range *policy.Data.Intermediaries {
						intermediaries[i] = Intermediary{
							Type:           IntermediaryType(inter.Type),
							TypeOthers:     inter.TypeOthers,
							Identification: inter.Identification,
							BrokerID:       inter.BrokerID,
							IdentificationType: func() *IntermediaryIdentificationType {
								if inter.IdentificationType == nil {
									return nil
								}
								it := IntermediaryIdentificationType(*inter.IdentificationType)
								return &it
							}(),
							IdentificationTypeOthers: inter.IdentificationTypeOthers,
							Name:                     inter.Name,
							PostCode:                 inter.PostCode,
							City:                     inter.City,
							State:                    inter.State,
							Country: func() *IntermediaryCountry {
								if inter.Country == nil {
									return nil
								}
								c := IntermediaryCountry(*inter.Country)
								return &c
							}(),
							Address: inter.Address,
						}
					}
					return &intermediaries
				}(),
				InsuredObjects: func() []InsurancePatrimonialInsuredObject {
					objects := make([]InsurancePatrimonialInsuredObject, len(policy.Data.InsuredObjects))
					for i, obj := range policy.Data.InsuredObjects {
						objects[i] = InsurancePatrimonialInsuredObject{
							Identification:     obj.Identification,
							Type:               InsurancePatrimonialInsuredObjectType(obj.Type),
							TypeAdditionalInfo: obj.TypeAdditionalInfo,
							Description:        obj.Description,
							Amount:             obj.Amount,
							Coverages: func() []InsurancePatrimonialInsuredObjectCoverage {
								coverages := make([]InsurancePatrimonialInsuredObjectCoverage, len(obj.Coverages))
								for j, cov := range obj.Coverages {
									coverages[j] = InsurancePatrimonialInsuredObjectCoverage{
										Branch:             cov.Branch,
										Code:               InsurancePatrimonialCoverageCode(cov.Code),
										Description:        cov.Description,
										InternalCode:       cov.InternalCode,
										SusepProcessNumber: cov.SusepProcessNumber,
										LMI:                cov.LMI,
										TermStartDate:      cov.TermStartDate,
										TermEndDate:        cov.TermEndDate,
										IsMainCoverage:     cov.IsMainCoverage,
										Feature:            InsurancePatrimonialInsuredObjectCoverageFeature(cov.Feature),
										Type:               InsurancePatrimonialInsuredObjectCoverageType(cov.Type),
										GracePeriod:        cov.GracePeriod,
										GracePeriodicity: func() *InsurancePatrimonialInsuredObjectCoverageGracePeriodicity {
											if cov.GracePeriodicity == nil {
												return nil
											}
											gp := InsurancePatrimonialInsuredObjectCoverageGracePeriodicity(*cov.GracePeriodicity)
											return &gp
										}(),
										GracePeriodCountingMethod: func() *InsurancePatrimonialInsuredObjectCoverageGracePeriodCountingMethod {
											if cov.GracePeriodCountingMethod == nil {
												return nil
											}
											gpcm := InsurancePatrimonialInsuredObjectCoverageGracePeriodCountingMethod(*cov.GracePeriodCountingMethod)
											return &gpcm
										}(),
										GracePeriodStartDate:     cov.GracePeriodStartDate,
										GracePeriodEndDate:       cov.GracePeriodEndDate,
										PremiumPeriodicity:       InsurancePatrimonialInsuredObjectCoveragePremiumPeriodicity(cov.PremiumPeriodicity),
										PremiumPeriodicityOthers: cov.PremiumPeriodicityOthers,
									}
								}
								return coverages
							}(),
						}
					}
					return objects
				}(),
				Coverages: func() *[]InsurancePatrimonialCoverage {
					if policy.Data.Coverages == nil {
						return nil
					}
					coverages := make([]InsurancePatrimonialCoverage, len(*policy.Data.Coverages))
					for i, cov := range *policy.Data.Coverages {
						coverages[i] = InsurancePatrimonialCoverage{
							Branch:      cov.Branch,
							Code:        InsurancePatrimonialCoverageCode(cov.Code),
							Description: cov.Description,
							Deductible: func() *Deductible {
								if cov.Deductible == nil {
									return nil
								}
								d := cov.Deductible
								return &Deductible{
									Type:               DeductibleType(d.Type),
									TypeAdditionalInfo: d.TypeAdditionalInfo,
									Amount:             d.Amount,
									Period:             d.Period,
									Periodicity:        DeductiblePeriodicity(d.Periodicity),
									PeriodCountingMethod: func() *DeductiblePeriodCountingMethod {
										if d.PeriodCountingMethod == nil {
											return nil
										}
										pcm := DeductiblePeriodCountingMethod(*d.PeriodCountingMethod)
										return &pcm
									}(),
									PeriodStartDate: d.PeriodStartDate,
									PeriodEndDate:   d.PeriodEndDate,
									Description:     d.Description,
								}
							}(),
							POS: func() *POS {
								if cov.POS == nil {
									return nil
								}
								pos := cov.POS
								return &POS{
									ApplicationType:       POSApplicationType(pos.ApplicationType),
									ApplicationTypeOthers: pos.Description,
									MinValue:              pos.MinValue,
									MaxValue:              pos.MaxValue,
									Percentage:            pos.Percentage,
									ValueOthers:           pos.ValueOthers,
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
				BranchInfo: func() *InsurancePatrimonialSpecificPolicyInfo {
					if policy.Data.BranchInfo == nil {
						return nil
					}
					bi := policy.Data.BranchInfo
					return &InsurancePatrimonialSpecificPolicyInfo{
						BasicCoverageIndex: func() *InsurancePatrimonialSpecificPolicyInfoBasicCoverageIndex {
							if bi.BasicCoverageIndex == nil {
								return nil
							}
							idx := InsurancePatrimonialSpecificPolicyInfoBasicCoverageIndex(bi.BasicCoverageIndex.Index)
							return &idx
						}(),
						InsuredObjects: func() *[]InsurancePatrimonialSpecificInsuredObject {
							if len(bi.InsuredObjects) == 0 {
								return nil
							}
							objects := make([]InsurancePatrimonialSpecificInsuredObject, len(bi.InsuredObjects))
							for i, obj := range bi.InsuredObjects {
								objects[i] = InsurancePatrimonialSpecificInsuredObject{
									Identification: obj.Identification,
									PropertyType: func() *InsurancePatrimonialSpecificInsuredObjectPropertyType {
										if obj.PropertyType == nil {
											return nil
										}
										pt := InsurancePatrimonialSpecificInsuredObjectPropertyType(*obj.PropertyType)
										return &pt
									}(),
									StructuringType: func() *InsurancePatrimonialSpecificInsuredObjectStructuringType {
										if obj.StructuringType == nil {
											return nil
										}
										st := InsurancePatrimonialSpecificInsuredObjectStructuringType(*obj.StructuringType)
										return &st
									}(),
									PostCode:         obj.PostCode,
									BusinessActivity: obj.BusinessActivity,
								}
							}
							return &objects
						}(),
					}
				}(),
			}
		}(),
		Links: *api.NewLinks(s.baseURL + "/insurance-patrimonial/" + string(req.PolicyID) + "/policy-info"),
		Meta:  *api.NewMeta(),
	}

	return GetInsurancePatrimonialpolicyIDPolicyInfo200JSONResponse{OKResponseInsurancePatrimonialPolicyInfoJSONResponse(resp)}, nil
}

func (s Server) GetInsurancePatrimonialpolicyIDPremium(ctx context.Context, req GetInsurancePatrimonialpolicyIDPremiumRequestObject) (GetInsurancePatrimonialpolicyIDPremiumResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, string(req.PolicyID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsurancePatrimonialPremium{
		Data: func() InsurancePatrimonialPremium {
			premium := policy.Data.Premium
			return InsurancePatrimonialPremium{
				PaymentsQuantity: premium.PaymentsQuantity,
				Amount:           premium.Amount,
				Coverages: func() []InsurancePatrimonialPremiumCoverage {
					coverages := make([]InsurancePatrimonialPremiumCoverage, len(premium.Coverages))
					for i, cov := range premium.Coverages {
						coverages[i] = InsurancePatrimonialPremiumCoverage{
							Branch:        cov.Branch,
							Code:          InsurancePatrimonialCoverageCode(cov.Code),
							Description:   cov.Description,
							PremiumAmount: cov.PremiumAmount,
						}
					}
					return coverages
				}(),
				Payments: func() []Payment {
					payments := make([]Payment, len(premium.Payments))
					for i, p := range premium.Payments {
						payments[i] = Payment{
							MovementDate: p.MovementDate,
							MovementType: PaymentMovementType(p.MovementType),
							MovementOrigin: func() *PaymentMovementOrigin {
								if p.MovementOrigin == nil {
									return nil
								}
								mo := PaymentMovementOrigin(*p.MovementOrigin)
								return &mo
							}(),
							MovementPaymentsNumber: p.MovementPaymentsNumber,
							Amount:                 p.Amount,
							MaturityDate:           p.MaturityDate,
							TellerID:               p.TellerID,
							TellerIDType: func() *PaymentTellerIDType {
								if p.TellerIDType == nil {
									return nil
								}
								tt := PaymentTellerIDType(*p.TellerIDType)
								return &tt
							}(),
							TellerName:               p.TellerName,
							FinancialInstitutionCode: p.FinancialInstitutionCode,
							PaymentType: func() *PaymentPaymentType {
								if p.PaymentType == nil {
									return nil
								}
								pt := PaymentPaymentType(*p.PaymentType)
								return &pt
							}(),
							PaymentTypeOthers: p.PaymentTypeOthers,
						}
					}
					return payments
				}(),
			}
		}(),
		Links: *api.NewLinks(s.baseURL + "/insurance-patrimonial/" + string(req.PolicyID) + "/premium"),
		Meta:  *api.NewMeta(),
	}

	return GetInsurancePatrimonialpolicyIDPremium200JSONResponse{OKResponseInsurancePatrimonialPremiumJSONResponse(resp)}, nil
}

func (s Server) GetInsurancePatrimonialpolicyIDClaims(ctx context.Context, req GetInsurancePatrimonialpolicyIDClaimsRequestObject) (GetInsurancePatrimonialpolicyIDClaimsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(nil, nil)
	claims, err := s.service.ConsentedClaims(ctx, string(req.PolicyID), consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsurancePatrimonialClaims{
		Data: func() []InsurancePatrimonialClaim {
			respClaims := make([]InsurancePatrimonialClaim, 0, len(claims.Records))
			for _, claim := range claims.Records {
				respClaims = append(respClaims, InsurancePatrimonialClaim{
					Identification:            claim.Data.Identification,
					DocumentationDeliveryDate: claim.Data.DocumentationDeliveryDate,
					Status:                    InsurancePatrimonialClaimStatus(claim.Data.Status),
					StatusAlterationDate:      claim.Data.StatusAlterationDate,
					OccurrenceDate:            claim.Data.OccurrenceDate,
					WarningDate:               claim.Data.WarningDate,
					ThirdPartyClaimDate:       claim.Data.ThirdPartyClaimDate,
					Amount:                    claim.Data.Amount,
					DenialJustification: func() *InsurancePatrimonialClaimDenialJustification {
						if claim.Data.DenialJustification == nil {
							return nil
						}
						dj := InsurancePatrimonialClaimDenialJustification(*claim.Data.DenialJustification)
						return &dj
					}(),
					DenialJustificationDescription: claim.Data.DenialJustificationDescription,
					Coverages: func() []InsurancePatrimonialClaimCoverage {
						coverages := make([]InsurancePatrimonialClaimCoverage, len(claim.Data.Coverages))
						for i, cov := range claim.Data.Coverages {
							coverages[i] = InsurancePatrimonialClaimCoverage{
								InsuredObjectID:     cov.InsuredObjectID,
								Branch:              cov.Branch,
								Code:                InsurancePatrimonialCoverageCode(cov.Code),
								Description:         cov.Description,
								WarningDate:         cov.WarningDate,
								ThirdPartyClaimDate: cov.ThirdPartyClaimDate,
							}
						}
						return coverages
					}(),
				})
			}
			return respClaims
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-patrimonial/"+string(req.PolicyID)+"/claim", claims),
		Meta:  *api.NewPaginatedMeta(claims),
	}

	return GetInsurancePatrimonialpolicyIDClaims200JSONResponse{OKResponseInsurancePatrimonialClaimsJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
