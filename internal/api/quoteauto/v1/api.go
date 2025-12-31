//go:generate oapi-codegen -config=./config.yml -package=v1 -o=./api_gen.go ./swagger.yml
package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/auto"
	"github.com/luikyv/mock-insurer/internal/customer"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/idempotency"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/quote"
	quoteauto "github.com/luikyv/mock-insurer/internal/quote/auto"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL            string
	service            quoteauto.Service
	idempotencyService idempotency.Service
	op                 *provider.Provider
}

func NewServer(
	host string,
	service quoteauto.Service,
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

	clientCredentialsMiddleware := middleware.Auth(s.op, goidc.GrantClientCredentials, quoteauto.Scope)
	clientCredentialsLeadMiddleware := middleware.Auth(s.op, goidc.GrantClientCredentials, quoteauto.ScopeLead)
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

	handler = http.HandlerFunc(wrapper.PatchQuoteAutoLead)
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
	lead := quoteauto.Lead{
		ConsentID: req.Body.Data.ConsentID,
		OrgID:     orgID,
		Data: quoteauto.LeadData{
			ExpirationDateTime: req.Body.Data.ExpirationDateTime,
			Customer: quote.Customer{
				Personal: func() *quote.PersonalData {
					identificationData, err := req.Body.Data.QuoteCustomer.IdentificationData.AsPersonalIdentificationData()
					if err != nil {
						return nil
					}
					return &quote.PersonalData{
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
							Contact: customer.PersonalContact{
								PostalAddresses: func() []customer.PersonalPostalAddress {
									addresses := make([]customer.PersonalPostalAddress, len(identificationData.Contact.PostalAddresses))
									for i, addr := range identificationData.Contact.PostalAddresses {
										addresses[i] = customer.PersonalPostalAddress{
											Address:            addr.Address,
											AdditionalInfo:     addr.AdditionalInfo,
											DistrictName:       addr.DistrictName,
											TownName:           addr.TownName,
											PostCode:           addr.PostCode,
											Country:            insurer.CountryCode(addr.Country),
											CountrySubDivision: insurer.CountrySubDivision(addr.CountrySubDivision),
										}
									}
									return addresses
								}(),
								Phones: func() *[]customer.Phone {
									if identificationData.Contact.Phones == nil {
										return nil
									}
									phones := make([]customer.Phone, len(*identificationData.Contact.Phones))
									for i, phone := range *identificationData.Contact.Phones {
										phones[i] = customer.Phone{
											CountryCallingCode: phone.CountryCallingCode,
											AreaCode: func() *insurer.PhoneAreaCode {
												if phone.AreaCode == nil {
													return nil
												}
												ac := insurer.PhoneAreaCode(*phone.AreaCode)
												return &ac
											}(),
											Number:         phone.Number,
											PhoneExtension: phone.PhoneExtension,
										}
									}
									return &phones
								}(),
							},
						},
						Qualification: func() *customer.PersonalQualificationData {
							qualificationData, err := req.Body.Data.QuoteCustomer.QualificationData.AsPersonalQualificationData()
							if err != nil {
								return nil
							}
							return &customer.PersonalQualificationData{
								UpdateDateTime:    qualificationData.UpdateDateTime,
								PEPIdentification: customer.PEPIdentification(qualificationData.PepIdentification),
								LifePensionPlans:  string(qualificationData.LifePensionPlans),
								Occupations: func() *[]customer.Occupation {
									if qualificationData.Occupation == nil {
										return nil
									}
									occupations := make([]customer.Occupation, len(*qualificationData.Occupation))
									for i, occ := range *qualificationData.Occupation {
										occupations[i] = customer.Occupation{
											Details:        occ.Details,
											OccupationCode: occ.OccupationCode,
											OccupationCodeType: func() *customer.OccupationCodeType {
												if occ.OccupationCodeType == nil {
													return nil
												}
												t := customer.OccupationCodeType(*occ.OccupationCodeType)
												return &t
											}(),
										}
									}
									return &occupations
								}(),
								InformedRevenue: func() *customer.PersonalInformedRevenue {
									if qualificationData.InformedRevenue == nil {
										return nil
									}
									return &customer.PersonalInformedRevenue{
										Amount: qualificationData.InformedRevenue.Amount,
										Currency: func() *insurer.Currency {
											if qualificationData.InformedRevenue.Currency == nil {
												return nil
											}
											c := insurer.Currency(*qualificationData.InformedRevenue.Currency)
											return &c
										}(),
										Date: qualificationData.InformedRevenue.Date,
										IncomeFrequency: func() *customer.IncomeFrequency {
											if qualificationData.InformedRevenue.IncomeFrequency == nil {
												return nil
											}
											f := customer.IncomeFrequency(*qualificationData.InformedRevenue.IncomeFrequency)
											return &f
										}(),
									}
								}(),
								InformedPatrimony: func() *customer.PersonalInformedPatrimony {
									if qualificationData.InformedPatrimony == nil {
										return nil
									}
									return &customer.PersonalInformedPatrimony{
										Amount: qualificationData.InformedPatrimony.Amount,
										Currency: func() *insurer.Currency {
											if qualificationData.InformedPatrimony.Currency == nil {
												return nil
											}
											c := insurer.Currency(*qualificationData.InformedPatrimony.Currency)
											return &c
										}(),
										Year: qualificationData.InformedPatrimony.Year,
									}
								}(),
							}
						}(),
						ComplimentaryInfo: func() *customer.PersonalComplimentaryInformationData {
							complimentaryInfoData, err := req.Body.Data.QuoteCustomer.ComplimentaryInformationData.AsPersonalComplimentaryInformationData()
							if err != nil {
								return nil
							}
							return &customer.PersonalComplimentaryInformationData{
								UpdateDateTime:        complimentaryInfoData.UpdateDateTime,
								StartDate:             complimentaryInfoData.StartDate,
								RelationshipBeginning: complimentaryInfoData.RelationshipBeginning,
								ProductsServices: func() []customer.ProductsAndServices {
									products := make([]customer.ProductsAndServices, len(complimentaryInfoData.ProductsServices))
									for i, ps := range complimentaryInfoData.ProductsServices {
										products[i] = customer.ProductsAndServices{
											Contract:          ps.Contract,
											InsuranceLineCode: ps.InsuranceLineCode,
											Type:              customer.ProductServiceType(ps.Type),
											Procurators: func() *[]customer.Procurator {
												if ps.Procurators == nil {
													return nil
												}
												procurators := make([]customer.Procurator, len(*ps.Procurators))
												for j, proc := range *ps.Procurators {
													procurators[j] = customer.Procurator{
														CivilName:  proc.CivilName,
														SocialName: proc.SocialName,
														CpfNumber:  proc.CpfNumber,
														Nature:     customer.ProcuratorNature(proc.Nature),
													}
												}
												return &procurators
											}(),
										}
									}
									return products
								}(),
							}
						}(),
					}
				}(),
				Business: func() *quote.BusinessData {
					identificationData, err := req.Body.Data.QuoteCustomer.IdentificationData.AsBusinessIdentificationData()
					if err != nil {
						return nil
					}
					return &quote.BusinessData{
						Identification: &customer.BusinessIdentificationData{
							UpdateDateTime:    identificationData.UpdateDateTime,
							BusinessID:        identificationData.BusinessID,
							BrandName:         identificationData.BrandName,
							BusinessName:      identificationData.BusinessName,
							BusinessTradeName: identificationData.BusinessTradeName,
							IncorporationDate: identificationData.IncorporationDate,
							CompanyInfo: customer.CompanyInfo{
								CNPJ: identificationData.CompanyInfo.CnpjNumber,
								Name: identificationData.CompanyInfo.Name,
							},
							Document: customer.BusinessDocument{
								CNPJNumber:                      identificationData.Document.BusinesscnpjNumber,
								RegistrationNumberOriginCountry: identificationData.Document.BusinessRegisterNumberOriginCountry,
								ExpirationDate:                  identificationData.Document.ExpirationDate,
								Country: func() *insurer.CountryCode {
									if identificationData.Document.Country == nil {
										return nil
									}
									c := insurer.CountryCode(*identificationData.Document.Country)
									return &c
								}(),
							},
							Type: func() *customer.BusinessType {
								if identificationData.Type == nil {
									return nil
								}
								t := customer.BusinessType(*identificationData.Type)
								return &t
							}(),
							Contact: customer.BusinessContact{
								PostalAddresses: func() []customer.BusinessPostalAddress {
									addresses := make([]customer.BusinessPostalAddress, len(identificationData.Contact.PostalAddresses))
									for i, addr := range identificationData.Contact.PostalAddresses {
										addresses[i] = customer.BusinessPostalAddress{
											Address:        addr.Address,
											AdditionalInfo: addr.AdditionalInfo,
											DistrictName:   addr.DistrictName,
											TownName:       addr.TownName,
											PostCode:       addr.PostCode,
											Country:        addr.Country,
											CountryCode: func() *insurer.CountryCode {
												if addr.CountryCode == nil {
													return nil
												}
												c := insurer.CountryCode(*addr.CountryCode)
												return &c
											}(),
											IBGETownCode:       addr.IbgeTownCode,
											CountrySubDivision: insurer.CountrySubDivision(addr.CountrySubDivision),
											GeographicCoordinates: func() *customer.GeographicCoordinates {
												if addr.GeographicCoordinates == nil || addr.GeographicCoordinates.Latitude == nil || addr.GeographicCoordinates.Longitude == nil {
													return nil
												}
												return &customer.GeographicCoordinates{
													Latitude:  *addr.GeographicCoordinates.Latitude,
													Longitude: *addr.GeographicCoordinates.Longitude,
												}
											}(),
										}
									}
									return addresses
								}(),
								Phones: func() *[]customer.Phone {
									if identificationData.Contact.Phones == nil {
										return nil
									}
									phones := make([]customer.Phone, len(*identificationData.Contact.Phones))
									for i, phone := range *identificationData.Contact.Phones {
										phones[i] = customer.Phone{
											CountryCallingCode: phone.CountryCallingCode,
											AreaCode: func() *insurer.PhoneAreaCode {
												if phone.AreaCode != nil {
													ac := insurer.PhoneAreaCode(*phone.AreaCode)
													return &ac
												}
												return nil
											}(),
											Number:         phone.Number,
											PhoneExtension: phone.PhoneExtension,
										}
									}
									return &phones
								}(),
								Emails: func() *[]customer.Email {
									if identificationData.Contact.Emails == nil {
										return nil
									}
									emails := make([]customer.Email, len(*identificationData.Contact.Emails))
									for i, email := range *identificationData.Contact.Emails {
										emails[i] = customer.Email{
											Email: email.Email,
										}
									}
									return &emails
								}(),
							},
							Parties: func() *[]customer.BusinessParty {
								if identificationData.Parties == nil {
									return nil
								}
								parties := make([]customer.BusinessParty, len(*identificationData.Parties))
								for i, party := range *identificationData.Parties {
									parties[i] = customer.BusinessParty{
										CivilName:              party.CivilName,
										SocialName:             party.SocialName,
										StartDate:              party.StartDate,
										Shareholding:           party.Shareholding,
										DocumentType:           party.DocumentType,
										DocumentNumber:         party.DocumentNumber,
										DocumentExpirationDate: party.DocumentExpirationDate,
										DocumentCountry: func() *insurer.CountryCode {
											if party.DocumentCountry != nil {
												c := insurer.CountryCode(*party.DocumentCountry)
												return &c
											}
											return nil
										}(),
										Type: func() *customer.BusinessPartyType {
											if party.Type != nil {
												t := customer.BusinessPartyType(*party.Type)
												return &t
											}
											return nil
										}(),
									}
								}
								return &parties
							}(),
						},
						Qualification: func() *customer.BusinessQualificationData {
							qualificationData, err := req.Body.Data.QuoteCustomer.QualificationData.AsBusinessQualificationData()
							if err != nil {
								return nil
							}
							return &customer.BusinessQualificationData{
								UpdateDateTime:  qualificationData.UpdateDateTime,
								MainBranch:      qualificationData.MainBranch,
								SecondaryBranch: qualificationData.SecondaryBranch,
								InformedRevenue: func() *customer.BusinessInformedRevenue {
									if qualificationData.InformedRevenue == nil {
										return nil
									}
									return &customer.BusinessInformedRevenue{
										Amount: qualificationData.InformedRevenue.Amount,
										Currency: func() *insurer.Currency {
											if qualificationData.InformedRevenue.Currency == nil {
												return nil
											}
											c := insurer.Currency(*qualificationData.InformedRevenue.Currency)
											return &c
										}(),
										IncomeFrequency: func() *customer.IncomeFrequency {
											if qualificationData.InformedRevenue.IncomeFrequency == nil {
												return nil
											}
											f := customer.IncomeFrequency(*qualificationData.InformedRevenue.IncomeFrequency)
											return &f
										}(),
										Year: qualificationData.InformedRevenue.Year,
									}
								}(),
								InformedPatrimony: func() *customer.BusinessInformedPatrimony {
									if qualificationData.InformedPatrimony == nil {
										return nil
									}
									return &customer.BusinessInformedPatrimony{
										Amount: qualificationData.InformedPatrimony.Amount,
										Currency: func() *insurer.Currency {
											if qualificationData.InformedPatrimony.Currency == nil {
												return nil
											}
											c := insurer.Currency(*qualificationData.InformedPatrimony.Currency)
											return &c
										}(),
										Date: qualificationData.InformedPatrimony.Date,
									}
								}(),
							}
						}(),
						ComplimentaryInfo: func() *customer.BusinessComplimentaryInformationData {
							complimentaryInfoData, err := req.Body.Data.QuoteCustomer.ComplimentaryInformationData.AsBusinessComplimentaryInformationData()
							if err != nil {
								return nil
							}
							return &customer.BusinessComplimentaryInformationData{
								UpdateDateTime:        complimentaryInfoData.UpdateDateTime,
								StartDate:             complimentaryInfoData.StartDate,
								RelationshipBeginning: complimentaryInfoData.RelationshipBeginning,
								ProductsServices: func() []customer.ProductsAndServices {
									products := make([]customer.ProductsAndServices, len(complimentaryInfoData.ProductsServices))
									for i, ps := range complimentaryInfoData.ProductsServices {
										products[i] = customer.ProductsAndServices{
											Contract:          ps.Contract,
											InsuranceLineCode: ps.InsuranceLineCode,
											Type:              customer.ProductServiceType(ps.Type),
											Procurators: func() *[]customer.Procurator {
												if ps.Procurators == nil {
													return nil
												}
												procurators := make([]customer.Procurator, len(*ps.Procurators))
												for j, proc := range *ps.Procurators {
													procurators[j] = customer.Procurator{
														CivilName:  proc.CivilName,
														SocialName: proc.SocialName,
														CpfNumber:  proc.CnpjCpfNumber,
														Nature:     customer.ProcuratorNature(proc.Nature),
													}
												}
												return &procurators
											}(),
										}
									}
									return products
								}(),
							}
						}(),
					}
				}(),
			},
			HistoricalData: func() *quoteauto.HistoricalData {
				if req.Body.Data.HistoricalData == nil {
					return nil
				}
				return &quoteauto.HistoricalData{
					Customer: func() *quote.Customer {
						if req.Body.Data.HistoricalData.Customer == nil {
							return nil
						}
						return &quote.Customer{
							Personal: func() *quote.PersonalData {
								identificationData, err := req.Body.Data.HistoricalData.Customer.IdentificationData.AsHistoricalPersonalIdentificationData()
								if err != nil {
									return nil
								}
								return &quote.PersonalData{
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
										Contact: customer.PersonalContact{
											PostalAddresses: func() []customer.PersonalPostalAddress {
												addresses := make([]customer.PersonalPostalAddress, len(identificationData.Contact.PostalAddresses))
												for i, addr := range identificationData.Contact.PostalAddresses {
													addresses[i] = customer.PersonalPostalAddress{
														Address:            addr.Address,
														AdditionalInfo:     addr.AdditionalInfo,
														DistrictName:       addr.DistrictName,
														TownName:           addr.TownName,
														PostCode:           addr.PostCode,
														Country:            insurer.CountryCode(addr.Country),
														CountrySubDivision: insurer.CountrySubDivision(addr.CountrySubDivision),
													}
												}
												return addresses
											}(),
											Phones: func() *[]customer.Phone {
												if identificationData.Contact.Phones == nil {
													return nil
												}
												phones := make([]customer.Phone, len(*identificationData.Contact.Phones))
												for i, phone := range *identificationData.Contact.Phones {
													phones[i] = customer.Phone{
														CountryCallingCode: phone.CountryCallingCode,
														AreaCode: func() *insurer.PhoneAreaCode {
															if phone.AreaCode == nil {
																return nil
															}
															ac := insurer.PhoneAreaCode(*phone.AreaCode)
															return &ac
														}(),
														Number:         phone.Number,
														PhoneExtension: phone.PhoneExtension,
													}
												}
												return &phones
											}(),
										},
									},
									Qualification: func() *customer.PersonalQualificationData {
										qualificationData, err := req.Body.Data.HistoricalData.Customer.QualificationData.AsHistoricalPersonalQualificationData()
										if err != nil {
											return nil
										}
										return &customer.PersonalQualificationData{
											UpdateDateTime:    qualificationData.UpdateDateTime,
											PEPIdentification: customer.PEPIdentification(qualificationData.PepIdentification),
											LifePensionPlans:  string(qualificationData.LifePensionPlans),
											Occupations: func() *[]customer.Occupation {
												if qualificationData.Occupation == nil {
													return nil
												}
												occupations := make([]customer.Occupation, len(*qualificationData.Occupation))
												for i, occ := range *qualificationData.Occupation {
													occupations[i] = customer.Occupation{
														Details:        occ.Details,
														OccupationCode: occ.OccupationCode,
														OccupationCodeType: func() *customer.OccupationCodeType {
															if occ.OccupationCodeType == nil {
																return nil
															}
															t := customer.OccupationCodeType(*occ.OccupationCodeType)
															return &t
														}(),
													}
												}
												return &occupations
											}(),
											InformedRevenue: func() *customer.PersonalInformedRevenue {
												if qualificationData.InformedRevenue == nil {
													return nil
												}
												return &customer.PersonalInformedRevenue{
													Amount: qualificationData.InformedRevenue.Amount,
													Currency: func() *insurer.Currency {
														if qualificationData.InformedRevenue.Currency == nil {
															return nil
														}
														c := insurer.Currency(*qualificationData.InformedRevenue.Currency)
														return &c
													}(),
													Date: qualificationData.InformedRevenue.Date,
													IncomeFrequency: func() *customer.IncomeFrequency {
														if qualificationData.InformedRevenue.IncomeFrequency == nil {
															return nil
														}
														f := customer.IncomeFrequency(*qualificationData.InformedRevenue.IncomeFrequency)
														return &f
													}(),
												}
											}(),
											InformedPatrimony: func() *customer.PersonalInformedPatrimony {
												if qualificationData.InformedPatrimony == nil {
													return nil
												}
												return &customer.PersonalInformedPatrimony{
													Amount: qualificationData.InformedPatrimony.Amount,
													Currency: func() *insurer.Currency {
														if qualificationData.InformedPatrimony.Currency == nil {
															return nil
														}
														c := insurer.Currency(*qualificationData.InformedPatrimony.Currency)
														return &c
													}(),
													Year: qualificationData.InformedPatrimony.Year,
												}
											}(),
										}
									}(),
									ComplimentaryInfo: func() *customer.PersonalComplimentaryInformationData {
										complimentaryInfoData, err := req.Body.Data.HistoricalData.Customer.ComplimentaryInformationData.AsHistoricalPersonalComplimentaryInformationData()
										if err != nil {
											return nil
										}
										return &customer.PersonalComplimentaryInformationData{
											UpdateDateTime:        complimentaryInfoData.UpdateDateTime,
											StartDate:             complimentaryInfoData.StartDate,
											RelationshipBeginning: complimentaryInfoData.RelationshipBeginning,
											ProductsServices: func() []customer.ProductsAndServices {
												products := make([]customer.ProductsAndServices, len(complimentaryInfoData.ProductsServices))
												for i, ps := range complimentaryInfoData.ProductsServices {
													products[i] = customer.ProductsAndServices{
														Contract:          ps.Contract,
														InsuranceLineCode: ps.InsuranceLineCode,
														Type:              customer.ProductServiceType(ps.Type),
														Procurators: func() *[]customer.Procurator {
															if ps.Procurators == nil {
																return nil
															}
															procurators := make([]customer.Procurator, len(*ps.Procurators))
															for j, proc := range *ps.Procurators {
																procurators[j] = customer.Procurator{
																	CivilName:  proc.CivilName,
																	SocialName: proc.SocialName,
																	CpfNumber:  proc.CpfNumber,
																	Nature:     customer.ProcuratorNature(proc.Nature),
																}
															}
															return &procurators
														}(),
													}
												}
												return products
											}(),
										}
									}(),
								}
							}(),
							Business: func() *quote.BusinessData {
								identificationData, err := req.Body.Data.HistoricalData.Customer.IdentificationData.AsHistoricalBusinessIdentificationData()
								if err != nil {
									return nil
								}
								return &quote.BusinessData{
									Identification: &customer.BusinessIdentificationData{
										UpdateDateTime:    identificationData.UpdateDateTime,
										BusinessID:        identificationData.BusinessID,
										BrandName:         identificationData.BrandName,
										BusinessName:      identificationData.BusinessName,
										BusinessTradeName: identificationData.BusinessTradeName,
										IncorporationDate: identificationData.IncorporationDate,
										CompanyInfo: customer.CompanyInfo{
											CNPJ: identificationData.CompanyInfo.CnpjNumber,
											Name: identificationData.CompanyInfo.Name,
										},
										Document: customer.BusinessDocument{
											CNPJNumber:                      identificationData.Document.BusinesscnpjNumber,
											RegistrationNumberOriginCountry: identificationData.Document.BusinessRegisterNumberOriginCountry,
											ExpirationDate:                  identificationData.Document.ExpirationDate,
											Country: func() *insurer.CountryCode {
												if identificationData.Document.Country == nil {
													return nil
												}
												c := insurer.CountryCode(*identificationData.Document.Country)
												return &c
											}(),
										},
										Type: func() *customer.BusinessType {
											if identificationData.Type == nil {
												return nil
											}
											t := customer.BusinessType(*identificationData.Type)
											return &t
										}(),
										Contact: customer.BusinessContact{
											PostalAddresses: func() []customer.BusinessPostalAddress {
												addresses := make([]customer.BusinessPostalAddress, len(identificationData.Contact.PostalAddresses))
												for i, addr := range identificationData.Contact.PostalAddresses {
													addresses[i] = customer.BusinessPostalAddress{
														Address:        addr.Address,
														AdditionalInfo: addr.AdditionalInfo,
														DistrictName:   addr.DistrictName,
														TownName:       addr.TownName,
														PostCode:       addr.PostCode,
														Country:        addr.Country,
														CountryCode: func() *insurer.CountryCode {
															if addr.CountryCode == nil {
																return nil
															}
															c := insurer.CountryCode(*addr.CountryCode)
															return &c
														}(),
														IBGETownCode:       addr.IbgeTownCode,
														CountrySubDivision: insurer.CountrySubDivision(addr.CountrySubDivision),
														GeographicCoordinates: func() *customer.GeographicCoordinates {
															if addr.GeographicCoordinates == nil || addr.GeographicCoordinates.Latitude == nil || addr.GeographicCoordinates.Longitude == nil {
																return nil
															}
															return &customer.GeographicCoordinates{
																Latitude:  *addr.GeographicCoordinates.Latitude,
																Longitude: *addr.GeographicCoordinates.Longitude,
															}
														}(),
													}
												}
												return addresses
											}(),
											Phones: func() *[]customer.Phone {
												if identificationData.Contact.Phones == nil {
													return nil
												}
												phones := make([]customer.Phone, len(*identificationData.Contact.Phones))
												for i, phone := range *identificationData.Contact.Phones {
													phones[i] = customer.Phone{
														CountryCallingCode: phone.CountryCallingCode,
														AreaCode: func() *insurer.PhoneAreaCode {
															if phone.AreaCode != nil {
																ac := insurer.PhoneAreaCode(*phone.AreaCode)
																return &ac
															}
															return nil
														}(),
														Number:         phone.Number,
														PhoneExtension: phone.PhoneExtension,
													}
												}
												return &phones
											}(),
											Emails: func() *[]customer.Email {
												if identificationData.Contact.Emails == nil {
													return nil
												}
												emails := make([]customer.Email, len(*identificationData.Contact.Emails))
												for i, email := range *identificationData.Contact.Emails {
													emails[i] = customer.Email{
														Email: email.Email,
													}
												}
												return &emails
											}(),
										},
										Parties: func() *[]customer.BusinessParty {
											if identificationData.Parties == nil {
												return nil
											}
											parties := make([]customer.BusinessParty, len(*identificationData.Parties))
											for i, party := range *identificationData.Parties {
												parties[i] = customer.BusinessParty{
													CivilName:              party.CivilName,
													SocialName:             party.SocialName,
													StartDate:              party.StartDate,
													Shareholding:           party.Shareholding,
													DocumentType:           party.DocumentType,
													DocumentNumber:         party.DocumentNumber,
													DocumentExpirationDate: party.DocumentExpirationDate,
													DocumentCountry: func() *insurer.CountryCode {
														if party.DocumentCountry != nil {
															c := insurer.CountryCode(*party.DocumentCountry)
															return &c
														}
														return nil
													}(),
													Type: func() *customer.BusinessPartyType {
														if party.Type != nil {
															t := customer.BusinessPartyType(*party.Type)
															return &t
														}
														return nil
													}(),
												}
											}
											return &parties
										}(),
									},
									Qualification: func() *customer.BusinessQualificationData {
										qualificationData, err := req.Body.Data.HistoricalData.Customer.QualificationData.AsHistoricalBusinessQualificationData()
										if err != nil {
											return nil
										}
										return &customer.BusinessQualificationData{
											UpdateDateTime:  qualificationData.UpdateDateTime,
											MainBranch:      qualificationData.MainBranch,
											SecondaryBranch: qualificationData.SecondaryBranch,
											InformedRevenue: func() *customer.BusinessInformedRevenue {
												if qualificationData.InformedRevenue == nil {
													return nil
												}
												return &customer.BusinessInformedRevenue{
													Amount: qualificationData.InformedRevenue.Amount,
													Currency: func() *insurer.Currency {
														if qualificationData.InformedRevenue.Currency == nil {
															return nil
														}
														c := insurer.Currency(*qualificationData.InformedRevenue.Currency)
														return &c
													}(),
													IncomeFrequency: func() *customer.IncomeFrequency {
														if qualificationData.InformedRevenue.IncomeFrequency == nil {
															return nil
														}
														f := customer.IncomeFrequency(*qualificationData.InformedRevenue.IncomeFrequency)
														return &f
													}(),
													Year: qualificationData.InformedRevenue.Year,
												}
											}(),
											InformedPatrimony: func() *customer.BusinessInformedPatrimony {
												if qualificationData.InformedPatrimony == nil {
													return nil
												}
												return &customer.BusinessInformedPatrimony{
													Amount: qualificationData.InformedPatrimony.Amount,
													Currency: func() *insurer.Currency {
														if qualificationData.InformedPatrimony.Currency == nil {
															return nil
														}
														c := insurer.Currency(*qualificationData.InformedPatrimony.Currency)
														return &c
													}(),
													Date: qualificationData.InformedPatrimony.Date,
												}
											}(),
										}
									}(),
									ComplimentaryInfo: func() *customer.BusinessComplimentaryInformationData {
										complimentaryInfoData, err := req.Body.Data.HistoricalData.Customer.ComplimentaryInformationData.AsHistoricalBusinessComplimentaryInformationData()
										if err != nil {
											return nil
										}
										return &customer.BusinessComplimentaryInformationData{
											UpdateDateTime:        complimentaryInfoData.UpdateDateTime,
											StartDate:             complimentaryInfoData.StartDate,
											RelationshipBeginning: complimentaryInfoData.RelationshipBeginning,
											ProductsServices: func() []customer.ProductsAndServices {
												products := make([]customer.ProductsAndServices, len(complimentaryInfoData.ProductsServices))
												for i, ps := range complimentaryInfoData.ProductsServices {
													products[i] = customer.ProductsAndServices{
														Contract:          ps.Contract,
														InsuranceLineCode: ps.InsuranceLineCode,
														Type:              customer.ProductServiceType(ps.Type),
														Procurators: func() *[]customer.Procurator {
															if ps.Procurators == nil {
																return nil
															}
															procurators := make([]customer.Procurator, len(*ps.Procurators))
															for j, proc := range *ps.Procurators {
																procurators[j] = customer.Procurator{
																	CivilName:  proc.CivilName,
																	SocialName: proc.SocialName,
																	CpfNumber:  proc.CnpjCpfNumber,
																	Nature:     customer.ProcuratorNature(proc.Nature),
																}
															}
															return &procurators
														}(),
													}
												}
												return products
											}(),
										}
									}(),
								}
							}(),
						}
					}(),
					Policies: nil, // TODO: Fill policies if needed
				}
			}(),
			Coverages: func() []quoteauto.LeadCoverage {
				if req.Body.Data.QuoteData.Coverages == nil {
					return nil
				}
				coverages := make([]quoteauto.LeadCoverage, len(req.Body.Data.QuoteData.Coverages))
				for i, cov := range req.Body.Data.QuoteData.Coverages {
					coverages[i] = quoteauto.LeadCoverage{
						Branch: cov.Branch,
						Code:   auto.CoverageCode(cov.Code),
					}
				}
				return coverages
			}(),
		},
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
	lead, err := s.service.CancelLead(ctx, request.ConsentID, orgID, quote.PatchData{
		AuthorIdentificationType:   insurer.IdentificationType(request.Body.Data.Author.IdentificationType),
		AuthorIdentificationNumber: request.Body.Data.Author.IdentificationNumber,
	})
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
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	quote := quoteauto.Quote{
		ConsentID: request.Body.Data.ConsentID,
		OrgID:     orgID,
		Data: quoteauto.Data{
			ExpirationDateTime: request.Body.Data.ExpirationDateTime,
			Customer: quote.Customer{
				Personal: func() *quote.PersonalData {
					identificationData, err := request.Body.Data.QuoteCustomer.IdentificationData.AsPersonalIdentificationData()
					if err != nil {
						return nil
					}
					return &quote.PersonalData{
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
							Contact: customer.PersonalContact{
								PostalAddresses: func() []customer.PersonalPostalAddress {
									addresses := make([]customer.PersonalPostalAddress, len(identificationData.Contact.PostalAddresses))
									for i, addr := range identificationData.Contact.PostalAddresses {
										addresses[i] = customer.PersonalPostalAddress{
											Address:            addr.Address,
											AdditionalInfo:     addr.AdditionalInfo,
											DistrictName:       addr.DistrictName,
											TownName:           addr.TownName,
											PostCode:           addr.PostCode,
											Country:            insurer.CountryCode(addr.Country),
											CountrySubDivision: insurer.CountrySubDivision(addr.CountrySubDivision),
										}
									}
									return addresses
								}(),
								Phones: func() *[]customer.Phone {
									if identificationData.Contact.Phones == nil {
										return nil
									}
									phones := make([]customer.Phone, len(*identificationData.Contact.Phones))
									for i, phone := range *identificationData.Contact.Phones {
										phones[i] = customer.Phone{
											CountryCallingCode: phone.CountryCallingCode,
											AreaCode: func() *insurer.PhoneAreaCode {
												if phone.AreaCode == nil {
													return nil
												}
												ac := insurer.PhoneAreaCode(*phone.AreaCode)
												return &ac
											}(),
											Number:         phone.Number,
											PhoneExtension: phone.PhoneExtension,
										}
									}
									return &phones
								}(),
							},
						},
						Qualification: func() *customer.PersonalQualificationData {
							qualificationData, err := request.Body.Data.QuoteCustomer.QualificationData.AsPersonalQualificationData()
							if err != nil {
								return nil
							}
							return &customer.PersonalQualificationData{
								UpdateDateTime:    qualificationData.UpdateDateTime,
								PEPIdentification: customer.PEPIdentification(qualificationData.PepIdentification),
								LifePensionPlans:  string(qualificationData.LifePensionPlans),
								Occupations: func() *[]customer.Occupation {
									if qualificationData.Occupation == nil {
										return nil
									}
									occupations := make([]customer.Occupation, len(*qualificationData.Occupation))
									for i, occ := range *qualificationData.Occupation {
										occupations[i] = customer.Occupation{
											Details:        occ.Details,
											OccupationCode: occ.OccupationCode,
											OccupationCodeType: func() *customer.OccupationCodeType {
												if occ.OccupationCodeType == nil {
													return nil
												}
												t := customer.OccupationCodeType(*occ.OccupationCodeType)
												return &t
											}(),
										}
									}
									return &occupations
								}(),
								InformedRevenue: func() *customer.PersonalInformedRevenue {
									if qualificationData.InformedRevenue == nil {
										return nil
									}
									return &customer.PersonalInformedRevenue{
										Amount: qualificationData.InformedRevenue.Amount,
										Currency: func() *insurer.Currency {
											if qualificationData.InformedRevenue.Currency == nil {
												return nil
											}
											c := insurer.Currency(*qualificationData.InformedRevenue.Currency)
											return &c
										}(),
										Date: qualificationData.InformedRevenue.Date,
										IncomeFrequency: func() *customer.IncomeFrequency {
											if qualificationData.InformedRevenue.IncomeFrequency == nil {
												return nil
											}
											f := customer.IncomeFrequency(*qualificationData.InformedRevenue.IncomeFrequency)
											return &f
										}(),
									}
								}(),
								InformedPatrimony: func() *customer.PersonalInformedPatrimony {
									if qualificationData.InformedPatrimony == nil {
										return nil
									}
									return &customer.PersonalInformedPatrimony{
										Amount: qualificationData.InformedPatrimony.Amount,
										Currency: func() *insurer.Currency {
											if qualificationData.InformedPatrimony.Currency == nil {
												return nil
											}
											c := insurer.Currency(*qualificationData.InformedPatrimony.Currency)
											return &c
										}(),
										Year: qualificationData.InformedPatrimony.Year,
									}
								}(),
							}
						}(),
						ComplimentaryInfo: func() *customer.PersonalComplimentaryInformationData {
							complimentaryInfoData, err := request.Body.Data.QuoteCustomer.ComplimentaryInformationData.AsPersonalComplimentaryInformationData()
							if err != nil {
								return nil
							}
							return &customer.PersonalComplimentaryInformationData{
								UpdateDateTime:        complimentaryInfoData.UpdateDateTime,
								StartDate:             complimentaryInfoData.StartDate,
								RelationshipBeginning: complimentaryInfoData.RelationshipBeginning,
								ProductsServices: func() []customer.ProductsAndServices {
									products := make([]customer.ProductsAndServices, len(complimentaryInfoData.ProductsServices))
									for i, ps := range complimentaryInfoData.ProductsServices {
										products[i] = customer.ProductsAndServices{
											Contract:          ps.Contract,
											InsuranceLineCode: ps.InsuranceLineCode,
											Type:              customer.ProductServiceType(ps.Type),
											Procurators: func() *[]customer.Procurator {
												if ps.Procurators == nil {
													return nil
												}
												procurators := make([]customer.Procurator, len(*ps.Procurators))
												for j, proc := range *ps.Procurators {
													procurators[j] = customer.Procurator{
														CivilName:  proc.CivilName,
														SocialName: proc.SocialName,
														CpfNumber:  proc.CpfNumber,
														Nature:     customer.ProcuratorNature(proc.Nature),
													}
												}
												return &procurators
											}(),
										}
									}
									return products
								}(),
							}
						}(),
					}
				}(),
				Business: func() *quote.BusinessData {
					identificationData, err := request.Body.Data.QuoteCustomer.IdentificationData.AsBusinessIdentificationData()
					if err != nil {
						return nil
					}
					return &quote.BusinessData{
						Identification: &customer.BusinessIdentificationData{
							UpdateDateTime:    identificationData.UpdateDateTime,
							BusinessID:        identificationData.BusinessID,
							BrandName:         identificationData.BrandName,
							BusinessName:      identificationData.BusinessName,
							BusinessTradeName: identificationData.BusinessTradeName,
							IncorporationDate: identificationData.IncorporationDate,
							CompanyInfo: customer.CompanyInfo{
								CNPJ: identificationData.CompanyInfo.CnpjNumber,
								Name: identificationData.CompanyInfo.Name,
							},
							Document: customer.BusinessDocument{
								CNPJNumber:                      identificationData.Document.BusinesscnpjNumber,
								RegistrationNumberOriginCountry: identificationData.Document.BusinessRegisterNumberOriginCountry,
								ExpirationDate:                  identificationData.Document.ExpirationDate,
								Country: func() *insurer.CountryCode {
									if identificationData.Document.Country == nil {
										return nil
									}
									c := insurer.CountryCode(*identificationData.Document.Country)
									return &c
								}(),
							},
							Type: func() *customer.BusinessType {
								if identificationData.Type == nil {
									return nil
								}
								t := customer.BusinessType(*identificationData.Type)
								return &t
							}(),
							Contact: customer.BusinessContact{
								PostalAddresses: func() []customer.BusinessPostalAddress {
									addresses := make([]customer.BusinessPostalAddress, len(identificationData.Contact.PostalAddresses))
									for i, addr := range identificationData.Contact.PostalAddresses {
										addresses[i] = customer.BusinessPostalAddress{
											Address:        addr.Address,
											AdditionalInfo: addr.AdditionalInfo,
											DistrictName:   addr.DistrictName,
											TownName:       addr.TownName,
											PostCode:       addr.PostCode,
											Country:        addr.Country,
											CountryCode: func() *insurer.CountryCode {
												if addr.CountryCode == nil {
													return nil
												}
												c := insurer.CountryCode(*addr.CountryCode)
												return &c
											}(),
											IBGETownCode:       addr.IbgeTownCode,
											CountrySubDivision: insurer.CountrySubDivision(addr.CountrySubDivision),
											GeographicCoordinates: func() *customer.GeographicCoordinates {
												if addr.GeographicCoordinates == nil || addr.GeographicCoordinates.Latitude == nil || addr.GeographicCoordinates.Longitude == nil {
													return nil
												}
												return &customer.GeographicCoordinates{
													Latitude:  *addr.GeographicCoordinates.Latitude,
													Longitude: *addr.GeographicCoordinates.Longitude,
												}
											}(),
										}
									}
									return addresses
								}(),
								Phones: func() *[]customer.Phone {
									if identificationData.Contact.Phones == nil {
										return nil
									}
									phones := make([]customer.Phone, len(*identificationData.Contact.Phones))
									for i, phone := range *identificationData.Contact.Phones {
										phones[i] = customer.Phone{
											CountryCallingCode: phone.CountryCallingCode,
											AreaCode: func() *insurer.PhoneAreaCode {
												if phone.AreaCode != nil {
													ac := insurer.PhoneAreaCode(*phone.AreaCode)
													return &ac
												}
												return nil
											}(),
											Number:         phone.Number,
											PhoneExtension: phone.PhoneExtension,
										}
									}
									return &phones
								}(),
								Emails: func() *[]customer.Email {
									if identificationData.Contact.Emails == nil {
										return nil
									}
									emails := make([]customer.Email, len(*identificationData.Contact.Emails))
									for i, email := range *identificationData.Contact.Emails {
										emails[i] = customer.Email{
											Email: email.Email,
										}
									}
									return &emails
								}(),
							},
							Parties: func() *[]customer.BusinessParty {
								if identificationData.Parties == nil {
									return nil
								}
								parties := make([]customer.BusinessParty, len(*identificationData.Parties))
								for i, party := range *identificationData.Parties {
									parties[i] = customer.BusinessParty{
										CivilName:              party.CivilName,
										SocialName:             party.SocialName,
										StartDate:              party.StartDate,
										Shareholding:           party.Shareholding,
										DocumentType:           party.DocumentType,
										DocumentNumber:         party.DocumentNumber,
										DocumentExpirationDate: party.DocumentExpirationDate,
										DocumentCountry: func() *insurer.CountryCode {
											if party.DocumentCountry != nil {
												c := insurer.CountryCode(*party.DocumentCountry)
												return &c
											}
											return nil
										}(),
										Type: func() *customer.BusinessPartyType {
											if party.Type != nil {
												t := customer.BusinessPartyType(*party.Type)
												return &t
											}
											return nil
										}(),
									}
								}
								return &parties
							}(),
						},
						Qualification: func() *customer.BusinessQualificationData {
							qualificationData, err := request.Body.Data.QuoteCustomer.QualificationData.AsBusinessQualificationData()
							if err != nil {
								return nil
							}
							return &customer.BusinessQualificationData{
								UpdateDateTime:  qualificationData.UpdateDateTime,
								MainBranch:      qualificationData.MainBranch,
								SecondaryBranch: qualificationData.SecondaryBranch,
								InformedRevenue: func() *customer.BusinessInformedRevenue {
									if qualificationData.InformedRevenue == nil {
										return nil
									}
									return &customer.BusinessInformedRevenue{
										Amount: qualificationData.InformedRevenue.Amount,
										Currency: func() *insurer.Currency {
											if qualificationData.InformedRevenue.Currency == nil {
												return nil
											}
											c := insurer.Currency(*qualificationData.InformedRevenue.Currency)
											return &c
										}(),
										IncomeFrequency: func() *customer.IncomeFrequency {
											if qualificationData.InformedRevenue.IncomeFrequency == nil {
												return nil
											}
											f := customer.IncomeFrequency(*qualificationData.InformedRevenue.IncomeFrequency)
											return &f
										}(),
										Year: qualificationData.InformedRevenue.Year,
									}
								}(),
								InformedPatrimony: func() *customer.BusinessInformedPatrimony {
									if qualificationData.InformedPatrimony == nil {
										return nil
									}
									return &customer.BusinessInformedPatrimony{
										Amount: qualificationData.InformedPatrimony.Amount,
										Currency: func() *insurer.Currency {
											if qualificationData.InformedPatrimony.Currency == nil {
												return nil
											}
											c := insurer.Currency(*qualificationData.InformedPatrimony.Currency)
											return &c
										}(),
										Date: qualificationData.InformedPatrimony.Date,
									}
								}(),
							}
						}(),
						ComplimentaryInfo: func() *customer.BusinessComplimentaryInformationData {
							complimentaryInfoData, err := request.Body.Data.QuoteCustomer.ComplimentaryInformationData.AsBusinessComplimentaryInformationData()
							if err != nil {
								return nil
							}
							return &customer.BusinessComplimentaryInformationData{
								UpdateDateTime:        complimentaryInfoData.UpdateDateTime,
								StartDate:             complimentaryInfoData.StartDate,
								RelationshipBeginning: complimentaryInfoData.RelationshipBeginning,
								ProductsServices: func() []customer.ProductsAndServices {
									products := make([]customer.ProductsAndServices, len(complimentaryInfoData.ProductsServices))
									for i, ps := range complimentaryInfoData.ProductsServices {
										products[i] = customer.ProductsAndServices{
											Contract:          ps.Contract,
											InsuranceLineCode: ps.InsuranceLineCode,
											Type:              customer.ProductServiceType(ps.Type),
											Procurators: func() *[]customer.Procurator {
												if ps.Procurators == nil {
													return nil
												}
												procurators := make([]customer.Procurator, len(*ps.Procurators))
												for j, proc := range *ps.Procurators {
													procurators[j] = customer.Procurator{
														CivilName:  proc.CivilName,
														SocialName: proc.SocialName,
														CpfNumber:  proc.CnpjCpfNumber,
														Nature:     customer.ProcuratorNature(proc.Nature),
													}
												}
												return &procurators
											}(),
										}
									}
									return products
								}(),
							}
						}(),
					}
				}(),
			},
			IsCollectiveStipulated: request.Body.Data.QuoteData.IsCollectiveStipulated,
			HasAnIndividualItem:    request.Body.Data.QuoteData.HasAnIndividualItem,
			TermStartDate:          request.Body.Data.QuoteData.TermStartDate,
			TermEndDate:            request.Body.Data.QuoteData.TermEndDate,
			TermType:               insurer.ValidityType(request.Body.Data.QuoteData.TermType),
			InsuranceType:          quoteauto.InsuranceType(request.Body.Data.QuoteData.InsuranceType),
			PolicyID:               request.Body.Data.QuoteData.PolicyID,
			InsurerID:              request.Body.Data.QuoteData.InsurerID,
			IdentifierCode:         request.Body.Data.QuoteData.IdentifierCode,
			BonusClass:             request.Body.Data.QuoteData.BonusClass,
			Currency:               insurer.Currency(request.Body.Data.QuoteData.Currency),
			InsuredObject: func() *quoteauto.InsuredObject {
				if request.Body.Data.QuoteData.InsuredObject == nil {
					return nil
				}
				obj := request.Body.Data.QuoteData.InsuredObject
				return &quoteauto.InsuredObject{
					Identification: obj.Identification,
					Model: func() *quoteauto.InsuredObjectModel {
						if obj.Model == nil {
							return nil
						}
						return &quoteauto.InsuredObjectModel{
							Brand:           obj.Model.Brand,
							ModelName:       obj.Model.ModelName,
							ModelYear:       obj.Model.ModelYear,
							ManufactureYear: obj.Model.ManufactureYear,
						}
					}(),
					Modality: func() *auto.InsuredObjectModality {
						if obj.Modality == nil {
							return nil
						}
						m := auto.InsuredObjectModality(*obj.Modality)
						return &m
					}(),
					TableUsed: func() *auto.AmountReferenceTable {
						if obj.TableUsed == nil {
							return nil
						}
						t := auto.AmountReferenceTable(*obj.TableUsed)
						return &t
					}(),
					ModelCode:        obj.ModelCode,
					AdjustmentFactor: obj.AdjustmentFactor,
					ValuedDetermined: obj.ValueDetermined,
					Tax: func() *quoteauto.Tax {
						if obj.Tax == nil {
							return nil
						}
						return &quoteauto.Tax{
							Exempt: obj.Tax.Exempt,
							Type: func() *quoteauto.TaxType {
								if obj.Tax.Type == nil {
									return nil
								}
								t := quoteauto.TaxType(*obj.Tax.Type)
								return &t
							}(),
							ExemptionPercentage: obj.Tax.ExemptionPercentage,
						}
					}(),
					DoorsNumber:  obj.DoorsNumber,
					Color:        obj.Color,
					LicensePlate: obj.LicensePlate,
					VehicleUsage: func() []auto.VehicleUsage {
						usage := make([]auto.VehicleUsage, len(obj.VehicleUse))
						for i, v := range obj.VehicleUse {
							usage[i] = auto.VehicleUsage(v)
						}
						return usage
					}(),
					CommercialActivityType: func() *[]quoteauto.CommercialActivityType {
						if obj.CommercialActivityType == nil {
							return nil
						}
						types := make([]quoteauto.CommercialActivityType, len(*obj.CommercialActivityType))
						for i, t := range *obj.CommercialActivityType {
							types[i] = quoteauto.CommercialActivityType(t)
						}
						return &types
					}(),
					RiskManagementSystem: func() *[]quoteauto.RiskManagementSystem {
						if obj.RiskManagementSystem == nil {
							return nil
						}
						systems := make([]quoteauto.RiskManagementSystem, len(*obj.RiskManagementSystem))
						for i, s := range *obj.RiskManagementSystem {
							systems[i] = quoteauto.RiskManagementSystem(s)
						}
						return &systems
					}(),
					IsTransportCargoInsurance: obj.IsTransportedCargoInsurance,
					LoadsCarriedInsured: func() *[]quoteauto.LoadType {
						if obj.LoadsCarriedinsured == nil {
							return nil
						}
						loads := make([]quoteauto.LoadType, len(*obj.LoadsCarriedinsured))
						for i, l := range *obj.LoadsCarriedinsured {
							loads[i] = quoteauto.LoadType(l)
						}
						return &loads
					}(),
					IsEquipmentAttached: obj.IsEquipmentsAttached,
					EquipmentsAttached: func() *quoteauto.EquipmentAttached {
						if obj.EquipmentsAttached == nil {
							return nil
						}
						return &quoteauto.EquipmentAttached{
							Amount:           obj.EquipmentsAttached.EquipmentsAttachedAmount,
							IsDesireCoverage: obj.EquipmentsAttached.IsDesireCoverage,
						}
					}(),
					Chasis:                         obj.Chassis,
					IsAuctionChassisRescheduled:    obj.IsAuctionChassisRescheduled,
					IsBrandNew:                     obj.IsBrandNew,
					DepartureDateFromCarDealership: obj.DepartureDateFromCarDealership,
					VehicleInvoice: func() *quoteauto.VehicleInvoice {
						if obj.VehicleInvoice == nil {
							return nil
						}
						return &quoteauto.VehicleInvoice{
							Amount: obj.VehicleInvoice.VehicleAmount,
							Number: obj.VehicleInvoice.VehicleNumber,
						}
					}(),
					Fuel:     quoteauto.Fuel(obj.Fuel),
					IsGasKit: obj.IsGasKit,
					GasKit: func() *quoteauto.GasKit {
						if obj.GasKit == nil {
							return nil
						}
						return &quoteauto.GasKit{
							IsDesireCoverage: &obj.GasKit.IsDesireCoverage,
							Amount:           obj.GasKit.GasKitAmount,
						}
					}(),
					IsArmouredVehicle:       obj.IsArmouredVehicle,
					IsActiveTrackingVehicle: &obj.IsActiveTrackingDevice,
					FrequentTrafficArea: func() *quoteauto.TrafficArea {
						if obj.FrequentTrafficArea == nil {
							return nil
						}
						area := quoteauto.TrafficArea(*obj.FrequentTrafficArea)
						return &area
					}(),
					OvernightPostCode: obj.OvernightPostCode,
					RiskLocation: func() *quoteauto.RiskLocation {
						if obj.RiskLocationInfo == nil {
							return nil
						}
						rl := obj.RiskLocationInfo
						return &quoteauto.RiskLocation{
							IsUsedCollege: rl.IsUsedCollege,
							UsedCollege: func() *quoteauto.RiskLocationUsage {
								if rl.UsedCollege == nil {
									return nil
								}
								return &quoteauto.RiskLocationUsage{
									IsKeptInGarage:        rl.UsedCollege.IsKeptInGarage,
									DistanceFromResidence: rl.UsedCollege.DistanceFromResidence,
								}
							}(),
							IsUsedCommuteWork: rl.IsUsedCommuteWork,
							UsedCommuteWork: func() *quoteauto.RiskLocationUsage {
								if rl.UsedCommuteWork == nil {
									return nil
								}
								return &quoteauto.RiskLocationUsage{
									IsKeptInGarage:        rl.UsedCommuteWork.IsKeptInGarage,
									DistanceFromResidence: rl.UsedCommuteWork.DistanceFromResidence,
								}
							}(),
							KmAveragePerWeek: rl.KmAveragePerWeek,
							Housing: func() *quoteauto.RiskLocationHousing {
								if rl.Housing == nil {
									return nil
								}
								return &quoteauto.RiskLocationHousing{
									Type: func() *quoteauto.HousingType {
										if rl.Housing.Type == nil {
											return nil
										}
										t := quoteauto.HousingType(*rl.Housing.Type)
										return &t
									}(),
									IsKeptInGarage: rl.Housing.IsKeptInGarage,
									GateType: func() *quoteauto.GateType {
										if rl.Housing.GateType == nil {
											return nil
										}
										gt := quoteauto.GateType(*rl.Housing.GateType)
										return &gt
									}(),
								}
							}(),
						}
					}(),
					IsExtendCoverageAgedBetween18And25: obj.IsExtendCoverageAgedBetween18And25,
					DriverBetween18And25YearsOldGender: func() *quoteauto.DriverGender {
						if obj.DriverBetween18And25YearsOldGender == nil {
							return nil
						}
						g := quoteauto.DriverGender(*obj.DriverBetween18And25YearsOldGender)
						return &g
					}(),
					WasThereAClaim: &obj.WasThereAClaim,
					ClaimNotifications: func() *[]quoteauto.ClaimNotification {
						if obj.ClaimNotifications == nil {
							return nil
						}
						notifications := make([]quoteauto.ClaimNotification, len(*obj.ClaimNotifications))
						for i, n := range *obj.ClaimNotifications {
							claimAmount := n.ClaimAmount
							claimDesc := n.ClaimDescription
							notifications[i] = quoteauto.ClaimNotification{
								ClaimAmount:      &claimAmount,
								ClaimDescription: &claimDesc,
							}
						}
						return &notifications
					}(),
				}
			}(),
			Coverages: func() *[]quoteauto.Coverage {
				if len(request.Body.Data.QuoteData.Coverages) == 0 {
					return nil
				}
				coverages := make([]quoteauto.Coverage, len(request.Body.Data.QuoteData.Coverages))
				for i, cov := range request.Body.Data.QuoteData.Coverages {
					coverages[i] = quoteauto.Coverage{
						Branch:                       cov.Branch,
						Code:                         auto.CoverageCode(cov.Code),
						Description:                  cov.Description,
						IsSeparateContractingAllowed: cov.IsSeparateContractingAllowed,
						MaxLMI:                       cov.MaxLMI,
						InternalCode:                 cov.InternalCode,
					}
				}
				return &coverages
			}(),
			CustomData: func() *quote.CustomData {
				if request.Body.Data.QuoteCustomData == nil {
					return nil
				}
				customData := request.Body.Data.QuoteCustomData
				return &quote.CustomData{
					CustomerIdentification: func() *[]quote.CustomDataField {
						if customData.CustomerIdentification == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.CustomerIdentification))
						for i, f := range *customData.CustomerIdentification {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
					CustomerQualification: func() *[]quote.CustomDataField {
						if customData.CustomerQualification == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.CustomerQualification))
						for i, f := range *customData.CustomerQualification {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
					CustomerComplimentaryInfo: func() *[]quote.CustomDataField {
						if customData.CustomerComplimentaryInfo == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.CustomerComplimentaryInfo))
						for i, f := range *customData.CustomerComplimentaryInfo {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
					GeneralQuoteInfo: func() *[]quote.CustomDataField {
						if customData.GeneralQuoteInfo == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.GeneralQuoteInfo))
						for i, f := range *customData.GeneralQuoteInfo {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
					RiskLocationInfo: func() *[]quote.CustomDataField {
						if customData.RiskLocationInfo == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.RiskLocationInfo))
						for i, f := range *customData.RiskLocationInfo {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
					InsuredObjects: func() *[]quote.CustomDataField {
						if customData.InsuredObjects == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.InsuredObjects))
						for i, f := range *customData.InsuredObjects {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
					Beneficiaries: func() *[]quote.CustomDataField {
						if customData.Beneficiaries == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.Beneficiaries))
						for i, f := range *customData.Beneficiaries {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
					Coverages: func() *[]quote.CustomDataField {
						if customData.Coverages == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.Coverages))
						for i, f := range *customData.Coverages {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
					GeneralClaimInfo: func() *[]quote.CustomDataField {
						if customData.GeneralClaimInfo == nil {
							return nil
						}
						fields := make([]quote.CustomDataField, len(*customData.GeneralClaimInfo))
						for i, f := range *customData.GeneralClaimInfo {
							fields[i] = quote.CustomDataField{
								FieldID: f.FieldID,
								Value:   f.Value,
							}
						}
						return &fields
					}(),
				}
			}(),
			HistoricalData: func() *quoteauto.HistoricalData {
				if request.Body.Data.HistoricalData == nil {
					return nil
				}
				return &quoteauto.HistoricalData{
					Customer: func() *quote.Customer {
						if request.Body.Data.HistoricalData.Customer == nil {
							return nil
						}
						return &quote.Customer{
							Personal: func() *quote.PersonalData {
								identificationData, err := request.Body.Data.HistoricalData.Customer.IdentificationData.AsHistoricalPersonalIdentificationData()
								if err != nil {
									return nil
								}
								return &quote.PersonalData{
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
										Contact: customer.PersonalContact{
											PostalAddresses: func() []customer.PersonalPostalAddress {
												addresses := make([]customer.PersonalPostalAddress, len(identificationData.Contact.PostalAddresses))
												for i, addr := range identificationData.Contact.PostalAddresses {
													addresses[i] = customer.PersonalPostalAddress{
														Address:            addr.Address,
														AdditionalInfo:     addr.AdditionalInfo,
														DistrictName:       addr.DistrictName,
														TownName:           addr.TownName,
														PostCode:           addr.PostCode,
														Country:            insurer.CountryCode(addr.Country),
														CountrySubDivision: insurer.CountrySubDivision(addr.CountrySubDivision),
													}
												}
												return addresses
											}(),
											Phones: func() *[]customer.Phone {
												if identificationData.Contact.Phones == nil {
													return nil
												}
												phones := make([]customer.Phone, len(*identificationData.Contact.Phones))
												for i, phone := range *identificationData.Contact.Phones {
													phones[i] = customer.Phone{
														CountryCallingCode: phone.CountryCallingCode,
														AreaCode: func() *insurer.PhoneAreaCode {
															if phone.AreaCode == nil {
																return nil
															}
															ac := insurer.PhoneAreaCode(*phone.AreaCode)
															return &ac
														}(),
														Number:         phone.Number,
														PhoneExtension: phone.PhoneExtension,
													}
												}
												return &phones
											}(),
										},
									},
									Qualification: func() *customer.PersonalQualificationData {
										qualificationData, err := request.Body.Data.HistoricalData.Customer.QualificationData.AsHistoricalPersonalQualificationData()
										if err != nil {
											return nil
										}
										return &customer.PersonalQualificationData{
											UpdateDateTime:    qualificationData.UpdateDateTime,
											PEPIdentification: customer.PEPIdentification(qualificationData.PepIdentification),
											LifePensionPlans:  string(qualificationData.LifePensionPlans),
											Occupations: func() *[]customer.Occupation {
												if qualificationData.Occupation == nil {
													return nil
												}
												occupations := make([]customer.Occupation, len(*qualificationData.Occupation))
												for i, occ := range *qualificationData.Occupation {
													occupations[i] = customer.Occupation{
														Details:        occ.Details,
														OccupationCode: occ.OccupationCode,
														OccupationCodeType: func() *customer.OccupationCodeType {
															if occ.OccupationCodeType == nil {
																return nil
															}
															t := customer.OccupationCodeType(*occ.OccupationCodeType)
															return &t
														}(),
													}
												}
												return &occupations
											}(),
											InformedRevenue: func() *customer.PersonalInformedRevenue {
												if qualificationData.InformedRevenue == nil {
													return nil
												}
												return &customer.PersonalInformedRevenue{
													Amount: qualificationData.InformedRevenue.Amount,
													Currency: func() *insurer.Currency {
														if qualificationData.InformedRevenue.Currency == nil {
															return nil
														}
														c := insurer.Currency(*qualificationData.InformedRevenue.Currency)
														return &c
													}(),
													Date: qualificationData.InformedRevenue.Date,
													IncomeFrequency: func() *customer.IncomeFrequency {
														if qualificationData.InformedRevenue.IncomeFrequency == nil {
															return nil
														}
														f := customer.IncomeFrequency(*qualificationData.InformedRevenue.IncomeFrequency)
														return &f
													}(),
												}
											}(),
											InformedPatrimony: func() *customer.PersonalInformedPatrimony {
												if qualificationData.InformedPatrimony == nil {
													return nil
												}
												return &customer.PersonalInformedPatrimony{
													Amount: qualificationData.InformedPatrimony.Amount,
													Currency: func() *insurer.Currency {
														if qualificationData.InformedPatrimony.Currency == nil {
															return nil
														}
														c := insurer.Currency(*qualificationData.InformedPatrimony.Currency)
														return &c
													}(),
													Year: qualificationData.InformedPatrimony.Year,
												}
											}(),
										}
									}(),
									ComplimentaryInfo: func() *customer.PersonalComplimentaryInformationData {
										complimentaryInfoData, err := request.Body.Data.HistoricalData.Customer.ComplimentaryInformationData.AsHistoricalPersonalComplimentaryInformationData()
										if err != nil {
											return nil
										}
										return &customer.PersonalComplimentaryInformationData{
											UpdateDateTime:        complimentaryInfoData.UpdateDateTime,
											StartDate:             complimentaryInfoData.StartDate,
											RelationshipBeginning: complimentaryInfoData.RelationshipBeginning,
											ProductsServices: func() []customer.ProductsAndServices {
												products := make([]customer.ProductsAndServices, len(complimentaryInfoData.ProductsServices))
												for i, ps := range complimentaryInfoData.ProductsServices {
													products[i] = customer.ProductsAndServices{
														Contract:          ps.Contract,
														InsuranceLineCode: ps.InsuranceLineCode,
														Type:              customer.ProductServiceType(ps.Type),
														Procurators: func() *[]customer.Procurator {
															if ps.Procurators == nil {
																return nil
															}
															procurators := make([]customer.Procurator, len(*ps.Procurators))
															for j, proc := range *ps.Procurators {
																procurators[j] = customer.Procurator{
																	CivilName:  proc.CivilName,
																	SocialName: proc.SocialName,
																	CpfNumber:  proc.CpfNumber,
																	Nature:     customer.ProcuratorNature(proc.Nature),
																}
															}
															return &procurators
														}(),
													}
												}
												return products
											}(),
										}
									}(),
								}
							}(),
							Business: func() *quote.BusinessData {
								identificationData, err := request.Body.Data.HistoricalData.Customer.IdentificationData.AsHistoricalBusinessIdentificationData()
								if err != nil {
									return nil
								}
								return &quote.BusinessData{
									Identification: &customer.BusinessIdentificationData{
										UpdateDateTime:    identificationData.UpdateDateTime,
										BusinessID:        identificationData.BusinessID,
										BrandName:         identificationData.BrandName,
										BusinessName:      identificationData.BusinessName,
										BusinessTradeName: identificationData.BusinessTradeName,
										IncorporationDate: identificationData.IncorporationDate,
										CompanyInfo: customer.CompanyInfo{
											CNPJ: identificationData.CompanyInfo.CnpjNumber,
											Name: identificationData.CompanyInfo.Name,
										},
										Document: customer.BusinessDocument{
											CNPJNumber:                      identificationData.Document.BusinesscnpjNumber,
											RegistrationNumberOriginCountry: identificationData.Document.BusinessRegisterNumberOriginCountry,
											ExpirationDate:                  identificationData.Document.ExpirationDate,
											Country: func() *insurer.CountryCode {
												if identificationData.Document.Country == nil {
													return nil
												}
												c := insurer.CountryCode(*identificationData.Document.Country)
												return &c
											}(),
										},
										Type: func() *customer.BusinessType {
											if identificationData.Type == nil {
												return nil
											}
											t := customer.BusinessType(*identificationData.Type)
											return &t
										}(),
										Contact: customer.BusinessContact{
											PostalAddresses: func() []customer.BusinessPostalAddress {
												addresses := make([]customer.BusinessPostalAddress, len(identificationData.Contact.PostalAddresses))
												for i, addr := range identificationData.Contact.PostalAddresses {
													addresses[i] = customer.BusinessPostalAddress{
														Address:        addr.Address,
														AdditionalInfo: addr.AdditionalInfo,
														DistrictName:   addr.DistrictName,
														TownName:       addr.TownName,
														PostCode:       addr.PostCode,
														Country:        addr.Country,
														CountryCode: func() *insurer.CountryCode {
															if addr.CountryCode == nil {
																return nil
															}
															c := insurer.CountryCode(*addr.CountryCode)
															return &c
														}(),
														IBGETownCode:       addr.IbgeTownCode,
														CountrySubDivision: insurer.CountrySubDivision(addr.CountrySubDivision),
														GeographicCoordinates: func() *customer.GeographicCoordinates {
															if addr.GeographicCoordinates == nil || addr.GeographicCoordinates.Latitude == nil || addr.GeographicCoordinates.Longitude == nil {
																return nil
															}
															return &customer.GeographicCoordinates{
																Latitude:  *addr.GeographicCoordinates.Latitude,
																Longitude: *addr.GeographicCoordinates.Longitude,
															}
														}(),
													}
												}
												return addresses
											}(),
											Phones: func() *[]customer.Phone {
												if identificationData.Contact.Phones == nil {
													return nil
												}
												phones := make([]customer.Phone, len(*identificationData.Contact.Phones))
												for i, phone := range *identificationData.Contact.Phones {
													phones[i] = customer.Phone{
														CountryCallingCode: phone.CountryCallingCode,
														AreaCode: func() *insurer.PhoneAreaCode {
															if phone.AreaCode != nil {
																ac := insurer.PhoneAreaCode(*phone.AreaCode)
																return &ac
															}
															return nil
														}(),
														Number:         phone.Number,
														PhoneExtension: phone.PhoneExtension,
													}
												}
												return &phones
											}(),
											Emails: func() *[]customer.Email {
												if identificationData.Contact.Emails == nil {
													return nil
												}
												emails := make([]customer.Email, len(*identificationData.Contact.Emails))
												for i, email := range *identificationData.Contact.Emails {
													emails[i] = customer.Email{
														Email: email.Email,
													}
												}
												return &emails
											}(),
										},
										Parties: func() *[]customer.BusinessParty {
											if identificationData.Parties == nil {
												return nil
											}
											parties := make([]customer.BusinessParty, len(*identificationData.Parties))
											for i, party := range *identificationData.Parties {
												parties[i] = customer.BusinessParty{
													CivilName:              party.CivilName,
													SocialName:             party.SocialName,
													StartDate:              party.StartDate,
													Shareholding:           party.Shareholding,
													DocumentType:           party.DocumentType,
													DocumentNumber:         party.DocumentNumber,
													DocumentExpirationDate: party.DocumentExpirationDate,
													DocumentCountry: func() *insurer.CountryCode {
														if party.DocumentCountry != nil {
															c := insurer.CountryCode(*party.DocumentCountry)
															return &c
														}
														return nil
													}(),
													Type: func() *customer.BusinessPartyType {
														if party.Type != nil {
															t := customer.BusinessPartyType(*party.Type)
															return &t
														}
														return nil
													}(),
												}
											}
											return &parties
										}(),
									},
									Qualification: func() *customer.BusinessQualificationData {
										qualificationData, err := request.Body.Data.HistoricalData.Customer.QualificationData.AsHistoricalBusinessQualificationData()
										if err != nil {
											return nil
										}
										return &customer.BusinessQualificationData{
											UpdateDateTime:  qualificationData.UpdateDateTime,
											MainBranch:      qualificationData.MainBranch,
											SecondaryBranch: qualificationData.SecondaryBranch,
											InformedRevenue: func() *customer.BusinessInformedRevenue {
												if qualificationData.InformedRevenue == nil {
													return nil
												}
												return &customer.BusinessInformedRevenue{
													Amount: qualificationData.InformedRevenue.Amount,
													Currency: func() *insurer.Currency {
														if qualificationData.InformedRevenue.Currency == nil {
															return nil
														}
														c := insurer.Currency(*qualificationData.InformedRevenue.Currency)
														return &c
													}(),
													IncomeFrequency: func() *customer.IncomeFrequency {
														if qualificationData.InformedRevenue.IncomeFrequency == nil {
															return nil
														}
														f := customer.IncomeFrequency(*qualificationData.InformedRevenue.IncomeFrequency)
														return &f
													}(),
													Year: qualificationData.InformedRevenue.Year,
												}
											}(),
											InformedPatrimony: func() *customer.BusinessInformedPatrimony {
												if qualificationData.InformedPatrimony == nil {
													return nil
												}
												return &customer.BusinessInformedPatrimony{
													Amount: qualificationData.InformedPatrimony.Amount,
													Currency: func() *insurer.Currency {
														if qualificationData.InformedPatrimony.Currency == nil {
															return nil
														}
														c := insurer.Currency(*qualificationData.InformedPatrimony.Currency)
														return &c
													}(),
													Date: qualificationData.InformedPatrimony.Date,
												}
											}(),
										}
									}(),
									ComplimentaryInfo: func() *customer.BusinessComplimentaryInformationData {
										complimentaryInfoData, err := request.Body.Data.HistoricalData.Customer.ComplimentaryInformationData.AsHistoricalBusinessComplimentaryInformationData()
										if err != nil {
											return nil
										}
										return &customer.BusinessComplimentaryInformationData{
											UpdateDateTime:        complimentaryInfoData.UpdateDateTime,
											StartDate:             complimentaryInfoData.StartDate,
											RelationshipBeginning: complimentaryInfoData.RelationshipBeginning,
											ProductsServices: func() []customer.ProductsAndServices {
												products := make([]customer.ProductsAndServices, len(complimentaryInfoData.ProductsServices))
												for i, ps := range complimentaryInfoData.ProductsServices {
													products[i] = customer.ProductsAndServices{
														Contract:          ps.Contract,
														InsuranceLineCode: ps.InsuranceLineCode,
														Type:              customer.ProductServiceType(ps.Type),
														Procurators: func() *[]customer.Procurator {
															if ps.Procurators == nil {
																return nil
															}
															procurators := make([]customer.Procurator, len(*ps.Procurators))
															for j, proc := range *ps.Procurators {
																procurators[j] = customer.Procurator{
																	CivilName:  proc.CivilName,
																	SocialName: proc.SocialName,
																	CpfNumber:  proc.CnpjCpfNumber,
																	Nature:     customer.ProcuratorNature(proc.Nature),
																}
															}
															return &procurators
														}(),
													}
												}
												return products
											}(),
										}
									}(),
								}
							}(),
						}
					}(),
					Policies: nil, // TODO: Fill policies if needed
				}
			}(),
		},
	}

	err := s.service.CreateQuote(ctx, &quote)
	if err != nil {
		return nil, err
	}

	resp := ResponseQuoteAuto{
		Data: QuoteStatus{
			Status:               QuoteStatusStatus(quote.Status),
			StatusUpdateDateTime: quote.StatusUpdatedAt,
		},
		Meta:  *api.NewMeta(),
		Links: *api.NewLinks(s.baseURL + "/request/" + quote.ConsentID + "/quote-status"),
	}
	return PostQuoteAuto201JSONResponse{CreatedResponseQuoteRequestAutoJSONResponse(resp)}, nil
}

