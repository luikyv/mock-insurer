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
	"github.com/luikyv/mock-insurer/internal/lifepension"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL        string
	service        lifepension.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(
	host string,
	service lifepension.Service,
	consentService consent.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:        host + "/open-insurance/insurance-life-pension/v1",
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	authCodeAuthMiddleware := middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, lifepension.Scope)
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

	handler = http.HandlerFunc(wrapper.GetInsuranceLifePension)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionLifePensionRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-life-pension/contracts", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceLifePensioncertificateIDContractInfo)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionLifePensionContractInfoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-life-pension/{certificateId}/contract-info", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceLifePensioncertificateIDPortabilities)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionLifePensionPortabilitiesRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-life-pension/{certificateId}/portabilities", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceLifePensioncertificateIDWithdrawals)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionLifePensionWithdrawalsRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-life-pension/{certificateId}/withdrawals", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceLifePensioncertificateIDMovements)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionLifePensionMovementsRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-life-pension/{certificateId}/movements", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceLifePensioncertificateIDPeople)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionLifePensionClaim)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-life-pension/{certificateId}/claim", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/insurance-life-pension/v1", handler), swaggerVersion
}

func (s Server) GetInsuranceLifePension(ctx context.Context, req GetInsuranceLifePensionRequestObject) (GetInsuranceLifePensionResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	contracts, err := s.service.ConsentedContracts(ctx, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceLifePension{
		Meta:  *api.NewPaginatedMeta(contracts),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-life-pension/contracts", contracts),
		Data: func() []struct {
			Brand struct {
				Companies []struct {
					CnpjNumber  string `json:"cnpjNumber"`
					CompanyName string `json:"companyName"`
					Contracts   []struct {
						CertificateID string `json:"certificateId"`
						ProductName   string `json:"productName"`
					} `json:"contracts"`
				} `json:"companies"`
				Name string `json:"name"`
			} `json:"brand"`
		} {
			respContracts := func() []struct {
				CertificateID string `json:"certificateId"`
				ProductName   string `json:"productName"`
			} {
				contractsList := make([]struct {
					CertificateID string `json:"certificateId"`
					ProductName   string `json:"productName"`
				}, 0, len(contracts.Records))
				for _, contract := range contracts.Records {
					contractsList = append(contractsList, struct {
						CertificateID string `json:"certificateId"`
						ProductName   string `json:"productName"`
					}{
						CertificateID: contract.ID,
						ProductName:   contract.Data.ProductName,
					})
				}
				return contractsList
			}()

			return []struct {
				Brand struct {
					Companies []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Contracts   []struct {
							CertificateID string `json:"certificateId"`
							ProductName   string `json:"productName"`
						} `json:"contracts"`
					} `json:"companies"`
					Name string `json:"name"`
				} `json:"brand"`
			}{{
				Brand: struct {
					Companies []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Contracts   []struct {
							CertificateID string `json:"certificateId"`
							ProductName   string `json:"productName"`
						} `json:"contracts"`
					} `json:"companies"`
					Name string `json:"name"`
				}{
					Name: insurer.Brand,
					Companies: []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Contracts   []struct {
							CertificateID string `json:"certificateId"`
							ProductName   string `json:"productName"`
						} `json:"contracts"`
					}{{
						CnpjNumber:  insurer.CNPJ,
						CompanyName: insurer.Brand,
						Contracts:   respContracts,
					}},
				},
			}}
		}(),
	}

	return GetInsuranceLifePension200JSONResponse{OKResponseInsuranceLifePensionJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceLifePensioncertificateIDContractInfo(ctx context.Context, req GetInsuranceLifePensioncertificateIDContractInfoRequestObject) (GetInsuranceLifePensioncertificateIDContractInfoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	contract, err := s.service.ConsentedContract(ctx, string(req.CertificateID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceLifePensionContractInfo{
		Data: InsuranceLifePensionContractInfo{
			CertificateID:      contract.ID,
			ProductCode:        contract.Data.ProductCode,
			ProposalID:         contract.Data.ProposalID,
			ContractID:         contract.Data.ContractID,
			ContractingType:    InsuranceLifePensionContractInfoContractingType(contract.Data.ContractingType),
			EffectiveDateStart: contract.Data.EffectiveDateStart,
			EffectiveDateEnd:   contract.Data.EffectiveDateEnd,
			CertificateActive:  contract.Data.CertificateActive,
			ConjugatedPlan:     contract.Data.ConjugatedPlan,
			PlanType: func() *InsuranceLifePensionContractInfoPlanType {
				if contract.Data.PlanType != nil {
					pt := InsuranceLifePensionContractInfoPlanType(*contract.Data.PlanType)
					return &pt
				}
				return nil
			}(),
			Insureds: InsuranceLifePensionDocumentsInsured{
				DocumentType:          InsuranceLifePensionDocumentsInsuredDocumentType(contract.Data.Insured.DocumentType),
				DocumentTypeOthers:    contract.Data.Insured.DocumentTypeOthers,
				DocumentNumber:        contract.Data.Insured.DocumentNumber,
				Name:                  contract.Data.Insured.Name,
				BirthDate:             contract.Data.Insured.BirthDate,
				Gender:                InsuranceLifePensionDocumentsInsuredGender(contract.Data.Insured.Gender),
				PostCode:              contract.Data.Insured.PostCode,
				TownName:              contract.Data.Insured.TownName,
				CountrySubDivision:    EnumCountrySubDivision(contract.Data.Insured.CountrySubDivision),
				CountryCode:           InsuranceLifePensionDocumentsInsuredCountryCode(contract.Data.Insured.CountryCode),
				Address:               contract.Data.Insured.Address,
				AddressAdditionalInfo: contract.Data.Insured.AddressAdditionalInfo,
				Email:                 contract.Data.Insured.Email,
			},
			Beneficiary: func() *[]InsuranceLifePensionDocumentsBeneficiary {
				if contract.Data.Beneficiaries == nil {
					return nil
				}
				beneficiaries := make([]InsuranceLifePensionDocumentsBeneficiary, len(*contract.Data.Beneficiaries))
				for i, b := range *contract.Data.Beneficiaries {
					beneficiaries[i] = InsuranceLifePensionDocumentsBeneficiary{
						DocumentNumber:     b.DocumentNumber,
						DocumentType:       InsuranceLifePensionDocumentsBeneficiaryDocumentType(b.DocumentType),
						DocumentTypeOthers: b.DocumentTypeOthers,
						Name:               b.Name,
						BirthDate:          b.BirthDate,
						Kinship: func() *InsuranceLifePensionDocumentsBeneficiaryKinship {
							if b.Kinship == nil {
								return nil
							}
							kinship := InsuranceLifePensionDocumentsBeneficiaryKinship(*b.Kinship)
							return &kinship
						}(),
						KinshipOthers:           b.KinshipOthers,
						ParticipationPercentage: b.ParticipationPercentage,
					}
				}
				return &beneficiaries
			}(),
			Periodicity:       InsuranceLifePensionContractInfoPeriodicity(contract.Data.Periodicity),
			PeriodicityOthers: contract.Data.PeriodicityOthers,
			TaxRegime:         InsuranceLifePensionContractInfoTaxRegime(contract.Data.TaxRegime),
			Intermediary: func() *InsuranceLifePensionDocumentsIntermediary {
				if contract.Data.Intermediary == nil {
					return nil
				}
				inter := contract.Data.Intermediary
				return &InsuranceLifePensionDocumentsIntermediary{
					Type:           InsuranceLifePensionDocumentsIntermediaryType(inter.Type),
					TypeOthers:     inter.TypeOthers,
					DocumentNumber: inter.DocumentNumber,
					IntermediaryID: inter.IntermediaryID,
					DocumentType: func() *InsuranceLifePensionDocumentsIntermediaryDocumentType {
						if inter.DocumentType == nil {
							return nil
						}
						dt := InsuranceLifePensionDocumentsIntermediaryDocumentType(*inter.DocumentType)
						return &dt
					}(),
					DocumentTypeOthers: inter.DocumentTypeOthers,
					Name:               inter.Name,
					PostCode:           inter.PostCode,
					TownName:           inter.TownName,
					CountrySubDivision: func() *EnumCountrySubDivision {
						if inter.CountrySubDivision == nil {
							return nil
						}
						csd := EnumCountrySubDivision(*inter.CountrySubDivision)
						return &csd
					}(),
					CountryCode: func() *InsuranceLifePensionDocumentsIntermediaryCountryCode {
						if inter.CountryCode == nil {
							return nil
						}
						cc := InsuranceLifePensionDocumentsIntermediaryCountryCode(*inter.CountryCode)
						return &cc
					}(),
					Address:        inter.Address,
					AdditionalInfo: inter.AdditionalInfo,
				}
			}(),
			Suseps: func() []InsuranceLifePensionSuseps {
				suseps := make([]InsuranceLifePensionSuseps, len(contract.Data.Suseps))
				for i, s := range contract.Data.Suseps {
					suseps[i] = InsuranceLifePensionSuseps{
						CoverageCode:                      s.CoverageCode,
						SusepProcessNumber:                s.SusepProcessNumber,
						StructureModality:                 InsuranceLifePensionSusepsStructureModality(s.StructureModality),
						Type:                              InsuranceLifePensionSusepsType(s.Type),
						TypeDetails:                       s.TypeDetails,
						LockedPlan:                        s.LockedPlan,
						QualifiedProposer:                 s.QualifiedProposer,
						BenefitPaymentMethod:              InsuranceLifePensionSusepsBenefitPaymentMethod(s.BenefitPaymentMethod),
						FinancialResultReversal:           s.FinancialResultReversal,
						FinancialResultReversalPercentage: s.FinancialResultReversalPercentage,
						CalculationBasis:                  InsuranceLifePensionSusepsCalculationBasis(s.CalculationBasis),
						FIE: func() []struct {
							FIECNPJ                string        `json:"FIECNPJ"`
							FIEName                string        `json:"FIEName"`
							FIETradeName           string        `json:"FIETradeName"`
							PmbacAmount            AmountDetails `json:"pmbacAmount"`
							ProvisionSurplusAmount AmountDetails `json:"provisionSurplusAmount"`
						} {
							fies := make([]struct {
								FIECNPJ                string        `json:"FIECNPJ"`
								FIEName                string        `json:"FIEName"`
								FIETradeName           string        `json:"FIETradeName"`
								PmbacAmount            AmountDetails `json:"pmbacAmount"`
								ProvisionSurplusAmount AmountDetails `json:"provisionSurplusAmount"`
							}, len(s.FIE))
							for j, fie := range s.FIE {
								fies[j] = struct {
									FIECNPJ                string        `json:"FIECNPJ"`
									FIEName                string        `json:"FIEName"`
									FIETradeName           string        `json:"FIETradeName"`
									PmbacAmount            AmountDetails `json:"pmbacAmount"`
									ProvisionSurplusAmount AmountDetails `json:"provisionSurplusAmount"`
								}{
									FIECNPJ:                fie.FIECNPJ,
									FIEName:                fie.FIEName,
									FIETradeName:           fie.FIETradeName,
									PmbacAmount:            fie.PmbacAmount,
									ProvisionSurplusAmount: fie.ProvisionSurplusAmount,
								}
							}
							return fies
						}(),
						BenefitAmount:     s.BenefitAmount,
						RentsInterestRate: s.RentsInterestRate,
						BiometricTable: func() *InsuranceLifePensionSusepsBiometricTable {
							if s.BiometricTable == nil {
								return nil
							}
							bt := InsuranceLifePensionSusepsBiometricTable(*s.BiometricTable)
							return &bt
						}(),
						PmbacInterestRate: s.PmbacInterestRate,
						PmbacGuaranteePriceIndex: func() *InsuranceLifePensionSusepsPmbacGuaranteePriceIndex {
							if s.PmbacGuaranteePriceIndex == nil {
								return nil
							}
							pgpi := InsuranceLifePensionSusepsPmbacGuaranteePriceIndex(*s.PmbacGuaranteePriceIndex)
							return &pgpi
						}(),
						PmbacGuaranteePriceOthers: s.PmbacGuaranteePriceOthers,
						PmbacIndexLagging:         s.PmbacIndexLagging,
						PdrOrVdrminimalGuaranteeIndex: func() *InsuranceLifePensionSusepsPdrOrVdrminimalGuaranteeIndex {
							if s.PdrOrVdrminimalGuaranteeIndex == nil {
								return nil
							}
							pvgmi := InsuranceLifePensionSusepsPdrOrVdrminimalGuaranteeIndex(*s.PdrOrVdrminimalGuaranteeIndex)
							return &pvgmi
						}(),
						PdrOrVdrminimalGuaranteeOthers:     s.PdrOrVdrminimalGuaranteeOthers,
						PdrOrVdrminimalGuaranteePercentage: s.PdrOrVdrminimalGuaranteePercentage,
						Grace: func() *[]InsuranceLifePensionPlansGrace {
							if s.Grace == nil {
								return nil
							}
							graces := make([]InsuranceLifePensionPlansGrace, len(*s.Grace))
							for k, g := range *s.Grace {
								graces[k] = InsuranceLifePensionPlansGrace{
									GraceType: func() *InsuranceLifePensionPlansGraceGraceType {
										if g.GraceType == nil {
											return nil
										}
										gt := InsuranceLifePensionPlansGraceGraceType(*g.GraceType)
										return &gt
									}(),
									GracePeriod: g.GracePeriod,
									GracePeriodicity: func() *InsuranceLifePensionPlansGraceGracePeriodicity {
										if g.GracePeriodicity == nil {
											return nil
										}
										gp := InsuranceLifePensionPlansGraceGracePeriodicity(*g.GracePeriodicity)
										return &gp
									}(),
									DayIndicator: func() *InsuranceLifePensionPlansGraceDayIndicator {
										if g.DayIndicator == nil {
											return nil
										}
										di := InsuranceLifePensionPlansGraceDayIndicator(*g.DayIndicator)
										return &di
									}(),
									GracePeriodStart:   g.GracePeriodStart,
									GracePeriodEnd:     g.GracePeriodEnd,
									GracePeriodBetween: g.GracePeriodBetween,
									GracePeriodBetweenType: func() *InsuranceLifePensionPlansGraceGracePeriodBetweenType {
										if g.GracePeriodBetweenType == nil {
											return nil
										}
										gpbt := InsuranceLifePensionPlansGraceGracePeriodBetweenType(*g.GracePeriodBetweenType)
										return &gpbt
									}(),
								}
							}
							return &graces
						}(),
					}
				}
				return suseps
			}(),
		},
		Links: *api.NewLinks(s.baseURL + "/insurance-life-pension/" + string(req.CertificateID) + "/contract-info"),
		Meta:  *api.NewMeta(),
	}
	return GetInsuranceLifePensioncertificateIDContractInfo200JSONResponse{OKResponseInsuranceLifePensionContractInfoJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceLifePensioncertificateIDPortabilities(ctx context.Context, req GetInsuranceLifePensioncertificateIDPortabilitiesRequestObject) (GetInsuranceLifePensioncertificateIDPortabilitiesResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	portabilities, err := s.service.ConsentedPortabilities(ctx, req.CertificateID, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceLifePensionPortabilities{
		Data: InsuranceLifePensionPortability{
			HasOccurredPortability: len(portabilities.Records) > 0,
			PortabilityInfo: func() *[]struct {
				FIE *[]struct {
					FIECNPJ      *string                                                     `json:"FIECNPJ,omitempty"`
					FIEName      *string                                                     `json:"FIEName,omitempty"`
					FIETradeName *string                                                     `json:"FIETradeName,omitempty"`
					PortedType   InsuranceLifePensionPortabilityPortabilityInfoFIEPortedType `json:"portedType"`
				} `json:"FIE,omitempty"`
				Amount              AmountDetails                                            `json:"amount"`
				Direction           InsuranceLifePensionPortabilityPortabilityInfoDirection  `json:"direction"`
				LiquidationDate     timeutil.DateTime                                        `json:"liquidationDate"`
				PostedChargedAmount AmountDetails                                            `json:"postedChargedAmount"`
				RequestDate         timeutil.DateTime                                        `json:"requestDate"`
				SourceEntity        *string                                                  `json:"sourceEntity,omitempty"`
				SusepProcess        *string                                                  `json:"susepProcess,omitempty"`
				TargetEntity        *string                                                  `json:"targetEntity,omitempty"`
				TaxRegime           *InsuranceLifePensionPortabilityPortabilityInfoTaxRegime `json:"taxRegime,omitempty"`
				Type                *InsuranceLifePensionPortabilityPortabilityInfoType      `json:"type,omitempty"`
			} {
				if len(portabilities.Records) == 0 {
					return nil
				}
				info := make([]struct {
					FIE *[]struct {
						FIECNPJ      *string                                                     `json:"FIECNPJ,omitempty"`
						FIEName      *string                                                     `json:"FIEName,omitempty"`
						FIETradeName *string                                                     `json:"FIETradeName,omitempty"`
						PortedType   InsuranceLifePensionPortabilityPortabilityInfoFIEPortedType `json:"portedType"`
					} `json:"FIE,omitempty"`
					Amount              AmountDetails                                            `json:"amount"`
					Direction           InsuranceLifePensionPortabilityPortabilityInfoDirection  `json:"direction"`
					LiquidationDate     timeutil.DateTime                                        `json:"liquidationDate"`
					PostedChargedAmount AmountDetails                                            `json:"postedChargedAmount"`
					RequestDate         timeutil.DateTime                                        `json:"requestDate"`
					SourceEntity        *string                                                  `json:"sourceEntity,omitempty"`
					SusepProcess        *string                                                  `json:"susepProcess,omitempty"`
					TargetEntity        *string                                                  `json:"targetEntity,omitempty"`
					TaxRegime           *InsuranceLifePensionPortabilityPortabilityInfoTaxRegime `json:"taxRegime,omitempty"`
					Type                *InsuranceLifePensionPortabilityPortabilityInfoType      `json:"type,omitempty"`
				}, len(portabilities.Records))
				for i, p := range portabilities.Records {
					info[i] = struct {
						FIE *[]struct {
							FIECNPJ      *string                                                     `json:"FIECNPJ,omitempty"`
							FIEName      *string                                                     `json:"FIEName,omitempty"`
							FIETradeName *string                                                     `json:"FIETradeName,omitempty"`
							PortedType   InsuranceLifePensionPortabilityPortabilityInfoFIEPortedType `json:"portedType"`
						} `json:"FIE,omitempty"`
						Amount              AmountDetails                                            `json:"amount"`
						Direction           InsuranceLifePensionPortabilityPortabilityInfoDirection  `json:"direction"`
						LiquidationDate     timeutil.DateTime                                        `json:"liquidationDate"`
						PostedChargedAmount AmountDetails                                            `json:"postedChargedAmount"`
						RequestDate         timeutil.DateTime                                        `json:"requestDate"`
						SourceEntity        *string                                                  `json:"sourceEntity,omitempty"`
						SusepProcess        *string                                                  `json:"susepProcess,omitempty"`
						TargetEntity        *string                                                  `json:"targetEntity,omitempty"`
						TaxRegime           *InsuranceLifePensionPortabilityPortabilityInfoTaxRegime `json:"taxRegime,omitempty"`
						Type                *InsuranceLifePensionPortabilityPortabilityInfoType      `json:"type,omitempty"`
					}{
						Amount:              p.Data.PortabilityAmount,
						Direction:           InsuranceLifePensionPortabilityPortabilityInfoDirection(p.Data.Direction),
						LiquidationDate:     p.Data.StatusDate.DateTime(),
						PostedChargedAmount: p.Data.PostedChargedAmount,
						RequestDate:         p.Data.PortabilityDate.DateTime(),
						SourceEntity:        &p.Data.SourceInstitution,
						TargetEntity:        &p.Data.DestinationInstitution,
						SusepProcess:        p.Data.SusepProcess,
						TaxRegime: func() *InsuranceLifePensionPortabilityPortabilityInfoTaxRegime {
							if p.Data.TaxRegime == nil {
								return nil
							}
							tr := InsuranceLifePensionPortabilityPortabilityInfoTaxRegime(*p.Data.TaxRegime)
							return &tr
						}(),
						Type: func() *InsuranceLifePensionPortabilityPortabilityInfoType {
							if p.Data.Type == nil {
								return nil
							}
							t := InsuranceLifePensionPortabilityPortabilityInfoType(*p.Data.Type)
							return &t
						}(),
						FIE: func() *[]struct {
							FIECNPJ      *string                                                     `json:"FIECNPJ,omitempty"`
							FIEName      *string                                                     `json:"FIEName,omitempty"`
							FIETradeName *string                                                     `json:"FIETradeName,omitempty"`
							PortedType   InsuranceLifePensionPortabilityPortabilityInfoFIEPortedType `json:"portedType"`
						} {
							if p.Data.FIE == nil {
								return nil
							}
							fies := make([]struct {
								FIECNPJ      *string                                                     `json:"FIECNPJ,omitempty"`
								FIEName      *string                                                     `json:"FIEName,omitempty"`
								FIETradeName *string                                                     `json:"FIETradeName,omitempty"`
								PortedType   InsuranceLifePensionPortabilityPortabilityInfoFIEPortedType `json:"portedType"`
							}, len(*p.Data.FIE))
							for j, fie := range *p.Data.FIE {
								fies[j] = struct {
									FIECNPJ      *string                                                     `json:"FIECNPJ,omitempty"`
									FIEName      *string                                                     `json:"FIEName,omitempty"`
									FIETradeName *string                                                     `json:"FIETradeName,omitempty"`
									PortedType   InsuranceLifePensionPortabilityPortabilityInfoFIEPortedType `json:"portedType"`
								}{
									FIECNPJ:      &fie.FIECNPJ,
									FIEName:      &fie.FIEName,
									FIETradeName: &fie.FIETradeName,
									PortedType:   InsuranceLifePensionPortabilityPortabilityInfoFIEPortedType(fie.PortedType),
								}
							}
							return &fies
						}(),
					}
				}
				return &info
			}(),
		},
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-life-pension/"+req.CertificateID+"/portabilities", portabilities),
		Meta:  *api.NewPaginatedMeta(portabilities),
	}

	return GetInsuranceLifePensioncertificateIDPortabilities200JSONResponse{OKResponseInsuranceLifePensionPortabilitiesJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceLifePensioncertificateIDWithdrawals(ctx context.Context, req GetInsuranceLifePensioncertificateIDWithdrawalsRequestObject) (GetInsuranceLifePensioncertificateIDWithdrawalsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	withdrawals, err := s.service.ConsentedWithdrawals(ctx, string(req.CertificateID), consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceLifePensionWithdrawal{
		Data: func() []InsuranceLifePensionWithdrawal {
			respWithdrawals := make([]InsuranceLifePensionWithdrawal, 0, len(withdrawals.Records))
			for _, withdrawal := range withdrawals.Records {
				respWithdrawals = append(respWithdrawals, InsuranceLifePensionWithdrawal{
					WithdrawalOccurence: withdrawal.Data.WithdrawalOccurence,
					Amount:              withdrawal.Data.Amount,
					LiquidationDate:     withdrawal.Data.LiquidationDate,
					RequestDate:         withdrawal.Data.RequestDate,
					Type: func() *InsuranceLifePensionWithdrawalType {
						if withdrawal.Data.Type == nil {
							return nil
						}
						wt := InsuranceLifePensionWithdrawalType(*withdrawal.Data.Type)
						return &wt
					}(),
					PostedChargedAmount: withdrawal.Data.PostedChargedAmount,
					Nature: func() *InsuranceLifePensionWithdrawalNature {
						if withdrawal.Data.Nature == nil {
							return nil
						}
						n := InsuranceLifePensionWithdrawalNature(*withdrawal.Data.Nature)
						return &n
					}(),
					FIE: func() *[]struct {
						FIECNPJ      *string `json:"FIECNPJ,omitempty"`
						FIEName      *string `json:"FIEName,omitempty"`
						FIETradeName *string `json:"FIETradeName,omitempty"`
					} {
						if withdrawal.Data.FIE == nil {
							return nil
						}
						fies := make([]struct {
							FIECNPJ      *string `json:"FIECNPJ,omitempty"`
							FIEName      *string `json:"FIEName,omitempty"`
							FIETradeName *string `json:"FIETradeName,omitempty"`
						}, len(*withdrawal.Data.FIE))
						for j, fie := range *withdrawal.Data.FIE {
							fies[j] = struct {
								FIECNPJ      *string `json:"FIECNPJ,omitempty"`
								FIEName      *string `json:"FIEName,omitempty"`
								FIETradeName *string `json:"FIETradeName,omitempty"`
							}{
								FIECNPJ:      fie.FIECNPJ,
								FIEName:      fie.FIEName,
								FIETradeName: fie.FIETradeName,
							}
						}
						return &fies
					}(),
				})
			}
			return respWithdrawals
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-life-pension/"+string(req.CertificateID)+"/withdrawals", withdrawals),
		Meta:  *api.NewPaginatedMeta(withdrawals),
	}
	return GetInsuranceLifePensioncertificateIDWithdrawals200JSONResponse{OKResponseInsuranceLifePensionWithdrawalJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceLifePensioncertificateIDPeople(ctx context.Context, req GetInsuranceLifePensioncertificateIDPeopleRequestObject) (GetInsuranceLifePensioncertificateIDPeopleResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	claims, err := s.service.ConsentedClaims(ctx, string(req.CertificateID), consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceLifePensionClaim{
		Data: func() []InsuranceLifePensionClaim {
			respClaims := make([]InsuranceLifePensionClaim, 0, len(claims.Records))
			for _, claim := range claims.Records {
				respClaims = append(respClaims, InsuranceLifePensionClaim{
					EventInfo: struct {
						EventAlertDate    timeutil.BrazilDate                           `json:"eventAlertDate"`
						EventRegisterDate timeutil.BrazilDate                           `json:"eventRegisterDate"`
						EventStatus       InsuranceLifePensionClaimEventInfoEventStatus `json:"eventStatus"`
					}{
						EventAlertDate:    claim.Data.EventInfo.EventAlertDate,
						EventRegisterDate: claim.Data.EventInfo.EventRegisterDate,
						EventStatus:       InsuranceLifePensionClaimEventInfoEventStatus(claim.Data.EventInfo.EventStatus),
					},
					IncomeInfo: func() *struct {
						BeneficiaryBirthDate     timeutil.BrazilDate                                        `json:"beneficiaryBirthDate"`
						BeneficiaryCategory      InsuranceLifePensionClaimIncomeInfoBeneficiaryCategory     `json:"beneficiaryCategory"`
						BeneficiaryDocTypeOthers *string                                                    `json:"beneficiaryDocTypeOthers,omitempty"`
						BeneficiaryDocument      string                                                     `json:"beneficiaryDocument"`
						BeneficiaryDocumentType  InsuranceLifePensionClaimIncomeInfoBeneficiaryDocumentType `json:"beneficiaryDocumentType"`
						BeneficiaryName          string                                                     `json:"beneficiaryName"`
						BenefitAmount            int                                                        `json:"benefitAmount"`
						DefermentDueDate         *timeutil.BrazilDate                                       `json:"defermentDueDate,omitempty"`
						GrantedDate              timeutil.BrazilDate                                        `json:"grantedDate"`
						IncomeAmount             AmountDetails                                              `json:"incomeAmount"`
						IncomeType               InsuranceLifePensionClaimIncomeInfoIncomeType              `json:"incomeType"`
						IncomeTypeDetails        *string                                                    `json:"incomeTypeDetails,omitempty"`
						LastUpdateDate           timeutil.BrazilDate                                        `json:"lastUpdateDate"`
						MonetaryUpdIndexOthers   *string                                                    `json:"monetaryUpdIndexOthers,omitempty"`
						MonetaryUpdateIndex      InsuranceLifePensionClaimIncomeInfoMonetaryUpdateIndex     `json:"monetaryUpdateIndex"`
						PaymentTerms             *string                                                    `json:"paymentTerms,omitempty"`
						ReversedIncome           *bool                                                      `json:"reversedIncome,omitempty"`
					} {
						if claim.Data.IncomeInfo == nil {
							return nil
						}
						incomeInfo := claim.Data.IncomeInfo
						return &struct {
							BeneficiaryBirthDate     timeutil.BrazilDate                                        `json:"beneficiaryBirthDate"`
							BeneficiaryCategory      InsuranceLifePensionClaimIncomeInfoBeneficiaryCategory     `json:"beneficiaryCategory"`
							BeneficiaryDocTypeOthers *string                                                    `json:"beneficiaryDocTypeOthers,omitempty"`
							BeneficiaryDocument      string                                                     `json:"beneficiaryDocument"`
							BeneficiaryDocumentType  InsuranceLifePensionClaimIncomeInfoBeneficiaryDocumentType `json:"beneficiaryDocumentType"`
							BeneficiaryName          string                                                     `json:"beneficiaryName"`
							BenefitAmount            int                                                        `json:"benefitAmount"`
							DefermentDueDate         *timeutil.BrazilDate                                       `json:"defermentDueDate,omitempty"`
							GrantedDate              timeutil.BrazilDate                                        `json:"grantedDate"`
							IncomeAmount             AmountDetails                                              `json:"incomeAmount"`
							IncomeType               InsuranceLifePensionClaimIncomeInfoIncomeType              `json:"incomeType"`
							IncomeTypeDetails        *string                                                    `json:"incomeTypeDetails,omitempty"`
							LastUpdateDate           timeutil.BrazilDate                                        `json:"lastUpdateDate"`
							MonetaryUpdIndexOthers   *string                                                    `json:"monetaryUpdIndexOthers,omitempty"`
							MonetaryUpdateIndex      InsuranceLifePensionClaimIncomeInfoMonetaryUpdateIndex     `json:"monetaryUpdateIndex"`
							PaymentTerms             *string                                                    `json:"paymentTerms,omitempty"`
							ReversedIncome           *bool                                                      `json:"reversedIncome,omitempty"`
						}{
							BeneficiaryBirthDate:     incomeInfo.BeneficiaryBirthDate,
							BeneficiaryCategory:      InsuranceLifePensionClaimIncomeInfoBeneficiaryCategory(incomeInfo.BeneficiaryCategory),
							BeneficiaryDocTypeOthers: incomeInfo.BeneficiaryDocTypeOthers,
							BeneficiaryDocument:      incomeInfo.BeneficiaryDocument,
							BeneficiaryDocumentType:  InsuranceLifePensionClaimIncomeInfoBeneficiaryDocumentType(incomeInfo.BeneficiaryDocumentType),
							BeneficiaryName:          incomeInfo.BeneficiaryName,
							BenefitAmount:            incomeInfo.BenefitAmount,
							DefermentDueDate:         incomeInfo.DefermentDueDate,
							GrantedDate:              incomeInfo.GrantedDate,
							IncomeAmount:             incomeInfo.IncomeAmount,
							IncomeType:               InsuranceLifePensionClaimIncomeInfoIncomeType(incomeInfo.IncomeType),
							IncomeTypeDetails:        incomeInfo.IncomeTypeDetails,
							LastUpdateDate:           incomeInfo.LastUpdateDate,
							MonetaryUpdIndexOthers:   incomeInfo.MonetaryUpdIndexOthers,
							MonetaryUpdateIndex:      InsuranceLifePensionClaimIncomeInfoMonetaryUpdateIndex(incomeInfo.MonetaryUpdateIndex),
							PaymentTerms:             incomeInfo.PaymentTerms,
							ReversedIncome:           incomeInfo.ReversedIncome,
						}
					}(),
				})
			}
			return respClaims
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-life-pension/"+string(req.CertificateID)+"/claim", claims),
		Meta:  *api.NewPaginatedMeta(claims),
	}

	return GetInsuranceLifePensioncertificateIDPeople200JSONResponse{OKResponseInsuranceLifePensionClaimJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceLifePensioncertificateIDMovements(ctx context.Context, req GetInsuranceLifePensioncertificateIDMovementsRequestObject) (GetInsuranceLifePensioncertificateIDMovementsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	contract, err := s.service.ConsentedContract(ctx, string(req.CertificateID), consentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceLifePensionMovements{
		Data: InsuranceLifePensionMovements{
			MovementContributions: func() *[]struct {
				ChargedInAdvanceAmount     AmountDetails                                                 `json:"chargedInAdvanceAmount"`
				ContributionAmount         AmountDetails                                                 `json:"contributionAmount"`
				ContributionExpirationDate timeutil.BrazilDate                                           `json:"contributionExpirationDate"`
				ContributionPaymentDate    timeutil.BrazilDate                                           `json:"contributionPaymentDate"`
				Periodicity                InsuranceLifePensionMovementsMovementContributionsPeriodicity `json:"periodicity"`
				PeriodicityOthers          *string                                                       `json:"periodicityOthers,omitempty"`
			} {
				if len(contract.Data.MovementContributions) == 0 {
					return nil
				}
				contributions := make([]struct {
					ChargedInAdvanceAmount     AmountDetails                                                 `json:"chargedInAdvanceAmount"`
					ContributionAmount         AmountDetails                                                 `json:"contributionAmount"`
					ContributionExpirationDate timeutil.BrazilDate                                           `json:"contributionExpirationDate"`
					ContributionPaymentDate    timeutil.BrazilDate                                           `json:"contributionPaymentDate"`
					Periodicity                InsuranceLifePensionMovementsMovementContributionsPeriodicity `json:"periodicity"`
					PeriodicityOthers          *string                                                       `json:"periodicityOthers,omitempty"`
				}, 0, len(contract.Data.MovementContributions))
				for _, m := range contract.Data.MovementContributions {
					contributions = append(contributions, struct {
						ChargedInAdvanceAmount     AmountDetails                                                 `json:"chargedInAdvanceAmount"`
						ContributionAmount         AmountDetails                                                 `json:"contributionAmount"`
						ContributionExpirationDate timeutil.BrazilDate                                           `json:"contributionExpirationDate"`
						ContributionPaymentDate    timeutil.BrazilDate                                           `json:"contributionPaymentDate"`
						Periodicity                InsuranceLifePensionMovementsMovementContributionsPeriodicity `json:"periodicity"`
						PeriodicityOthers          *string                                                       `json:"periodicityOthers,omitempty"`
					}{
						ContributionAmount:         m.Amount,
						ContributionPaymentDate:    m.PaymentDate,
						ContributionExpirationDate: m.ExpirationDate,
						ChargedInAdvanceAmount:     m.ChargedInAdvanceAmount,
						Periodicity:                InsuranceLifePensionMovementsMovementContributionsPeriodicity(m.Periodicity),
						PeriodicityOthers:          m.PeriodicityOthers,
					})
				}
				return &contributions
			}(),
			MovementBenefits: func() *[]struct {
				BenefitAmount      AmountDetails       `json:"benefitAmount"`
				BenefitPaymentDate timeutil.BrazilDate `json:"benefitPaymentDate"`
			} {
				if len(contract.Data.MovementBenefits) == 0 {
					return nil
				}
				benefits := make([]struct {
					BenefitAmount      AmountDetails       `json:"benefitAmount"`
					BenefitPaymentDate timeutil.BrazilDate `json:"benefitPaymentDate"`
				}, 0, len(contract.Data.MovementBenefits))
				for _, m := range contract.Data.MovementBenefits {
					benefits = append(benefits, struct {
						BenefitAmount      AmountDetails       `json:"benefitAmount"`
						BenefitPaymentDate timeutil.BrazilDate `json:"benefitPaymentDate"`
					}{
						BenefitAmount:      m.Amount,
						BenefitPaymentDate: m.PaymentDate,
					})
				}
				return &benefits
			}(),
		},
		Links: *api.NewLinks(s.baseURL + "/insurance-life-pension/" + req.CertificateID + "/movements"),
		Meta: func() api.Meta {
			movements := make([]any, 0, len(contract.Data.MovementContributions)+len(contract.Data.MovementBenefits))
			for _, m := range contract.Data.MovementContributions {
				movements = append(movements, m)
			}
			for _, m := range contract.Data.MovementBenefits {
				movements = append(movements, m)
			}
			meta := api.NewPaginatedMeta(page.New(movements, page.NewPagination(nil, nil), len(movements)))
			return *meta
		}(),
	}

	return GetInsuranceLifePensioncertificateIDMovements200JSONResponse{OKResponseInsuranceLifePensionMovementsJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
