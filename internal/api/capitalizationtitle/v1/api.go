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
	"github.com/luikyv/mock-insurer/internal/capitalizationtitle"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

var _ StrictServerInterface = Server{}

type Server struct {
	baseURL        string
	service        capitalizationtitle.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(
	host string,
	service capitalizationtitle.Service,
	consentService consent.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:        host + "/open-insurance/insurance-capitalization-title/v1",
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	authCodeAuthMiddleware := middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, capitalizationtitle.Scope)
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

	handler = http.HandlerFunc(wrapper.GetInsuranceCapitalizationTitle)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionCapitalizationTitleRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-capitalization-title/plans", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceCapitalizationTitleplanIDEvents)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionCapitalizationTitleEventsRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-capitalization-title/{planId}/events", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceCapitalizationTitleplanIDPlanInfo)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionCapitalizationTitlePlanInfoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-capitalization-title/{planId}/plan-info", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceCapitalizationTitleplanIDSettlement)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionCapitalizationTitleSettlementsRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-capitalization-title/{planId}/settlements", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/insurance-capitalization-title/v1", handler), swaggerVersion
}

func (s Server) GetInsuranceCapitalizationTitle(ctx context.Context, req GetInsuranceCapitalizationTitleRequestObject) (GetInsuranceCapitalizationTitleResponseObject, error) {
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	plans, err := s.service.ConsentedPlans(ctx, consentID, orgID, page.NewPagination(req.Params.Page, req.Params.PageSize))
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceCapitalizationTitle{
		Data: func() []struct {
			Brand struct {
				Companies []struct {
					CnpjNumber  string `json:"cnpjNumber"`
					CompanyName string `json:"companyName"`
					Products    []struct {
						PlanID      string `json:"planId"`
						ProductName string `json:"productName"`
					} `json:"products"`
				} `json:"companies"`
				Name string `json:"name"`
			} `json:"brand"`
		} {
			respProducts := []struct {
				PlanID      string `json:"planId"`
				ProductName string `json:"productName"`
			}{}
			for _, plan := range plans.Records {
				productName := ""
				if len(plan.Data.Series) > 0 && plan.Data.Series[0].CommercialName != nil {
					productName = *plan.Data.Series[0].CommercialName
				}
				respProducts = append(respProducts, struct {
					PlanID      string `json:"planId"`
					ProductName string `json:"productName"`
				}{
					PlanID:      plan.ID.String(),
					ProductName: productName,
				})
			}
			return []struct {
				Brand struct {
					Companies []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Products    []struct {
							PlanID      string `json:"planId"`
							ProductName string `json:"productName"`
						} `json:"products"`
					} `json:"companies"`
					Name string `json:"name"`
				} `json:"brand"`
			}{{
				Brand: struct {
					Companies []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Products    []struct {
							PlanID      string `json:"planId"`
							ProductName string `json:"productName"`
						} `json:"products"`
					} `json:"companies"`
					Name string `json:"name"`
				}{
					Name: insurer.Brand,
					Companies: []struct {
						CnpjNumber  string `json:"cnpjNumber"`
						CompanyName string `json:"companyName"`
						Products    []struct {
							PlanID      string `json:"planId"`
							ProductName string `json:"productName"`
						} `json:"products"`
					}{{
						CnpjNumber:  insurer.CNPJ,
						CompanyName: insurer.Brand,
						Products:    respProducts,
					}},
				},
			}}
		}(),
		Meta:  *api.NewPaginatedMeta(plans),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-capitalization-title/plans", plans),
	}
	return GetInsuranceCapitalizationTitle200JSONResponse{OKResponseInsuranceCapitalizationTitleJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceCapitalizationTitleplanIDEvents(ctx context.Context, req GetInsuranceCapitalizationTitleplanIDEventsRequestObject) (GetInsuranceCapitalizationTitleplanIDEventsResponseObject, error) {
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	events, err := s.service.ConsentedEvents(ctx, req.PlanID, consentID, orgID, page.NewPagination(req.Params.Page, req.Params.PageSize))
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceCapitalizationTitleEvent{
		Data: func() []struct {
			Event *struct {
				Raffle *struct {
					RaffleAmount         AmountDetails       `json:"raffleAmount"`
					RaffleDate           timeutil.BrazilDate `json:"raffleDate"`
					RaffleSettlementDate timeutil.BrazilDate `json:"raffleSettlementDate"`
				} `json:"raffle,omitempty"`
				Redemption *struct {
					RedemptionAmount         AmountDetails                                                  `json:"redemptionAmount"`
					RedemptionBonusAmount    AmountDetails                                                  `json:"redemptionBonusAmount"`
					RedemptionRequestDate    timeutil.BrazilDate                                            `json:"redemptionRequestDate"`
					RedemptionSettlementDate timeutil.BrazilDate                                            `json:"redemptionSettlementDate"`
					RedemptionType           InsuranceCapitalizationTitleEventEventRedemptionRedemptionType `json:"redemptionType"`
					UnreturnedAmount         *AmountDetails                                                 `json:"unreturnedAmount,omitempty"`
				} `json:"redemption,omitempty"`
			} `json:"event,omitempty"`
			EventType *InsuranceCapitalizationTitleEventEventType `json:"eventType,omitempty"`
			TitleID   *string                                     `json:"titleId,omitempty"`
		} {
			respEvents := make([]struct {
				Event *struct {
					Raffle *struct {
						RaffleAmount         AmountDetails       `json:"raffleAmount"`
						RaffleDate           timeutil.BrazilDate `json:"raffleDate"`
						RaffleSettlementDate timeutil.BrazilDate `json:"raffleSettlementDate"`
					} `json:"raffle,omitempty"`
					Redemption *struct {
						RedemptionAmount         AmountDetails                                                  `json:"redemptionAmount"`
						RedemptionBonusAmount    AmountDetails                                                  `json:"redemptionBonusAmount"`
						RedemptionRequestDate    timeutil.BrazilDate                                            `json:"redemptionRequestDate"`
						RedemptionSettlementDate timeutil.BrazilDate                                            `json:"redemptionSettlementDate"`
						RedemptionType           InsuranceCapitalizationTitleEventEventRedemptionRedemptionType `json:"redemptionType"`
						UnreturnedAmount         *AmountDetails                                                 `json:"unreturnedAmount,omitempty"`
					} `json:"redemption,omitempty"`
				} `json:"event,omitempty"`
				EventType *InsuranceCapitalizationTitleEventEventType `json:"eventType,omitempty"`
				TitleID   *string                                     `json:"titleId,omitempty"`
			}, 0, len(events.Records))

			for _, event := range events.Records {
				eventResp := struct {
					Event *struct {
						Raffle *struct {
							RaffleAmount         AmountDetails       `json:"raffleAmount"`
							RaffleDate           timeutil.BrazilDate `json:"raffleDate"`
							RaffleSettlementDate timeutil.BrazilDate `json:"raffleSettlementDate"`
						} `json:"raffle,omitempty"`
						Redemption *struct {
							RedemptionAmount         AmountDetails                                                  `json:"redemptionAmount"`
							RedemptionBonusAmount    AmountDetails                                                  `json:"redemptionBonusAmount"`
							RedemptionRequestDate    timeutil.BrazilDate                                            `json:"redemptionRequestDate"`
							RedemptionSettlementDate timeutil.BrazilDate                                            `json:"redemptionSettlementDate"`
							RedemptionType           InsuranceCapitalizationTitleEventEventRedemptionRedemptionType `json:"redemptionType"`
							UnreturnedAmount         *AmountDetails                                                 `json:"unreturnedAmount,omitempty"`
						} `json:"redemption,omitempty"`
					} `json:"event,omitempty"`
					EventType *InsuranceCapitalizationTitleEventEventType `json:"eventType,omitempty"`
					TitleID   *string                                     `json:"titleId,omitempty"`
				}{
					TitleID: event.Data.TitleID,
					EventType: func() *InsuranceCapitalizationTitleEventEventType {
						if event.Data.Type == nil {
							return nil
						}
						eventType := InsuranceCapitalizationTitleEventEventType(*event.Data.Type)
						return &eventType
					}(),
					Event: func() *struct {
						Raffle *struct {
							RaffleAmount         AmountDetails       `json:"raffleAmount"`
							RaffleDate           timeutil.BrazilDate `json:"raffleDate"`
							RaffleSettlementDate timeutil.BrazilDate `json:"raffleSettlementDate"`
						} `json:"raffle,omitempty"`
						Redemption *struct {
							RedemptionAmount         AmountDetails                                                  `json:"redemptionAmount"`
							RedemptionBonusAmount    AmountDetails                                                  `json:"redemptionBonusAmount"`
							RedemptionRequestDate    timeutil.BrazilDate                                            `json:"redemptionRequestDate"`
							RedemptionSettlementDate timeutil.BrazilDate                                            `json:"redemptionSettlementDate"`
							RedemptionType           InsuranceCapitalizationTitleEventEventRedemptionRedemptionType `json:"redemptionType"`
							UnreturnedAmount         *AmountDetails                                                 `json:"unreturnedAmount,omitempty"`
						} `json:"redemption,omitempty"`
					} {
						if event.Data.Raffle == nil && event.Data.Redemption == nil {
							return nil
						}
						return &struct {
							Raffle *struct {
								RaffleAmount         AmountDetails       `json:"raffleAmount"`
								RaffleDate           timeutil.BrazilDate `json:"raffleDate"`
								RaffleSettlementDate timeutil.BrazilDate `json:"raffleSettlementDate"`
							} `json:"raffle,omitempty"`
							Redemption *struct {
								RedemptionAmount         AmountDetails                                                  `json:"redemptionAmount"`
								RedemptionBonusAmount    AmountDetails                                                  `json:"redemptionBonusAmount"`
								RedemptionRequestDate    timeutil.BrazilDate                                            `json:"redemptionRequestDate"`
								RedemptionSettlementDate timeutil.BrazilDate                                            `json:"redemptionSettlementDate"`
								RedemptionType           InsuranceCapitalizationTitleEventEventRedemptionRedemptionType `json:"redemptionType"`
								UnreturnedAmount         *AmountDetails                                                 `json:"unreturnedAmount,omitempty"`
							} `json:"redemption,omitempty"`
						}{
							Raffle: func() *struct {
								RaffleAmount         AmountDetails       `json:"raffleAmount"`
								RaffleDate           timeutil.BrazilDate `json:"raffleDate"`
								RaffleSettlementDate timeutil.BrazilDate `json:"raffleSettlementDate"`
							} {
								if event.Data.Raffle == nil {
									return nil
								}
								return &struct {
									RaffleAmount         AmountDetails       `json:"raffleAmount"`
									RaffleDate           timeutil.BrazilDate `json:"raffleDate"`
									RaffleSettlementDate timeutil.BrazilDate `json:"raffleSettlementDate"`
								}{
									RaffleAmount:         event.Data.Raffle.Amount,
									RaffleDate:           event.Data.Raffle.Date,
									RaffleSettlementDate: event.Data.Raffle.SettlementDate,
								}
							}(),
							Redemption: func() *struct {
								RedemptionAmount         AmountDetails                                                  `json:"redemptionAmount"`
								RedemptionBonusAmount    AmountDetails                                                  `json:"redemptionBonusAmount"`
								RedemptionRequestDate    timeutil.BrazilDate                                            `json:"redemptionRequestDate"`
								RedemptionSettlementDate timeutil.BrazilDate                                            `json:"redemptionSettlementDate"`
								RedemptionType           InsuranceCapitalizationTitleEventEventRedemptionRedemptionType `json:"redemptionType"`
								UnreturnedAmount         *AmountDetails                                                 `json:"unreturnedAmount,omitempty"`
							} {
								if event.Data.Redemption == nil {
									return nil
								}
								return &struct {
									RedemptionAmount         AmountDetails                                                  `json:"redemptionAmount"`
									RedemptionBonusAmount    AmountDetails                                                  `json:"redemptionBonusAmount"`
									RedemptionRequestDate    timeutil.BrazilDate                                            `json:"redemptionRequestDate"`
									RedemptionSettlementDate timeutil.BrazilDate                                            `json:"redemptionSettlementDate"`
									RedemptionType           InsuranceCapitalizationTitleEventEventRedemptionRedemptionType `json:"redemptionType"`
									UnreturnedAmount         *AmountDetails                                                 `json:"unreturnedAmount,omitempty"`
								}{
									RedemptionAmount:         event.Data.Redemption.Amount,
									RedemptionBonusAmount:    event.Data.Redemption.BonusAmount,
									RedemptionRequestDate:    event.Data.Redemption.Date,
									RedemptionSettlementDate: event.Data.Redemption.SettlementDate,
									RedemptionType:           InsuranceCapitalizationTitleEventEventRedemptionRedemptionType(event.Data.Redemption.Type),
									UnreturnedAmount:         event.Data.Redemption.UnreturnedAmount,
								}
							}(),
						}
					}(),
				}

				respEvents = append(respEvents, eventResp)
			}
			return respEvents
		}(),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-capitalization-title/"+req.PlanID+"/events", events),
		Meta:  *api.NewPaginatedMeta(events),
	}
	return GetInsuranceCapitalizationTitleplanIDEvents200JSONResponse{OKResponseInsuranceCapitalizationTitleEventJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceCapitalizationTitleplanIDPlanInfo(ctx context.Context, req GetInsuranceCapitalizationTitleplanIDPlanInfoRequestObject) (GetInsuranceCapitalizationTitleplanIDPlanInfoResponseObject, error) {
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	plan, err := s.service.ConsentedPlan(ctx, req.PlanID, consentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceCapitalizationTitlePlanInfo{
		Data: func() InsuranceCapitalizationTitlePlanInfo {
			planIDStr := plan.ID.String()
			return InsuranceCapitalizationTitlePlanInfo{
				Series: func() []InsuranceCapitalizationTitleSeries {
					series := make([]InsuranceCapitalizationTitleSeries, 0, len(plan.Data.Series))
					for _, s := range plan.Data.Series {
						seriesResp := InsuranceCapitalizationTitleSeries{
							SeriesID:                     s.ID,
							Modality:                     InsuranceCapitalizationTitleSeriesModality(s.Modality),
							SusepProcessNumber:           s.SusepProcessNumber,
							CommercialName:               s.CommercialName,
							SerieSize:                    s.SerieSize,
							GracePeriodRedemption:        s.GracePeriodRedemption,
							GracePeriodForFullRedemption: s.GracePeriodForFullRedemption,
							UpdateIndex:                  InsuranceCapitalizationTitleSeriesUpdateIndex(s.UpdateIndex),
							UpdateIndexOthers:            s.UpdateIndexOthers,
							ReadjustmentIndex:            InsuranceCapitalizationTitleSeriesReadjustmentIndex(s.ReadjustmentIndex),
							ReadjustmentIndexOthers:      s.ReadjustmentIndexOthers,
							BonusClause:                  s.BonusClause,
							Frequency:                    InsuranceCapitalizationTitleSeriesFrequency(s.Frequency),
							FrequencyDescription:         s.FrequencyDescription,
							InterestRate:                 PercentageDetails(s.InterestRate),
							PlanID:                       &planIDStr,
							Quotas: func() []CapitalizationTitleQuotas {
								quotas := make([]CapitalizationTitleQuotas, len(s.Quotas))
								for i, q := range s.Quotas {
									quotas[i] = CapitalizationTitleQuotas{
										Quota:               q.Number,
										CapitalizationQuota: q.CapitalizationQuota,
										ChargingQuota:       q.ChargingQuota,
										RaffleQuota:         q.RaffleQuota,
									}
								}
								return quotas
							}(),
							Broker: func() *[]InsuranceCapitalizationTitleBroker {
								if s.Brokers == nil {
									return nil
								}
								brokers := make([]InsuranceCapitalizationTitleBroker, len(*s.Brokers))
								for i, b := range *s.Brokers {
									brokers[i] = InsuranceCapitalizationTitleBroker{
										SusepBrokerCode:   b.SusepBrokerCode,
										BrokerDescription: b.BrokerDescription,
									}
								}
								return &brokers
							}(),
							Titles: func() []InsuranceCapitalizationTitleTitle {
								titles := make([]InsuranceCapitalizationTitleTitle, len(s.Titles))
								for i, t := range s.Titles {
									titles[i] = InsuranceCapitalizationTitleTitle{
										TitleID:             t.ID,
										RegistrationForm:    t.RegistrationForm,
										IssueTitleDate:      t.IssueTitleDate,
										TermStartDate:       t.TermStartDate,
										TermEndDate:         t.TermEndDate,
										RafflePremiumAmount: t.RafflePremiumAmount,
										ContributionAmount:  t.ContributionAmount,
										Subscriber: func() []InsuranceCapitalizationTitleSubscriber {
											subscribers := make([]InsuranceCapitalizationTitleSubscriber, len(t.Subscribers))
											for j, sub := range t.Subscribers {
												subscribers[j] = InsuranceCapitalizationTitleSubscriber{
													SubscriberName:                  sub.Name,
													SubscriberDocumentType:          InsuranceCapitalizationTitleSubscriberSubscriberDocumentType(sub.DocumentType),
													SubscriberDocumentTypeOthers:    sub.DocumentTypeOthers,
													SubscriberDocumentNumber:        sub.DocumentNumber,
													SubscriberAddress:               sub.Address,
													SubscriberAddressAdditionalInfo: sub.AddressAdditionalInfo,
													SubscriberTownName:              sub.TownName,
													SubscriberCountrySubDivision:    EnumCountrySubDivision(sub.CountrySubDivision),
													SubscriberCountryCode:           sub.CountryCode,
													SubscriberPostCode:              sub.PostCode,
													SubscriberPhones: func() *[]RequestorPhone {
														if sub.Phones == nil {
															return nil
														}
														phones := make([]RequestorPhone, len(*sub.Phones))
														for k, p := range *sub.Phones {
															phones[k] = RequestorPhone{
																CountryCallingCode: p.CountryCallingCode,
																Number:             p.Number,
																AreaCode: func() *EnumAreaCode {
																	if p.AreaCode == nil {
																		return nil
																	}
																	areaCode := EnumAreaCode(*p.AreaCode)
																	return &areaCode
																}(),
															}
														}
														return &phones
													}(),
													Holder: func() *[]InsuranceCapitalizationTitleHolder {
														if sub.Holders == nil {
															return nil
														}
														holders := make([]InsuranceCapitalizationTitleHolder, len(*sub.Holders))
														for k, h := range *sub.Holders {
															holders[k] = InsuranceCapitalizationTitleHolder{
																HolderName:                  h.Name,
																HolderDocumentType:          InsuranceCapitalizationTitleHolderHolderDocumentType(h.DocumentType),
																HolderDocumentTypeOthers:    h.DocumentTypeOthers,
																HolderDocumentNumber:        h.DocumentNumber,
																HolderAddress:               h.Address,
																HolderAddressAdditionalInfo: h.AddressAdditionalInfo,
																HolderTownName:              h.TownName,
																HolderCountrySubDivision:    EnumCountrySubDivision(h.CountrySubDivision),
																HolderCountryCode:           h.CountryCode,
																HolderPostCode:              h.PostCode,
																HolderRedemption:            h.Redemption,
																HolderRaffle:                h.Raffle,
																HolderPhones: func() *[]RequestorPhone {
																	if h.Phones == nil {
																		return nil
																	}
																	holderPhones := make([]RequestorPhone, len(*h.Phones))
																	for l, p := range *h.Phones {
																		holderPhones[l] = RequestorPhone{
																			CountryCallingCode: p.CountryCallingCode,
																			Number:             p.Number,
																			AreaCode: func() *EnumAreaCode {
																				if p.AreaCode == nil {
																					return nil
																				}
																				areaCode := EnumAreaCode(*p.AreaCode)
																				return &areaCode
																			}(),
																		}
																	}
																	return &holderPhones
																}(),
															}
														}
														return &holders
													}(),
												}
											}
											return subscribers
										}(),
										TechnicalProvisions: func() []InsuranceCapitalizationTitleTechnicalProvisions {
											technicalProvisions := make([]InsuranceCapitalizationTitleTechnicalProvisions, len(t.TechnicalProvisions))
											for j, tp := range t.TechnicalProvisions {
												technicalProvisions[j] = InsuranceCapitalizationTitleTechnicalProvisions{
													PmcAmount: tp.PMCAmount,
													PrAmount:  tp.PRAmount,
													PspAmount: tp.PSPAmount,
													PdbAmount: tp.PDBAmount,
												}
											}
											return technicalProvisions
										}(),
									}
								}
								return titles
							}(),
						}
						series = append(series, seriesResp)
					}
					return series
				}(),
			}
		}(),
		Meta:  *api.NewMeta(),
		Links: *api.NewLinks(s.baseURL + "/insurance-capitalization-title/" + req.PlanID + "/plan-info"),
	}
	return GetInsuranceCapitalizationTitleplanIDPlanInfo200JSONResponse{OKResponseInsuranceCapitalizationTitlePlanInfoJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceCapitalizationTitleplanIDSettlement(ctx context.Context, req GetInsuranceCapitalizationTitleplanIDSettlementRequestObject) (GetInsuranceCapitalizationTitleplanIDSettlementResponseObject, error) {
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	settlements, err := s.service.ConsentedSettlements(ctx, req.PlanID, consentID, orgID, page.NewPagination(req.Params.Page, req.Params.PageSize))
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceCapitalizationTitleSettlement{
		Data: func() []struct {
			SettlementDueDate         timeutil.BrazilDate `json:"settlementDueDate"`
			SettlementFinancialAmount AmountDetails       `json:"settlementFinancialAmount"`
			SettlementID              string              `json:"settlementId"`
			SettlementPaymentDate     timeutil.BrazilDate `json:"settlementPaymentDate"`
		} {
			respSettlements := make([]struct {
				SettlementDueDate         timeutil.BrazilDate `json:"settlementDueDate"`
				SettlementFinancialAmount AmountDetails       `json:"settlementFinancialAmount"`
				SettlementID              string              `json:"settlementId"`
				SettlementPaymentDate     timeutil.BrazilDate `json:"settlementPaymentDate"`
			}, 0, len(settlements.Records))
			for _, settlement := range settlements.Records {
				respSettlements = append(respSettlements, struct {
					SettlementDueDate         timeutil.BrazilDate `json:"settlementDueDate"`
					SettlementFinancialAmount AmountDetails       `json:"settlementFinancialAmount"`
					SettlementID              string              `json:"settlementId"`
					SettlementPaymentDate     timeutil.BrazilDate `json:"settlementPaymentDate"`
				}{
					SettlementID:              settlement.ID.String(),
					SettlementFinancialAmount: settlement.Data.FinancialAmount,
					SettlementPaymentDate:     settlement.Data.PaymentDate,
					SettlementDueDate:         settlement.Data.DueDate,
				})
			}
			return respSettlements
		}(),
		Meta:  *api.NewPaginatedMeta(settlements),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-capitalization-title/"+req.PlanID+"/settlements", settlements),
	}
	return GetInsuranceCapitalizationTitleplanIDSettlement200JSONResponse{OKResponseInsuranceCapitalizationTitleSettlementJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