func (s Server) PatchQuoteAuto(ctx context.Context, request PatchQuoteAutoRequestObject) (PatchQuoteAutoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	quote, err := s.service.Update(ctx, request.ConsentID, orgID, quote.PatchData{
		Status:                     quote.Status(request.Body.Data.Status),
		InsurerQuoteID:             request.Body.Data.InsurerQuoteID,
		AuthorIdentificationType:   insurer.IdentificationType(request.Body.Data.Author.IdentificationType),
		AuthorIdentificationNumber: request.Body.Data.Author.IdentificationNumber,
	})
	if err != nil {
		return nil, err
	}

	resp := ResponsePatch{
		Data: struct {
			InsurerQuoteID *string `json:"insurerQuoteId,omitempty"`
			Links          *struct {
				Redirect string `json:"redirect"`
			} `json:"links,omitempty"`
			ProtocolDateTime *timeutil.DateTime      `json:"protocolDateTime,omitempty"`
			ProtocolNumber   *string                 `json:"protocolNumber,omitempty"`
			Status           ResponsePatchDataStatus `json:"status"`
		}{
			InsurerQuoteID: quote.Data.InsurerQuoteID,
			Links: func() *struct {
				Redirect string `json:"redirect"`
			} {
				if quote.Data.RedirectLink == nil {
					return nil
				}
				return &struct {
					Redirect string `json:"redirect"`
				}{
					Redirect: *quote.Data.RedirectLink,
				}
			}(),
			ProtocolDateTime: quote.Data.ProtocolDateTime,
			ProtocolNumber:   quote.Data.ProtocolNumber,
			Status:           ResponsePatchDataStatus(quote.Status),
		},
	}
	return PatchQuoteAuto200JSONResponse{N200UpdatedQuoteAutoJSONResponse(resp)}, nil
}

