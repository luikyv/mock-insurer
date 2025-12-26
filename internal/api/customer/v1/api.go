//go:generate oapi-codegen -config=./config.yml -package=v1 -o=./api_gen.go ./swagger.yml
package v1

import (
	"context"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/customer"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL         string
	customerService customer.Service
	consentService  consent.Service
	op              *provider.Provider
}

func NewServer(host string, customerService customer.Service, consentService consent.Service, op *provider.Provider) Server {
	return Server{
		baseURL:         host + "/open-insurance/customers/v1",
		customerService: customerService,
		consentService:  consentService,
		op:              op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	swaggerMiddleware, swaggerVersion := middleware.Swagger(GetSwagger, func(err error) api.Error {
		return api.NewError("INVALID_REQUEST", http.StatusBadRequest, err.Error())
	})

	wrapper := ServerInterfaceWrapper{
		Handler: NewStrictHandlerWithOptions(s, nil, StrictHTTPServerOptions{
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				writeResponseError(w, r, err)
			},
		}),
		HandlerMiddlewares: []MiddlewareFunc{swaggerMiddleware, middleware.FAPIID()},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusBadRequest, err.Error()))
		},
	}

	var handler http.Handler

	handler = http.HandlerFunc(wrapper.CustomersGetPersonalIdentifications)
	handler = middleware.Permission(s.consentService, consent.PermissionResourcesRead, consent.PermissionCustomersPersonalIdentificationsRead)(handler)
	handler = middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, consent.ScopeID)(handler)
	mux.Handle("GET /personal/identifications", handler)

	handler = http.HandlerFunc(wrapper.CustomersGetPersonalQualifications)
	handler = middleware.Permission(s.consentService, consent.PermissionResourcesRead, consent.PermissionCustomersPersonalQualificationRead)(handler)
	handler = middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, consent.ScopeID)(handler)
	mux.Handle("GET /personal/qualifications", handler)

	handler = http.HandlerFunc(wrapper.CustomersGetPersonalComplimentaryInformation)
	handler = middleware.Permission(s.consentService, consent.PermissionResourcesRead, consent.PermissionCustomersPersonalAdditionalInfoRead)(handler)
	handler = middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, consent.ScopeID)(handler)
	mux.Handle("GET /personal/complimentary-information", handler)

	handler = http.HandlerFunc(wrapper.CustomersGetBusinessIdentifications)
	handler = middleware.Permission(s.consentService, consent.PermissionResourcesRead, consent.PermissionCustomersBusinessIdentificationsRead)(handler)
	handler = middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, consent.ScopeID)(handler)
	mux.Handle("GET /business/identifications", handler)

	handler = http.HandlerFunc(wrapper.CustomersGetBusinessQualifications)
	handler = middleware.Permission(s.consentService, consent.PermissionResourcesRead, consent.PermissionCustomersBusinessQualificationRead)(handler)
	handler = middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, consent.ScopeID)(handler)
	mux.Handle("GET /business/qualifications", handler)

	handler = http.HandlerFunc(wrapper.CustomersGetBusinessComplimentaryInformation)
	handler = middleware.Permission(s.consentService, consent.PermissionResourcesRead, consent.PermissionCustomersBusinessAdditionalInfoRead)(handler)
	handler = middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, consent.ScopeID)(handler)
	mux.Handle("GET /business/complimentary-information", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/customers/v1", handler), swaggerVersion
}

func (s Server) CustomersGetPersonalIdentifications(ctx context.Context, req CustomersGetPersonalIdentificationsRequestObject) (CustomersGetPersonalIdentificationsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	ownerID := ctx.Value(api.CtxKeySubject).(string)
	pag := page.NewPagination(nil, nil)
	identifications, err := s.customerService.PersonalIdentifications(ctx, ownerID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponsePersonalCustomersIdentification{
		Data: func() []PersonalIdentificationData {
			data := make([]PersonalIdentificationData, 0, len(identifications.Records))
			for _, ident := range identifications.Records {
				identData := PersonalIdentificationData{
					UpdateDateTime:          ident.Data.UpdateDateTime,
					PersonalID:              ident.Data.PersonalID,
					BrandName:               ident.Data.BrandName,
					CivilName:               ident.Data.CivilName,
					SocialName:              ident.Data.SocialName,
					CpfNumber:               ident.Data.CPF,
					HasBrazilianNationality: ident.Data.HasBrazilianNationality,
					Sex:                     ident.Data.Sex,
					BirthDate:               ident.Data.BirthDate,
					Contact: PersonalContact{
						PostalAddresses: func() []PersonalPostalAddress {
							addresses := make([]PersonalPostalAddress, len(ident.Data.Contact.PostalAddresses))
							for i, addr := range ident.Data.Contact.PostalAddresses {
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
							if ident.Data.Contact.Phones == nil {
								return nil
							}
							phones := make([]CustomerPhone, len(*ident.Data.Contact.Phones))
							for i, phone := range *ident.Data.Contact.Phones {
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
							if ident.Data.Contact.Emails == nil {
								return nil
							}
							emails := make([]CustomerEmail, len(*ident.Data.Contact.Emails))
							for i, email := range *ident.Data.Contact.Emails {
								emails[i] = CustomerEmail{
									Email: email.Email,
								}
							}
							return &emails
						}(),
					},
					CivilStatusCode: func() *EnumCivilStatusCode {
						if ident.Data.CivilStatus != nil {
							civilStatus := EnumCivilStatusCode(*ident.Data.CivilStatus)
							return &civilStatus
						}
						return nil
					}(),
					CivilStatusCodeOthers: ident.Data.CivilStatusOthers,
					CompanyInfo: struct {
						CnpjNumber string `json:"cnpjNumber"`
						Name       string `json:"name"`
					}{
						CnpjNumber: ident.Data.CompanyInfo.CNPJ,
						Name:       ident.Data.CompanyInfo.Name,
					},
					Documents: func() *PersonalDocuments {
						if ident.Data.Documents == nil {
							return nil
						}
						docs := make(PersonalDocuments, len(*ident.Data.Documents))
						for i, doc := range *ident.Data.Documents {
							docs[i] = struct {
								DocumentTypeOthers *string                   `json:"documentTypeOthers,omitempty"`
								ExpirationDate     *timeutil.BrazilDate      `json:"expirationDate,omitempty"`
								IssueLocation      *string                   `json:"issueLocation,omitempty"`
								Number             *string                   `json:"number,omitempty"`
								Type               *EnumPersonalDocumentType `json:"type,omitempty"`
							}{
								Type: func() *EnumPersonalDocumentType {
									if doc.Type != nil {
										t := EnumPersonalDocumentType(*doc.Type)
										return &t
									}
									return nil
								}(),
								Number:         doc.Number,
								ExpirationDate: doc.ExpirationDate,
								IssueLocation:  doc.IssueLocation,
							}
						}
						return &docs
					}(),
					OtherDocuments: func() *OtherPersonalDocuments {
						if ident.Data.OtherDocuments == nil {
							return nil
						}
						return &OtherPersonalDocuments{
							Type:           ident.Data.OtherDocuments.Type,
							Number:         ident.Data.OtherDocuments.Number,
							Country:        ident.Data.OtherDocuments.Country,
							ExpirationDate: ident.Data.OtherDocuments.ExpirationDate,
						}
					}(),
					Filiation: func() *struct {
						CivilName *string            `json:"civilName,omitempty"`
						Type      *EnumFiliationType `json:"type,omitempty"`
					} {
						if ident.Data.Filiation == nil {
							return nil
						}
						return &struct {
							CivilName *string            `json:"civilName,omitempty"`
							Type      *EnumFiliationType `json:"type,omitempty"`
						}{
							CivilName: ident.Data.Filiation.CivilName,
							Type: func() *EnumFiliationType {
								if ident.Data.Filiation.Type != nil {
									t := EnumFiliationType(*ident.Data.Filiation.Type)
									return &t
								}
								return nil
							}(),
						}
					}(),
					IdentificationDetails: func() *struct {
						CivilName *string `json:"civilName,omitempty"`
						CpfNumber *string `json:"cpfNumber,omitempty"`
					} {
						if ident.Data.IdentificationDetails == nil {
							return nil
						}
						return &struct {
							CivilName *string `json:"civilName,omitempty"`
							CpfNumber *string `json:"cpfNumber,omitempty"`
						}{
							CivilName: ident.Data.IdentificationDetails.CivilName,
							CpfNumber: ident.Data.IdentificationDetails.CpfNumber,
						}
					}(),
				}
				data = append(data, identData)
			}
			return data
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/personal/identifications", identifications),
		Meta:  *api.NewPaginatedMeta(identifications),
	}
	return CustomersGetPersonalIdentifications200JSONResponse{OKResponsePersonalCustomersIdentificationJSONResponse(resp)}, nil
}

func (s Server) CustomersGetPersonalComplimentaryInformation(ctx context.Context, req CustomersGetPersonalComplimentaryInformationRequestObject) (CustomersGetPersonalComplimentaryInformationResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	ownerID := ctx.Value(api.CtxKeySubject).(string)
	pag := page.NewPagination(nil, nil)
	informations, err := s.customerService.PersonalComplimentaryInformations(ctx, ownerID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponsePersonalCustomersComplimentaryInformation{
		Data: func() []PersonalComplimentaryInformationData {
			data := make([]PersonalComplimentaryInformationData, 0, len(informations.Records))
			for _, info := range informations.Records {
				if info != nil {
					infoData := PersonalComplimentaryInformationData{
						UpdateDateTime:        info.Data.UpdateDateTime,
						StartDate:             info.Data.StartDate,
						RelationshipBeginning: info.Data.RelationshipBeginning,
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
							}, len(info.Data.ProductsServices))
							for i, ps := range info.Data.ProductsServices {
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
					data = append(data, infoData)
				}
			}
			return data
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/personal/complimentary-information", informations),
		Meta:  *api.NewPaginatedMeta(informations),
	}
	return CustomersGetPersonalComplimentaryInformation200JSONResponse{OKResponsePersonalCustomersComplimentaryInformationJSONResponse(resp)}, nil
}

func (s Server) CustomersGetPersonalQualifications(ctx context.Context, request CustomersGetPersonalQualificationsRequestObject) (CustomersGetPersonalQualificationsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	ownerID := ctx.Value(api.CtxKeySubject).(string)
	pag := page.NewPagination(nil, nil)
	qualifications, err := s.customerService.PersonalQualifications(ctx, ownerID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponsePersonalCustomersQualification{
		Data: func() []PersonalQualificationData {
			data := make([]PersonalQualificationData, 0, len(qualifications.Records))
			for _, qual := range qualifications.Records {
				if qual != nil {
					qualData := PersonalQualificationData{
						UpdateDateTime:    qual.Data.UpdateDateTime,
						PepIdentification: PersonalQualificationDataPepIdentification(qual.Data.PEPIdentification),
						LifePensionPlans:  PersonalQualificationDataLifePensionPlans(qual.Data.LifePensionPlans),
						Occupation: func() *[]struct {
							Details                  *string                                                `json:"details,omitempty"`
							OccupationCode           *string                                                `json:"occupationCode,omitempty"`
							OccupationCodeType       *PersonalQualificationDataOccupationOccupationCodeType `json:"occupationCodeType,omitempty"`
							OccupationCodeTypeOthers *string                                                `json:"occupationCodeTypeOthers,omitempty"`
						} {
							if qual.Data.Occupations == nil {
								return nil
							}
							occupations := make([]struct {
								Details                  *string                                                `json:"details,omitempty"`
								OccupationCode           *string                                                `json:"occupationCode,omitempty"`
								OccupationCodeType       *PersonalQualificationDataOccupationOccupationCodeType `json:"occupationCodeType,omitempty"`
								OccupationCodeTypeOthers *string                                                `json:"occupationCodeTypeOthers,omitempty"`
							}, len(*qual.Data.Occupations))
							for i, occ := range *qual.Data.Occupations {
								occupations[i] = struct {
									Details                  *string                                                `json:"details,omitempty"`
									OccupationCode           *string                                                `json:"occupationCode,omitempty"`
									OccupationCodeType       *PersonalQualificationDataOccupationOccupationCodeType `json:"occupationCodeType,omitempty"`
									OccupationCodeTypeOthers *string                                                `json:"occupationCodeTypeOthers,omitempty"`
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
									OccupationCodeTypeOthers: occ.OccupationCodeTypeOthers,
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
							if qual.Data.InformedRevenue == nil {
								return nil
							}
							return &struct {
								Amount          *string                                           `json:"amount"`
								Currency        *PersonalQualificationDataInformedRevenueCurrency `json:"currency,omitempty"`
								Date            *timeutil.BrazilDate                              `json:"date,omitempty"`
								IncomeFrequency *EnumIncomeFrequency                              `json:"incomeFrequency,omitempty"`
							}{
								Amount: qual.Data.InformedRevenue.Amount,
								Currency: func() *PersonalQualificationDataInformedRevenueCurrency {
									if qual.Data.InformedRevenue.Currency == nil {
										return nil
									}
									c := PersonalQualificationDataInformedRevenueCurrency(*qual.Data.InformedRevenue.Currency)
									return &c
								}(),
								Date: qual.Data.InformedRevenue.Date,
								IncomeFrequency: func() *EnumIncomeFrequency {
									if qual.Data.InformedRevenue.IncomeFrequency == nil {
										return nil
									}
									f := EnumIncomeFrequency(*qual.Data.InformedRevenue.IncomeFrequency)
									return &f
								}(),
							}
						}(),
						InformedPatrimony: func() *struct {
							Amount   *string                                             `json:"amount"`
							Currency *PersonalQualificationDataInformedPatrimonyCurrency `json:"currency,omitempty"`
							Year     *string                                             `json:"year,omitempty"`
						} {
							if qual.Data.InformedPatrimony == nil {
								return nil
							}
							return &struct {
								Amount   *string                                             `json:"amount"`
								Currency *PersonalQualificationDataInformedPatrimonyCurrency `json:"currency,omitempty"`
								Year     *string                                             `json:"year,omitempty"`
							}{
								Amount: qual.Data.InformedPatrimony.Amount,
								Currency: func() *PersonalQualificationDataInformedPatrimonyCurrency {
									if qual.Data.InformedPatrimony.Currency == nil {
										return nil
									}
									c := PersonalQualificationDataInformedPatrimonyCurrency(*qual.Data.InformedPatrimony.Currency)
									return &c
								}(),
								Year: qual.Data.InformedPatrimony.Year,
							}
						}(),
					}
					data = append(data, qualData)
				}
			}
			return data
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/personal/qualifications", qualifications),
		Meta:  *api.NewPaginatedMeta(qualifications),
	}
	return CustomersGetPersonalQualifications200JSONResponse{OKResponsePersonalCustomersQualificationJSONResponse(resp)}, nil
}

func (s Server) CustomersGetBusinessComplimentaryInformation(ctx context.Context, request CustomersGetBusinessComplimentaryInformationRequestObject) (CustomersGetBusinessComplimentaryInformationResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	ownerID := ctx.Value(api.CtxKeySubject).(string)
	pag := page.NewPagination(nil, nil)
	informations, err := s.customerService.BusinessComplimentaryInformations(ctx, ownerID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseBusinessCustomersComplimentaryInformation{
		Data: func() []BusinessComplimentaryInformationData {
			data := make([]BusinessComplimentaryInformationData, 0, len(informations.Records))
			for _, info := range informations.Records {
				if info != nil {
					infoData := BusinessComplimentaryInformationData{
						UpdateDateTime:        info.Data.UpdateDateTime,
						StartDate:             info.Data.StartDate,
						RelationshipBeginning: info.Data.RelationshipBeginning,
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
							}, len(info.Data.ProductsServices))
							for i, ps := range info.Data.ProductsServices {
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
					data = append(data, infoData)
				}
			}
			return data
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/business/complimentary-information", informations),
		Meta:  *api.NewPaginatedMeta(informations),
	}
	return CustomersGetBusinessComplimentaryInformation200JSONResponse{OKResponseBusinessCustomersComplimentaryInformationJSONResponse(resp)}, nil
}

func (s Server) CustomersGetBusinessIdentifications(ctx context.Context, request CustomersGetBusinessIdentificationsRequestObject) (CustomersGetBusinessIdentificationsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	ownerID := ctx.Value(api.CtxKeySubject).(string)
	pag := page.NewPagination(nil, nil)
	identifications, err := s.customerService.BusinessIdentifications(ctx, ownerID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseBusinessCustomersIdentification{
		Data: func() []BusinessIdentificationData {
			data := make([]BusinessIdentificationData, 0, len(identifications.Records))
			for _, ident := range identifications.Records {
				if ident != nil {
					identData := BusinessIdentificationData{
						UpdateDateTime:    ident.Data.UpdateDateTime,
						BusinessID:        ident.Data.BusinessID,
						BrandName:         ident.Data.BrandName,
						BusinessName:      ident.Data.BusinessName,
						BusinessTradeName: ident.Data.BusinessTradeName,
						IncorporationDate: ident.Data.IncorporationDate,
						CompanyInfo: struct {
							CnpjNumber string `json:"cnpjNumber"`
							Name       string `json:"name"`
						}{
							CnpjNumber: ident.Data.CompanyInfo.CNPJ,
							Name:       ident.Data.CompanyInfo.Name,
						},
						Document: BusinessDocument{
							BusinesscnpjNumber:                  ident.Data.Document.CNPJNumber,
							BusinessRegisterNumberOriginCountry: ident.Data.Document.RegistrationNumberOriginCountry,
							ExpirationDate:                      ident.Data.Document.ExpirationDate,
							Country: func() *BusinessDocumentCountry {
								if ident.Data.Document.Country == nil {
									return nil
								}
								c := BusinessDocumentCountry(*ident.Data.Document.Country)
								return &c
							}(),
						},
						Type: func() *BusinessIdentificationDataType {
							if ident.Data.Type == nil {
								return nil
							}
							t := BusinessIdentificationDataType(*ident.Data.Type)
							return &t
						}(),
						Contact: BusinessContact{
							PostalAddresses: func() []BusinessPostalAddress {
								addresses := make([]BusinessPostalAddress, len(ident.Data.Contact.PostalAddresses))
								for i, addr := range ident.Data.Contact.PostalAddresses {
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
											return &GeographicCoordinates{
												Latitude:  addr.GeographicCoordinates.Latitude,
												Longitude: addr.GeographicCoordinates.Longitude,
											}
										}(),
									}
								}
								return addresses
							}(),
							Phones: func() *[]CustomerPhone {
								if ident.Data.Contact.Phones == nil {
									return nil
								}
								phones := make([]CustomerPhone, len(*ident.Data.Contact.Phones))
								for i, phone := range *ident.Data.Contact.Phones {
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
								if ident.Data.Contact.Emails == nil {
									return nil
								}
								emails := make([]CustomerEmail, len(*ident.Data.Contact.Emails))
								for i, email := range *ident.Data.Contact.Emails {
									emails[i] = CustomerEmail{
										Email: email.Email,
									}
								}
								return &emails
							}(),
						},
						Parties: func() *BusinessParties {
							if ident.Data.Parties == nil {
								return nil
							}
							parties := make(BusinessParties, len(*ident.Data.Parties))
							for i, party := range *ident.Data.Parties {
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
					data = append(data, identData)
				}
			}
			return data
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/business/identifications", identifications),
		Meta:  *api.NewPaginatedMeta(identifications),
	}
	return CustomersGetBusinessIdentifications200JSONResponse{OKResponseBusinessCustomersIdentificationJSONResponse(resp)}, nil
}

func (s Server) CustomersGetBusinessQualifications(ctx context.Context, request CustomersGetBusinessQualificationsRequestObject) (CustomersGetBusinessQualificationsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	ownerID := ctx.Value(api.CtxKeySubject).(string)
	pag := page.NewPagination(nil, nil)
	qualifications, err := s.customerService.BusinessQualifications(ctx, ownerID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseBusinessCustomersQualification{
		Data: func() []BusinessQualificationData {
			data := make([]BusinessQualificationData, 0, len(qualifications.Records))
			for _, qual := range qualifications.Records {
				if qual != nil {
					qualData := BusinessQualificationData{
						UpdateDateTime:  qual.Data.UpdateDateTime,
						MainBranch:      qual.Data.MainBranch,
						SecondaryBranch: qual.Data.SecondaryBranch,
						InformedRevenue: func() *struct {
							Amount          *string                                           `json:"amount"`
							Currency        *BusinessQualificationDataInformedRevenueCurrency `json:"currency,omitempty"`
							IncomeFrequency *EnumIncomeFrequency                              `json:"incomeFrequency,omitempty"`
							Year            *string                                           `json:"year,omitempty"`
						} {
							if qual.Data.InformedRevenue == nil {
								return nil
							}
							return &struct {
								Amount          *string                                           `json:"amount"`
								Currency        *BusinessQualificationDataInformedRevenueCurrency `json:"currency,omitempty"`
								IncomeFrequency *EnumIncomeFrequency                              `json:"incomeFrequency,omitempty"`
								Year            *string                                           `json:"year,omitempty"`
							}{
								Amount: qual.Data.InformedRevenue.Amount,
								Currency: func() *BusinessQualificationDataInformedRevenueCurrency {
									if qual.Data.InformedRevenue.Currency == nil {
										return nil
									}
									c := BusinessQualificationDataInformedRevenueCurrency(*qual.Data.InformedRevenue.Currency)
									return &c
								}(),
								IncomeFrequency: func() *EnumIncomeFrequency {
									if qual.Data.InformedRevenue.IncomeFrequency == nil {
										return nil
									}
									f := EnumIncomeFrequency(*qual.Data.InformedRevenue.IncomeFrequency)
									return &f
								}(),
								Year: qual.Data.InformedRevenue.Year,
							}
						}(),
						InformedPatrimony: func() *struct {
							Amount   *string                                             `json:"amount"`
							Currency *BusinessQualificationDataInformedPatrimonyCurrency `json:"currency,omitempty"`
							Date     *timeutil.BrazilDate                                `json:"date,omitempty"`
						} {
							if qual.Data.InformedPatrimony == nil {
								return nil
							}
							return &struct {
								Amount   *string                                             `json:"amount"`
								Currency *BusinessQualificationDataInformedPatrimonyCurrency `json:"currency,omitempty"`
								Date     *timeutil.BrazilDate                                `json:"date,omitempty"`
							}{
								Amount: qual.Data.InformedPatrimony.Amount,
								Currency: func() *BusinessQualificationDataInformedPatrimonyCurrency {
									if qual.Data.InformedPatrimony.Currency == nil {
										return nil
									}
									c := BusinessQualificationDataInformedPatrimonyCurrency(*qual.Data.InformedPatrimony.Currency)
									return &c
								}(),
								Date: qual.Data.InformedPatrimony.Date,
							}
						}(),
					}
					data = append(data, qualData)
				}
			}
			return data
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/business/qualifications", qualifications),
		Meta:  *api.NewPaginatedMeta(qualifications),
	}
	return CustomersGetBusinessQualifications200JSONResponse{OKResponseBusinessCustomersQualificationJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	api.WriteError(w, r, err)
}