func (s Server) GetQuoteAuto(ctx context.Context, request GetQuoteAutoRequestObject) (GetQuoteAutoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	q, err := s.service.Quote(ctx, request.ConsentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseQuoteStatusAuto{
		Data: struct {
			QuoteInfo            *QuoteStatusAuto                  `json:"quoteInfo,omitempty"`
			RejectionReason      *string                           `json:"rejectionReason,omitempty"`
			Status               ResponseQuoteStatusAutoDataStatus `json:"status"`
			StatusUpdateDateTime timeutil.DateTime                 `json:"statusUpdateDateTime"`
		}{
			Status:               ResponseQuoteStatusAutoDataStatus(q.Status),
			StatusUpdateDateTime: q.StatusUpdatedAt,
			RejectionReason:      q.Data.RejectionReason,
			QuoteInfo: func() *QuoteStatusAuto {
				if q.Status != quote.StatusAccepted {
					return nil
				}
				return &QuoteStatusAuto{
					QuoteCustomData: func() *QuoteCustomData {
						if q.Data.CustomData == nil {
							return nil
						}
						customData := q.Data.CustomData
						return &QuoteCustomData{
							CustomerIdentification: func() *[]CustomInfoData {
								if customData.CustomerIdentification == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.CustomerIdentification))
								for i, f := range *customData.CustomerIdentification {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
							CustomerQualification: func() *[]CustomInfoData {
								if customData.CustomerQualification == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.CustomerQualification))
								for i, f := range *customData.CustomerQualification {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
							CustomerComplimentaryInfo: func() *[]CustomInfoData {
								if customData.CustomerComplimentaryInfo == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.CustomerComplimentaryInfo))
								for i, f := range *customData.CustomerComplimentaryInfo {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
							GeneralQuoteInfo: func() *[]CustomInfoData {
								if customData.GeneralQuoteInfo == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.GeneralQuoteInfo))
								for i, f := range *customData.GeneralQuoteInfo {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
							RiskLocationInfo: func() *[]CustomInfoData {
								if customData.RiskLocationInfo == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.RiskLocationInfo))
								for i, f := range *customData.RiskLocationInfo {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
							InsuredObjects: func() *[]CustomInfoData {
								if customData.InsuredObjects == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.InsuredObjects))
								for i, f := range *customData.InsuredObjects {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
							Beneficiaries: func() *[]CustomInfoData {
								if customData.Beneficiaries == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.Beneficiaries))
								for i, f := range *customData.Beneficiaries {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
							Coverages: func() *[]CustomInfoData {
								if customData.Coverages == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.Coverages))
								for i, f := range *customData.Coverages {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
							GeneralClaimInfo: func() *[]CustomInfoData {
								if customData.GeneralClaimInfo == nil {
									return nil
								}
								fields := make([]CustomInfoData, len(*customData.GeneralClaimInfo))
								for i, f := range *customData.GeneralClaimInfo {
									fields[i] = CustomInfoData{
										FieldID: f.FieldID,
										Value:   f.Value,
									}
								}
								return &fields
							}(),
						}
					}(),
					QuoteCustomer: func() QuoteStatusAuto_QuoteCustomer {
						var quoteCustomer QuoteStatusAuto_QuoteCustomer
						if q.Data.Customer.Personal != nil {
							if err := quoteCustomer.FromPersonalCustomerInfo(PersonalCustomerInfo{
								Identification: func() *PersonalIdentificationData {
									if q.Data.Customer.Personal.Identification == nil {
										return nil
									}
									ident := q.Data.Customer.Personal.Identification
									return &PersonalIdentificationData{
										UpdateDateTime:          ident.UpdateDateTime,
										PersonalID:              ident.PersonalID,
										BrandName:               ident.BrandName,
										CivilName:               ident.CivilName,
										SocialName:              ident.SocialName,
										CpfNumber:               ident.CPF,
										HasBrazilianNationality: ident.HasBrazilianNationality,
										CompanyInfo: struct {
											CnpjNumber string `json:"cnpjNumber"`
											Name       string `json:"name"`
										}{
											CnpjNumber: ident.CompanyInfo.CNPJ,
											Name:       ident.CompanyInfo.Name,
										},
										Contact: PersonalContact{
											PostalAddresses: func() []PersonalPostalAddress {
												addresses := make([]PersonalPostalAddress, len(ident.Contact.PostalAddresses))
												for i, addr := range ident.Contact.PostalAddresses {
													addresses[i] = PersonalPostalAddress{
														Address:            addr.Address,
														AdditionalInfo:     addr.AdditionalInfo,
														DistrictName:       addr.DistrictName,
														TownName:           addr.TownName,
														PostCode:           addr.PostCode,
														Country:            PersonalPostalAddressCountry(addr.Country),
														CountrySubDivision: EnumCountrySubDivision(addr.CountrySubDivision),
													}
												}
												return addresses
											}(),
											Phones: func() *[]CustomerPhone {
												if ident.Contact.Phones == nil {
													return nil
												}
												phones := make([]CustomerPhone, len(*ident.Contact.Phones))
												for i, phone := range *ident.Contact.Phones {
													phones[i] = CustomerPhone{
														CountryCallingCode: phone.CountryCallingCode,
														AreaCode: func() *EnumAreaCode {
															if phone.AreaCode != nil {
																ac := EnumAreaCode(*phone.AreaCode)
																return &ac
															}
															return nil
														}(),
														Number:         phone.Number,
														PhoneExtension: phone.PhoneExtension,
													}
												}
												return &phones
											}(),
										},
									}
								}(),
								Qualification: func() *PersonalQualificationData {
									if q.Data.Customer.Personal.Qualification == nil {
										return nil
									}
									qual := q.Data.Customer.Personal.Qualification
									return &PersonalQualificationData{
										UpdateDateTime:    qual.UpdateDateTime,
										PepIdentification: PersonalQualificationDataPepIdentification(qual.PEPIdentification),
										LifePensionPlans:  PersonalQualificationDataLifePensionPlans(qual.LifePensionPlans),
										Occupation: func() *[]struct {
											Details            *string                                                `json:"details,omitempty"`
											OccupationCode     *string                                                `json:"occupationCode,omitempty"`
											OccupationCodeType *PersonalQualificationDataOccupationOccupationCodeType `json:"occupationCodeType,omitempty"`
										} {
											if qual.Occupations == nil {
												return nil
											}
											occupations := make([]struct {
												Details            *string                                                `json:"details,omitempty"`
												OccupationCode     *string                                                `json:"occupationCode,omitempty"`
												OccupationCodeType *PersonalQualificationDataOccupationOccupationCodeType `json:"occupationCodeType,omitempty"`
											}, len(*qual.Occupations))
											for i, occ := range *qual.Occupations {
												occupations[i] = struct {
													Details            *string                                                `json:"details,omitempty"`
													OccupationCode     *string                                                `json:"occupationCode,omitempty"`
													OccupationCodeType *PersonalQualificationDataOccupationOccupationCodeType `json:"occupationCodeType,omitempty"`
												}{
													Details:        occ.Details,
													OccupationCode: occ.OccupationCode,
													OccupationCodeType: func() *PersonalQualificationDataOccupationOccupationCodeType {
														if occ.OccupationCodeType == nil {
															return nil
														}
														t := PersonalQualificationDataOccupationOccupationCodeType(*occ.OccupationCodeType)
														return &t
													}(),
												}
											}
											return &occupations
										}(),
										InformedRevenue: func() *struct {
											Amount          *string                                           `json:"amount"`
											Currency        *PersonalQualificationDataInformedRevenueCurrency `json:"currency,omitempty"`
											Date            *timeutil.BrazilDate                              `json:"date,omitempty"`
											IncomeFrequency *EnumIncomeFrequency                              `json:"incomeFrequency,omitempty"`
										} {
											if qual.InformedRevenue == nil {
												return nil
											}
											return &struct {
												Amount          *string                                           `json:"amount"`
												Currency        *PersonalQualificationDataInformedRevenueCurrency `json:"currency,omitempty"`
												Date            *timeutil.BrazilDate                              `json:"date,omitempty"`
												IncomeFrequency *EnumIncomeFrequency                              `json:"incomeFrequency,omitempty"`
											}{
												Amount: qual.InformedRevenue.Amount,
												Currency: func() *PersonalQualificationDataInformedRevenueCurrency {
													if qual.InformedRevenue.Currency == nil {
														return nil
													}
													c := PersonalQualificationDataInformedRevenueCurrency(*qual.InformedRevenue.Currency)
													return &c
												}(),
												Date: qual.InformedRevenue.Date,
												IncomeFrequency: func() *EnumIncomeFrequency {
													if qual.InformedRevenue.IncomeFrequency == nil {
														return nil
													}
													f := EnumIncomeFrequency(*qual.InformedRevenue.IncomeFrequency)
													return &f
												}(),
											}
										}(),
										InformedPatrimony: func() *struct {
											Amount   *string                                             `json:"amount"`
											Currency *PersonalQualificationDataInformedPatrimonyCurrency `json:"currency,omitempty"`
											Year     *string                                             `json:"year,omitempty"`
										} {
											if qual.InformedPatrimony == nil {
												return nil
											}
											return &struct {
												Amount   *string                                             `json:"amount"`
												Currency *PersonalQualificationDataInformedPatrimonyCurrency `json:"currency,omitempty"`
												Year     *string                                             `json:"year,omitempty"`
											}{
												Amount: qual.InformedPatrimony.Amount,
												Currency: func() *PersonalQualificationDataInformedPatrimonyCurrency {
													if qual.InformedPatrimony.Currency == nil {
														return nil
													}
													c := PersonalQualificationDataInformedPatrimonyCurrency(*qual.InformedPatrimony.Currency)
													return &c
												}(),
												Year: qual.InformedPatrimony.Year,
											}
										}(),
									}
								}(),
								ComplimentaryInfo: func() *PersonalComplimentaryInformationData {
									if q.Data.Customer.Personal.ComplimentaryInfo == nil {
										return nil
									}
									comp := q.Data.Customer.Personal.ComplimentaryInfo
									return &PersonalComplimentaryInformationData{
										UpdateDateTime:        comp.UpdateDateTime,
										StartDate:             comp.StartDate,
										RelationshipBeginning: comp.RelationshipBeginning,
										ProductsServices: func() []struct {
											Contract          string                 `json:"contract"`
											InsuranceLineCode *string                `json:"insuranceLineCode,omitempty"`
											Procurators       *[]PersonalProcurator  `json:"procurators,omitempty"`
											Type              EnumProductServiceType `json:"type"`
										} {
											products := make([]struct {
												Contract          string                 `json:"contract"`
												InsuranceLineCode *string                `json:"insuranceLineCode,omitempty"`
												Procurators       *[]PersonalProcurator  `json:"procurators,omitempty"`
												Type              EnumProductServiceType `json:"type"`
											}, len(comp.ProductsServices))
											for i, ps := range comp.ProductsServices {
												products[i] = struct {
													Contract          string                 `json:"contract"`
													InsuranceLineCode *string                `json:"insuranceLineCode,omitempty"`
													Procurators       *[]PersonalProcurator  `json:"procurators,omitempty"`
													Type              EnumProductServiceType `json:"type"`
												}{
													Contract:          ps.Contract,
													InsuranceLineCode: ps.InsuranceLineCode,
													Type:              EnumProductServiceType(ps.Type),
													Procurators: func() *[]PersonalProcurator {
														if ps.Procurators == nil {
															return nil
														}
														procurators := make([]PersonalProcurator, len(*ps.Procurators))
														for j, proc := range *ps.Procurators {
															procurators[j] = PersonalProcurator{
																CivilName:  proc.CivilName,
																SocialName: proc.SocialName,
																CpfNumber:  proc.CpfNumber,
																Nature:     EnumProcuratorsNaturePersonal(proc.Nature),
															}
														}
														return &procurators
													}(),
												}
											}
											return products
										}(),
									}
								}(),
							}); err != nil {
								slog.ErrorContext(ctx, "failed to convert personal customer info", "error", err.Error())
							}
						} else if q.Data.Customer.Business != nil {
							businessInfo := BusinessCustomerInfo{
								Identification: func() *BusinessIdentificationData {
									if q.Data.Customer.Business.Identification == nil {
										return nil
									}
									ident := q.Data.Customer.Business.Identification
									return &BusinessIdentificationData{
										UpdateDateTime:    ident.UpdateDateTime,
										BusinessID:        ident.BusinessID,
										BrandName:         ident.BrandName,
										BusinessName:      ident.BusinessName,
										BusinessTradeName: ident.BusinessTradeName,
										IncorporationDate: ident.IncorporationDate,
										CompanyInfo: struct {
											CnpjNumber string `json:"cnpjNumber"`
											Name       string `json:"name"`
										}{
											CnpjNumber: ident.CompanyInfo.CNPJ,
											Name:       ident.CompanyInfo.Name,
										},
										Document: BusinessDocument{
											BusinesscnpjNumber:                  ident.Document.CNPJNumber,
											BusinessRegisterNumberOriginCountry: ident.Document.RegistrationNumberOriginCountry,
											ExpirationDate:                      ident.Document.ExpirationDate,
											Country: func() *BusinessDocumentCountry {
												if ident.Document.Country == nil {
													return nil
												}
												c := BusinessDocumentCountry(*ident.Document.Country)
												return &c
											}(),
										},
										Type: func() *BusinessIdentificationDataType {
											if ident.Type == nil {
												return nil
											}
											t := BusinessIdentificationDataType(*ident.Type)
											return &t
										}(),
										Contact: BusinessContact{
											PostalAddresses: func() []BusinessPostalAddress {
												addresses := make([]BusinessPostalAddress, len(ident.Contact.PostalAddresses))
												for i, addr := range ident.Contact.PostalAddresses {
													addresses[i] = BusinessPostalAddress{
														Address:        addr.Address,
														AdditionalInfo: addr.AdditionalInfo,
														DistrictName:   addr.DistrictName,
														TownName:       addr.TownName,
														PostCode:       addr.PostCode,
														Country:        addr.Country,
														CountryCode: func() *BusinessPostalAddressCountryCode {
															if addr.CountryCode == nil {
																return nil
															}
															c := BusinessPostalAddressCountryCode(*addr.CountryCode)
															return &c
														}(),
														IbgeTownCode:       addr.IBGETownCode,
														CountrySubDivision: EnumCountrySubDivision(addr.CountrySubDivision),
														GeographicCoordinates: func() *GeographicCoordinates {
															if addr.GeographicCoordinates == nil {
																return nil
															}
															lat := addr.GeographicCoordinates.Latitude
															lon := addr.GeographicCoordinates.Longitude
															return &GeographicCoordinates{
																Latitude:  &lat,
																Longitude: &lon,
															}
														}(),
													}
												}
												return addresses
											}(),
											Phones: func() *[]CustomerPhone {
												if ident.Contact.Phones == nil {
													return nil
												}
												phones := make([]CustomerPhone, len(*ident.Contact.Phones))
												for i, phone := range *ident.Contact.Phones {
													phones[i] = CustomerPhone{
														CountryCallingCode: phone.CountryCallingCode,
														AreaCode: func() *EnumAreaCode {
															if phone.AreaCode != nil {
																ac := EnumAreaCode(*phone.AreaCode)
																return &ac
															}
															return nil
														}(),
														Number:         phone.Number,
														PhoneExtension: phone.PhoneExtension,
													}
												}
												return &phones
											}(),
											Emails: func() *[]CustomerEmail {
												if ident.Contact.Emails == nil {
													return nil
												}
												emails := make([]CustomerEmail, len(*ident.Contact.Emails))
												for i, email := range *ident.Contact.Emails {
													emails[i] = CustomerEmail{
														Email: email.Email,
													}
												}
												return &emails
											}(),
										},
										Parties: func() *BusinessParties {
											if ident.Parties == nil {
												return nil
											}
											parties := make(BusinessParties, len(*ident.Parties))
											for i, party := range *ident.Parties {
												parties[i] = struct {
													CivilName              *string                         `json:"civilName,omitempty"`
													DocumentCountry        *BusinessPartiesDocumentCountry `json:"documentCountry,omitempty"`
													DocumentExpirationDate *timeutil.BrazilDate            `json:"documentExpirationDate,omitempty"`
													DocumentNumber         *string                         `json:"documentNumber,omitempty"`
													DocumentType           *string                         `json:"documentType,omitempty"`
													Shareholding           *string                         `json:"shareholding,omitempty"`
													SocialName             *string                         `json:"socialName,omitempty"`
													StartDate              *timeutil.BrazilDate            `json:"startDate,omitempty"`
													Type                   *BusinessPartiesType            `json:"type,omitempty"`
												}{
													CivilName:              party.CivilName,
													SocialName:             party.SocialName,
													StartDate:              party.StartDate,
													Shareholding:           party.Shareholding,
													DocumentType:           party.DocumentType,
													DocumentNumber:         party.DocumentNumber,
													DocumentExpirationDate: party.DocumentExpirationDate,
													DocumentCountry: func() *BusinessPartiesDocumentCountry {
														if party.DocumentCountry != nil {
															c := BusinessPartiesDocumentCountry(*party.DocumentCountry)
															return &c
														}
														return nil
													}(),
													Type: func() *BusinessPartiesType {
														if party.Type != nil {
															t := BusinessPartiesType(*party.Type)
															return &t
														}
														return nil
													}(),
												}
											}
											return &parties
										}(),
									}
								}(),
								Qualification: func() *BusinessQualificationData {
									if q.Data.Customer.Business.Qualification == nil {
										return nil
									}
									qual := q.Data.Customer.Business.Qualification
									return &BusinessQualificationData{
										UpdateDateTime:  qual.UpdateDateTime,
										MainBranch:      qual.MainBranch,
										SecondaryBranch: qual.SecondaryBranch,
										InformedRevenue: func() *struct {
											Amount          *string                                           `json:"amount"`
											Currency        *BusinessQualificationDataInformedRevenueCurrency `json:"currency,omitempty"`
											IncomeFrequency *EnumIncomeFrequency                              `json:"incomeFrequency,omitempty"`
											Year            *string                                           `json:"year,omitempty"`
										} {
											if qual.InformedRevenue == nil {
												return nil
											}
											return &struct {
												Amount          *string                                           `json:"amount"`
												Currency        *BusinessQualificationDataInformedRevenueCurrency `json:"currency,omitempty"`
												IncomeFrequency *EnumIncomeFrequency                              `json:"incomeFrequency,omitempty"`
												Year            *string                                           `json:"year,omitempty"`
											}{
												Amount: qual.InformedRevenue.Amount,
												Currency: func() *BusinessQualificationDataInformedRevenueCurrency {
													if qual.InformedRevenue.Currency == nil {
														return nil
													}
													c := BusinessQualificationDataInformedRevenueCurrency(*qual.InformedRevenue.Currency)
													return &c
												}(),
												IncomeFrequency: func() *EnumIncomeFrequency {
													if qual.InformedRevenue.IncomeFrequency == nil {
														return nil
													}
													f := EnumIncomeFrequency(*qual.InformedRevenue.IncomeFrequency)
													return &f
												}(),
												Year: qual.InformedRevenue.Year,
											}
										}(),
										InformedPatrimony: func() *struct {
											Amount   *string                                             `json:"amount"`
											Currency *BusinessQualificationDataInformedPatrimonyCurrency `json:"currency,omitempty"`
											Date     *timeutil.BrazilDate                                `json:"date,omitempty"`
										} {
											if qual.InformedPatrimony == nil {
												return nil
											}
											return &struct {
												Amount   *string                                             `json:"amount"`
												Currency *BusinessQualificationDataInformedPatrimonyCurrency `json:"currency,omitempty"`
												Date     *timeutil.BrazilDate                                `json:"date,omitempty"`
											}{
												Amount: qual.InformedPatrimony.Amount,
												Currency: func() *BusinessQualificationDataInformedPatrimonyCurrency {
													if qual.InformedPatrimony.Currency == nil {
														return nil
													}
													c := BusinessQualificationDataInformedPatrimonyCurrency(*qual.InformedPatrimony.Currency)
													return &c
												}(),
												Date: qual.InformedPatrimony.Date,
											}
										}(),
									}
								}(),
								ComplimentaryInfo: func() *BusinessComplimentaryInformationData {
									if q.Data.Customer.Business.ComplimentaryInfo == nil {
										return nil
									}
									comp := q.Data.Customer.Business.ComplimentaryInfo
									return &BusinessComplimentaryInformationData{
										UpdateDateTime:        comp.UpdateDateTime,
										StartDate:             comp.StartDate,
										RelationshipBeginning: comp.RelationshipBeginning,
										ProductsServices: func() []struct {
											Contract          string                 `json:"contract"`
											InsuranceLineCode *string                `json:"insuranceLineCode,omitempty"`
											Procurators       *[]BusinessProcurator  `json:"procurators,omitempty"`
											Type              EnumProductServiceType `json:"type"`
										} {
											products := make([]struct {
												Contract          string                 `json:"contract"`
												InsuranceLineCode *string                `json:"insuranceLineCode,omitempty"`
												Procurators       *[]BusinessProcurator  `json:"procurators,omitempty"`
												Type              EnumProductServiceType `json:"type"`
											}, len(comp.ProductsServices))
											for i, ps := range comp.ProductsServices {
												products[i] = struct {
													Contract          string                 `json:"contract"`
													InsuranceLineCode *string                `json:"insuranceLineCode,omitempty"`
													Procurators       *[]BusinessProcurator  `json:"procurators,omitempty"`
													Type              EnumProductServiceType `json:"type"`
												}{
													Contract:          ps.Contract,
													InsuranceLineCode: ps.InsuranceLineCode,
													Type:              EnumProductServiceType(ps.Type),
													Procurators: func() *[]BusinessProcurator {
														if ps.Procurators == nil {
															return nil
														}
														procurators := make([]BusinessProcurator, len(*ps.Procurators))
														for j, proc := range *ps.Procurators {
															procurators[j] = BusinessProcurator{
																CivilName:     proc.CivilName,
																SocialName:    proc.SocialName,
																CnpjCpfNumber: proc.CpfNumber,
																Nature:        EnumProcuratorsNatureBusiness(proc.Nature),
															}
														}
														return &procurators
													}(),
												}
											}
											return products
										}(),
									}
								}(),
							}
							if err := quoteCustomer.FromBusinessCustomerInfo(businessInfo); err != nil {
								slog.ErrorContext(ctx, "failed to convert business customer info", "error", err.Error())
							}
						}
						return quoteCustomer
					}(),
					QuoteData: ResultQuoteQuoteAuto{
						BonusClass:                 q.Data.BonusClass,
						Currency:                   ResultQuoteQuoteAutoCurrency(q.Data.Currency),
						HasAnIndividualItem:        q.Data.HasAnIndividualItem,
						IdentifierCode:             q.Data.IdentifierCode,
						IncludesAssistanceServices: q.Data.IncludesAssistanceServices,
						InsuranceType:              ResultQuoteQuoteAutoInsuranceType(q.Data.InsuranceType),
						InsuredObjects: func() *[]QuoteAutoResultInsuredObject {
							if q.Data.InsuredObject == nil {
								return nil
							}
							obj := q.Data.InsuredObject
							insuredObject := QuoteAutoResultInsuredObject{
								AdjustmentFactor: obj.AdjustmentFactor,
								Chassis:          obj.Chasis,
								ClaimNotifications: func() *[]ClaimNotification {
									if obj.ClaimNotifications == nil {
										return nil
									}
									notifications := make([]ClaimNotification, len(*obj.ClaimNotifications))
									for i, cn := range *obj.ClaimNotifications {
										notifications[i] = ClaimNotification{
											ClaimAmount: func() AmountDetails {
												if cn.ClaimAmount == nil {
													return AmountDetails{}
												}
												return *cn.ClaimAmount
											}(),
											ClaimDescription: func() string {
												if cn.ClaimDescription == nil {
													return ""
												}
												return *cn.ClaimDescription
											}(),
										}
									}
									return &notifications
								}(),
								Color:          obj.Color,
								DoorsNumber:    obj.DoorsNumber,
								Identification: obj.Identification,
								Fuel:           QuoteAutoResultInsuredObjectFuel(obj.Fuel),
								IsActiveTrackingDevice: func() bool {
									if obj.IsActiveTrackingVehicle == nil {
										return false
									}
									return *obj.IsActiveTrackingVehicle
								}(),
								IsArmouredVehicle: obj.IsArmouredVehicle,
								ArmouredVehicle: func() *struct {
									ArmouredVehicleAmount *AmountDetails `json:"armouredVehicleAmount,omitempty"`
									IsDesireCoverage      *bool          `json:"isDesireCoverage,omitempty"`
								} {
									if obj.ArmouredVehicle == nil {
										return nil
									}
									return &struct {
										ArmouredVehicleAmount *AmountDetails `json:"armouredVehicleAmount,omitempty"`
										IsDesireCoverage      *bool          `json:"isDesireCoverage,omitempty"`
									}{
										ArmouredVehicleAmount: obj.ArmouredVehicle.Amount,
										IsDesireCoverage:      obj.ArmouredVehicle.IsDesireCoverage,
									}
								}(),
								IsAuctionChassisRescheduled:        obj.IsAuctionChassisRescheduled,
								IsBrandNew:                         obj.IsBrandNew,
								IsEquipmentsAttached:               obj.IsEquipmentAttached,
								IsExtendCoverageAgedBetween18And25: obj.IsExtendCoverageAgedBetween18And25,
								IsGasKit:                           obj.IsGasKit,
								IsTransportedCargoInsurance:        obj.IsTransportCargoInsurance,
								LicensePlate:                       obj.LicensePlate,
								LicensePlateType: func() []QuoteAutoResultInsuredObjectLicensePlateType {
									types := make([]QuoteAutoResultInsuredObjectLicensePlateType, len(obj.LicensePlateType))
									for i, t := range obj.LicensePlateType {
										types[i] = QuoteAutoResultInsuredObjectLicensePlateType(t)
									}
									return types
								}(),
								Tariff: func() *QuoteAutoResultInsuredObjectTariff {
									if obj.Tariff == nil {
										return nil
									}
									t := QuoteAutoResultInsuredObjectTariff(*obj.Tariff)
									return &t
								}(),
								OvernightPostCode: obj.OvernightPostCode,
								VehicleUse: func() []QuoteAutoResultInsuredObjectVehicleUse {
									uses := make([]QuoteAutoResultInsuredObjectVehicleUse, len(obj.VehicleUsage))
									for i, v := range obj.VehicleUsage {
										uses[i] = QuoteAutoResultInsuredObjectVehicleUse(v)
									}
									return uses
								}(),
								WasThereAClaim: func() bool {
									if obj.WasThereAClaim == nil {
										return false
									}
									return *obj.WasThereAClaim
								}(),
								Coverages: func() []QuoteAutoResultInsuredObjectCoverage {
									if q.Data.Coverages == nil {
										return []QuoteAutoResultInsuredObjectCoverage{}
									}
									coverages := make([]QuoteAutoResultInsuredObjectCoverage, len(*q.Data.Coverages))
									for i, c := range *q.Data.Coverages {
										coverages[i] = QuoteAutoResultInsuredObjectCoverage{
											LMI:                      c.MaxLMI,
											Branch:                   c.Branch,
											Code:                     QuoteAutoResultInsuredObjectCoverageCode(c.Code),
											Description:              c.Description,
											InternalCode:             c.InternalCode,
											DaysForTotalCompensation: "", // TODO: Fill if available
										}
									}
									return coverages
								}(),
								CommercialActivityType: func() *[]QuoteAutoResultInsuredObjectCommercialActivityType {
									if obj.CommercialActivityType == nil {
										return nil
									}
									types := make([]QuoteAutoResultInsuredObjectCommercialActivityType, len(*obj.CommercialActivityType))
									for i, t := range *obj.CommercialActivityType {
										types[i] = QuoteAutoResultInsuredObjectCommercialActivityType(t)
									}
									return &types
								}(),
								DepartureDateFromCarDealership: obj.DepartureDateFromCarDealership,
								DriverBetween18And25YearsOldGender: func() *QuoteAutoResultInsuredObjectDriverBetween18And25YearsOldGender {
									if obj.DriverBetween18And25YearsOldGender == nil {
										return nil
									}
									g := QuoteAutoResultInsuredObjectDriverBetween18And25YearsOldGender(*obj.DriverBetween18And25YearsOldGender)
									return &g
								}(),
								EquipmentsAttached: func() *struct {
									EquipmentsAttachedAmount *AmountDetails `json:"equipmentsAttachedAmount,omitempty"`
									IsDesireCoverage         *bool          `json:"isDesireCoverage,omitempty"`
								} {
									if obj.EquipmentsAttached == nil {
										return nil
									}
									return &struct {
										EquipmentsAttachedAmount *AmountDetails `json:"equipmentsAttachedAmount,omitempty"`
										IsDesireCoverage         *bool          `json:"isDesireCoverage,omitempty"`
									}{
										EquipmentsAttachedAmount: obj.EquipmentsAttached.Amount,
										IsDesireCoverage:         obj.EquipmentsAttached.IsDesireCoverage,
									}
								}(),
								FrequentTrafficArea: func() *QuoteAutoResultInsuredObjectFrequentTrafficArea {
									if obj.FrequentTrafficArea == nil {
										return nil
									}
									area := QuoteAutoResultInsuredObjectFrequentTrafficArea(*obj.FrequentTrafficArea)
									return &area
								}(),
								GasKit: func() *struct {
									GasKitAmount     *AmountDetails `json:"gasKitAmount,omitempty"`
									IsDesireCoverage bool           `json:"isDesireCoverage"`
								} {
									if obj.GasKit == nil {
										return nil
									}
									return &struct {
										GasKitAmount     *AmountDetails `json:"gasKitAmount,omitempty"`
										IsDesireCoverage bool           `json:"isDesireCoverage"`
									}{
										GasKitAmount: obj.GasKit.Amount,
										IsDesireCoverage: func() bool {
											if obj.GasKit.IsDesireCoverage == nil {
												return false
											}
											return *obj.GasKit.IsDesireCoverage
										}(),
									}
								}(),
								LoadsCarriedinsured: func() *[]QuoteAutoResultInsuredObjectLoadsCarriedinsured {
									if obj.LoadsCarriedInsured == nil {
										return nil
									}
									loads := make([]QuoteAutoResultInsuredObjectLoadsCarriedinsured, len(*obj.LoadsCarriedInsured))
									for i, l := range *obj.LoadsCarriedInsured {
										loads[i] = QuoteAutoResultInsuredObjectLoadsCarriedinsured(l)
									}
									return &loads
								}(),
								Modality: func() *QuoteAutoResultInsuredObjectModality {
									if obj.Modality == nil {
										return nil
									}
									m := QuoteAutoResultInsuredObjectModality(*obj.Modality)
									return &m
								}(),
								Model: func() *struct {
									Brand           string  `json:"brand"`
									ManufactureYear *string `json:"manufactureYear,omitempty"`
									ModelName       string  `json:"modelName"`
									ModelYear       *string `json:"modelYear,omitempty"`
								} {
									if obj.Model == nil {
										return nil
									}
									return &struct {
										Brand           string  `json:"brand"`
										ManufactureYear *string `json:"manufactureYear,omitempty"`
										ModelName       string  `json:"modelName"`
										ModelYear       *string `json:"modelYear,omitempty"`
									}{
										Brand:           obj.Model.Brand,
										ManufactureYear: obj.Model.ManufactureYear,
										ModelName:       obj.Model.ModelName,
										ModelYear:       obj.Model.ModelYear,
									}
								}(),
								ModelCode: obj.ModelCode,
								RiskLocationInfo: func() *QuoteAutoRiskLocation {
									if obj.RiskLocation == nil {
										return nil
									}
									rl := obj.RiskLocation
									return &QuoteAutoRiskLocation{
										Housing: func() *struct {
											GateType       *QuoteAutoRiskLocationHousingGateType `json:"gateType,omitempty"`
											IsKeptInGarage *bool                                 `json:"isKeptInGarage,omitempty"`
											Type           *QuoteAutoRiskLocationHousingType     `json:"type,omitempty"`
										} {
											if rl.Housing == nil {
												return nil
											}
											return &struct {
												GateType       *QuoteAutoRiskLocationHousingGateType `json:"gateType,omitempty"`
												IsKeptInGarage *bool                                 `json:"isKeptInGarage,omitempty"`
												Type           *QuoteAutoRiskLocationHousingType     `json:"type,omitempty"`
											}{
												Type: func() *QuoteAutoRiskLocationHousingType {
													if rl.Housing.Type == nil {
														return nil
													}
													ht := QuoteAutoRiskLocationHousingType(*rl.Housing.Type)
													return &ht
												}(),
												IsKeptInGarage: rl.Housing.IsKeptInGarage,
												GateType: func() *QuoteAutoRiskLocationHousingGateType {
													if rl.Housing.GateType == nil {
														return nil
													}
													gt := QuoteAutoRiskLocationHousingGateType(*rl.Housing.GateType)
													return &gt
												}(),
											}
										}(),
										IsUsedCollege: rl.IsUsedCollege,
										UsedCollege: func() *struct {
											DistanceFromResidence *string `json:"distanceFromResidence,omitempty"`
											IsKeptInGarage        *bool   `json:"isKeptInGarage,omitempty"`
										} {
											if rl.UsedCollege == nil {
												return nil
											}
											return &struct {
												DistanceFromResidence *string `json:"distanceFromResidence,omitempty"`
												IsKeptInGarage        *bool   `json:"isKeptInGarage,omitempty"`
											}{
												IsKeptInGarage:        rl.UsedCollege.IsKeptInGarage,
												DistanceFromResidence: rl.UsedCollege.DistanceFromResidence,
											}
										}(),
										IsUsedCommuteWork: rl.IsUsedCommuteWork,
										UsedCommuteWork: func() *struct {
											DistanceFromResidence *string `json:"distanceFromResidence,omitempty"`
											IsKeptInGarage        *bool   `json:"isKeptInGarage,omitempty"`
										} {
											if rl.UsedCommuteWork == nil {
												return nil
											}
											return &struct {
												DistanceFromResidence *string `json:"distanceFromResidence,omitempty"`
												IsKeptInGarage        *bool   `json:"isKeptInGarage,omitempty"`
											}{
												IsKeptInGarage:        rl.UsedCommuteWork.IsKeptInGarage,
												DistanceFromResidence: rl.UsedCommuteWork.DistanceFromResidence,
											}
										}(),
										KmAveragePerWeek: rl.KmAveragePerWeek,
									}
								}(),
								RiskManagementSystem: func() *[]QuoteAutoResultInsuredObjectRiskManagementSystem {
									if obj.RiskManagementSystem == nil {
										return nil
									}
									systems := make([]QuoteAutoResultInsuredObjectRiskManagementSystem, len(*obj.RiskManagementSystem))
									for i, s := range *obj.RiskManagementSystem {
										systems[i] = QuoteAutoResultInsuredObjectRiskManagementSystem(s)
									}
									return &systems
								}(),
								TableUsed: func() *QuoteAutoResultInsuredObjectTableUsed {
									if obj.TableUsed == nil {
										return nil
									}
									t := QuoteAutoResultInsuredObjectTableUsed(*obj.TableUsed)
									return &t
								}(),
								Tax: func() *struct {
									Exempt              bool                                 `json:"exempt"`
									ExemptionPercentage *string                              `json:"exemptionPercentage,omitempty"`
									Type                *QuoteAutoResultInsuredObjectTaxType `json:"type,omitempty"`
								} {
									if obj.Tax == nil {
										return nil
									}
									return &struct {
										Exempt              bool                                 `json:"exempt"`
										ExemptionPercentage *string                              `json:"exemptionPercentage,omitempty"`
										Type                *QuoteAutoResultInsuredObjectTaxType `json:"type,omitempty"`
									}{
										Exempt:              obj.Tax.Exempt,
										ExemptionPercentage: obj.Tax.ExemptionPercentage,
										Type: func() *QuoteAutoResultInsuredObjectTaxType {
											if obj.Tax.Type == nil {
												return nil
											}
											t := QuoteAutoResultInsuredObjectTaxType(*obj.Tax.Type)
											return &t
										}(),
									}
								}(),
								ValueDetermined: obj.ValuedDetermined,
								VehicleInvoice: func() *struct {
									VehicleAmount *AmountDetails `json:"vehicleAmount,omitempty"`
									VehicleNumber *string        `json:"vehicleNumber,omitempty"`
								} {
									if obj.VehicleInvoice == nil {
										return nil
									}
									return &struct {
										VehicleAmount *AmountDetails `json:"vehicleAmount,omitempty"`
										VehicleNumber *string        `json:"vehicleNumber,omitempty"`
									}{
										VehicleAmount: obj.VehicleInvoice.Amount,
										VehicleNumber: obj.VehicleInvoice.Number,
									}
								}(),
							}
							return &[]QuoteAutoResultInsuredObject{insuredObject}
						}(),
						InsurerID:              q.Data.InsurerID,
						IsCollectiveStipulated: q.Data.IsCollectiveStipulated,
						PolicyID:               q.Data.PolicyID,
						TermEndDate:            q.Data.TermEndDate,
						TermStartDate:          q.Data.TermStartDate,
						TermType:               ResultQuoteQuoteAutoTermType(q.Data.TermType),
					},
					Quotes: func() []struct {
						Assistances         []QuoteResultAssistance         `json:"assistances"`
						Coverages           *[]QuoteAutoQuoteResultCoverage `json:"coverages,omitempty"`
						InsurerQuoteID      string                          `json:"insurerQuoteId"`
						PremiumInfo         QuoteResultPremium              `json:"premiumInfo"`
						SusepProcessNumbers []string                        `json:"susepProcessNumbers"`
					} {
						if q.Data.Quotes == nil {
							return nil
						}
						quotes := make([]struct {
							Assistances         []QuoteResultAssistance         `json:"assistances"`
							Coverages           *[]QuoteAutoQuoteResultCoverage `json:"coverages,omitempty"`
							InsurerQuoteID      string                          `json:"insurerQuoteId"`
							PremiumInfo         QuoteResultPremium              `json:"premiumInfo"`
							SusepProcessNumbers []string                        `json:"susepProcessNumbers"`
						}, len(*q.Data.Quotes))
						for i, offer := range *q.Data.Quotes {
							quotes[i] = struct {
								Assistances         []QuoteResultAssistance         `json:"assistances"`
								Coverages           *[]QuoteAutoQuoteResultCoverage `json:"coverages,omitempty"`
								InsurerQuoteID      string                          `json:"insurerQuoteId"`
								PremiumInfo         QuoteResultPremium              `json:"premiumInfo"`
								SusepProcessNumbers []string                        `json:"susepProcessNumbers"`
							}{
								InsurerQuoteID:      offer.InsurerQuoteID,
								SusepProcessNumbers: offer.SusepProcessNumbers,
								Assistances: func() []QuoteResultAssistance {
									assistances := make([]QuoteResultAssistance, len(offer.Assistances))
									for j, a := range offer.Assistances {
										assistances[j] = QuoteResultAssistance{
											Type:                    QuoteResultAssistanceType(a.Type),
											Service:                 QuoteResultAssistanceService(a.Service),
											Description:             a.Description,
											AssistancePremiumAmount: a.PremiumAmount,
										}
									}
									return assistances
								}(),
								Coverages: func() *[]QuoteAutoQuoteResultCoverage {
									if len(offer.Coverages) == 0 {
										return nil
									}
									coverages := make([]QuoteAutoQuoteResultCoverage, len(offer.Coverages))
									for j, c := range offer.Coverages {
										coverages[j] = QuoteAutoQuoteResultCoverage{
											Branch:                       c.Branch,
											Code:                         QuoteAutoQuoteResultCoverageCode(c.Code),
											Description:                  c.Description,
											InternalCode:                 c.InternalCode,
											IsSeparateContractingAllowed: c.IsSeparateContractingAllowed,
											GracePeriod:                  c.GracePeriod,
											GracePeriodicity: func() *QuoteAutoQuoteResultCoverageGracePeriodicity {
												if c.GracePeriodicity == nil {
													return nil
												}
												gp := QuoteAutoQuoteResultCoverageGracePeriodicity(*c.GracePeriodicity)
												return &gp
											}(),
											GracePeriodCountingMethod: func() *QuoteAutoQuoteResultCoverageGracePeriodCountingMethod {
												if c.GracePeriodCountingMethod == nil {
													return nil
												}
												gpcm := QuoteAutoQuoteResultCoverageGracePeriodCountingMethod(*c.GracePeriodCountingMethod)
												return &gpcm
											}(),
											GracePeriodStartDate: c.GracePeriodStartDate,
											GracePeriodEndDate:   c.GracePeriodEndDate,
											Deductible: func() *QuoteAutoResultDeductible {
												if c.Deductible == nil {
													return nil
												}
												d := c.Deductible
												return &QuoteAutoResultDeductible{
													Type:               QuoteAutoResultDeductibleType(d.Type),
													TypeAdditionalInfo: d.TypeOthers,
													DeductibleAmount:   d.Amount,
													Description:        d.Description,
													Period:             d.Period,
													Periodicity: func() *QuoteAutoResultDeductiblePeriodicity {
														if d.Periodicity == nil {
															return nil
														}
														p := QuoteAutoResultDeductiblePeriodicity(*d.Periodicity)
														return &p
													}(),
													PeriodCountingMethod: func() *QuoteAutoResultDeductiblePeriodCountingMethod {
														if d.PeriodCountingMethod == nil {
															return nil
														}
														pcm := QuoteAutoResultDeductiblePeriodCountingMethod(*d.PeriodCountingMethod)
														return &pcm
													}(),
													PeriodStartDate: d.PeriodStartDate,
													PeriodEndDate:   d.PeriodEndDate,
												}
											}(),
											POS: POS{
												ApplicationType: POSApplicationType(c.POS.ApplicationType),
												Description: func() string {
													if c.POS.Description == nil {
														return ""
													}
													return *c.POS.Description
												}(),
												MinValue:    c.POS.MinValue,
												MaxValue:    c.POS.MaxValue,
												Percentage:  c.POS.Percentage,
												ValueOthers: c.POS.ValueOthers,
											},
											FullIndemnity: QuoteAutoQuoteResultCoverageFullIndemnity(c.FullIndemnity),
										}
									}
									return &coverages
								}(),
								PremiumInfo: QuoteResultPremium{
									PaymentsQuantity:         offer.Premium.PaymentsQuantity,
									TotalNetAmount:           offer.Premium.TotalNetAmount,
									TotalPremiumAmount:       offer.Premium.TotalAmount,
									IOF:                      offer.Premium.IOF,
									InterestRateOverPayments: offer.Premium.InterestRateOverPayments,
									Coverages: func() []QuoteResultPremiumCoverage {
										premiumCoverages := make([]QuoteResultPremiumCoverage, len(offer.Premium.Coverages))
										for j, pc := range offer.Premium.Coverages {
											premiumCoverages[j] = QuoteResultPremiumCoverage{
												Branch:        pc.Branch,
												Code:          QuoteResultPremiumCoverageCode(pc.Code),
												Description:   pc.Description,
												PremiumAmount: pc.PremiumAmount,
											}
										}
										return premiumCoverages
									}(),
									Payments: func() []QuoteResultPayment {
										payments := make([]QuoteResultPayment, len(offer.Premium.Payments))
										for j, p := range offer.Premium.Payments {
											payments[j] = QuoteResultPayment{
												Amount: p.Amount,
												PaymentType: func() QuoteResultPaymentPaymentType {
													if p.PaymentType == nil {
														return ""
													}
													return QuoteResultPaymentPaymentType(*p.PaymentType)
												}(),
											}
										}
										return payments
									}(),
								},
							}
						}
						return quotes
					}(),
				}
			}(),
		},
		Meta:  *api.NewMeta(),
		Links: *api.NewLinks(s.baseURL + "/request/" + q.ConsentID + "/quote-status"),
	}
	return GetQuoteAuto200JSONResponse{N200QuoteStatusAutoJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
