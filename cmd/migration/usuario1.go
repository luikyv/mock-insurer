package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/auto"
	"github.com/luikyv/mock-insurer/internal/capitalizationtitle"
	"github.com/luikyv/mock-insurer/internal/customer"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"github.com/luikyv/mock-insurer/internal/user"
	"gorm.io/gorm"
)

// nolint:cyclop
func seedUsuario1(ctx context.Context, db *gorm.DB) error {
	testUser := &user.User{
		ID:        uuid.MustParse("ff8cd4db-a1c8-4966-a9ca-26ab0b19c6d1"),
		Username:  "usuario1@seguradoramodelo.com.br",
		Name:      "Usuário 1",
		CPF:       "76109277673",
		CNPJ:      pointerOf("50685362006773"),
		CrossOrg:  true,
		UpdatedAt: timeutil.DateTimeNow(),
		OrgID:     OrgID,
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testUser).Error; err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}

	testCustomerPersonalIdentification := &customer.PersonalIdentification{
		ID:      uuid.MustParse("f5e00af4-8753-47ef-b6f9-a4d6df8e5ab2"),
		OwnerID: testUser.ID,
		Data: customer.PersonalIdentificationData{
			UpdateDateTime: mustParseDateTime("2021-05-21T08:30:00Z"),
			PersonalID:     pointerOf("578-psd-71md6971kjh-2d414"),
			BrandName:      "Organização A",
			CivilName:      "Juan Kaique Cláudio Fernandes",
			SocialName:     pointerOf("string"),
			CPF:            "03320847094",
			CompanyInfo: customer.CompanyInfo{
				CNPJ: "01773247000563",
				Name: "Empresa da Organização A",
			},
			Documents: &[]customer.PersonalDocument{
				{
					Type:           pointerOf(customer.PersonalDocumentTypeCNH),
					Number:         pointerOf("15291908"),
					ExpirationDate: pointerOf(mustParseBrazilDate("2023-05-21")),
					IssueLocation:  pointerOf("string"),
				},
			},
			HasBrazilianNationality: pointerOf(false),
			OtherDocuments: &customer.OtherPersonalDocument{
				Type:           pointerOf("SOCIAL SEC"),
				Number:         pointerOf("15291908"),
				Country:        pointerOf("string"),
				ExpirationDate: pointerOf(mustParseBrazilDate("2023-05-21")),
			},
			Contact: customer.PersonalContact{
				PostalAddresses: []customer.PersonalPostalAddress{
					{
						Address:            "Av Naburo Ykesaki, 1270",
						AdditionalInfo:     pointerOf("Fundos"),
						DistrictName:       pointerOf("Centro"),
						TownName:           "Marília",
						CountrySubDivision: insurer.CountrySubDivision("SP"),
						PostCode:           "17500001",
						Country:            insurer.CountryCode("AFG"),
					},
				},
				Phones: &[]customer.Phone{
					{
						CountryCallingCode: pointerOf("55"),
						AreaCode:           pointerOf(insurer.PhoneAreaCode("19")),
						Number:             pointerOf("29875132"),
						PhoneExtension:     pointerOf("932"),
					},
				},
				Emails: &[]customer.Email{
					{
						Email: pointerOf("nome@br.net"),
					},
				},
			},
			CivilStatus:       pointerOf(insurer.CivilStatusSingle),
			CivilStatusOthers: pointerOf("string"),
			Sex:               pointerOf("FEMININO"),
			BirthDate:         pointerOf(mustParseBrazilDate("2021-05-21")),
			Filiation: &customer.Filiation{
				Type:      pointerOf(customer.FiliationTypeFather),
				CivilName: pointerOf("Marcelo Cláudio Fernandes"),
			},
			IdentificationDetails: &customer.IdentificationDetails{
				CivilName: pointerOf("Juan Kaique Cláudio Fernandes"),
				CpfNumber: pointerOf("44725754465"),
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testCustomerPersonalIdentification).Error; err != nil {
		return fmt.Errorf("failed to create test customer personal identification: %w", err)
	}

	testCustomerPersonalQualification := &customer.PersonalQualification{
		ID:      uuid.MustParse("633acb59-d8bc-48ff-ab2f-199e219872fa"),
		OwnerID: testUser.ID,
		Data: customer.PersonalQualificationData{
			UpdateDateTime:    mustParseDateTime("2021-05-21T08:30:00Z"),
			PEPIdentification: customer.PEPIdentificationNotExposed,
			Occupations: &[]customer.Occupation{
				{
					Details:                  pointerOf("string"),
					OccupationCode:           pointerOf("RECEITA_FEDERAL"),
					OccupationCodeType:       pointerOf(customer.OccupationCodeTypeRFB),
					OccupationCodeTypeOthers: pointerOf("string"),
				},
			},
			LifePensionPlans: "SIM",
			InformedRevenue: &customer.PersonalInformedRevenue{
				IncomeFrequency: pointerOf(customer.IncomeFrequencyMonthly),
				Currency:        pointerOf(insurer.CurrencyBRL),
				Amount:          pointerOf("100000.04"),
				Date:            pointerOf(mustParseBrazilDate("2012-05-21")),
			},
			InformedPatrimony: &customer.PersonalInformedPatrimony{
				Currency: pointerOf(insurer.CurrencyBRL),
				Amount:   pointerOf("100000.04"),
				Year:     pointerOf("2010"),
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testCustomerPersonalQualification).Error; err != nil {
		return fmt.Errorf("failed to create test customer personal qualification: %w", err)
	}

	testCustomerPersonalComplimentaryInformation := &customer.PersonalComplimentaryInformation{
		ID:      uuid.MustParse("c77607bd-f8f7-4729-a16c-1ac7e100c2a2"),
		OwnerID: testUser.ID,
		Data: customer.PersonalComplimentaryInformationData{
			UpdateDateTime:        mustParseDateTime("2021-05-21T08:30:00Z"),
			StartDate:             mustParseBrazilDate("2014-05-21"),
			RelationshipBeginning: pointerOf(mustParseBrazilDate("2014-05-21")),
			ProductsServices: []customer.ProductsAndServices{
				{
					Contract:          "string",
					Type:              customer.ProductServiceTypeMicroinsurance,
					InsuranceLineCode: pointerOf("6272"),
					Procurators: &[]customer.Procurator{
						{
							Nature:     customer.ProcuratorNatureProcurator,
							CpfNumber:  pointerOf("73677831148"),
							CivilName:  pointerOf("Elza Milena Stefany Teixeira"),
							SocialName: pointerOf("string"),
						},
					},
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testCustomerPersonalComplimentaryInformation).Error; err != nil {
		return fmt.Errorf("failed to create test customer personal complimentary information: %w", err)
	}

	testAutoPolicy := &auto.Policy{
		ID: "111111",
		Data: auto.PolicyData{
			ProductName:         "Auto Policy",
			DocumentType:        auto.DocumentTypeIndividual,
			SusepProcessNumber:  pointerOf("string"),
			GroupCertificateID:  pointerOf("string"),
			IssuanceType:        auto.IssuanceTypeOwn,
			IssuanceDate:        mustParseBrazilDate("2022-12-31"),
			TermStartDate:       mustParseBrazilDate("2022-12-31"),
			TermEndDate:         mustParseBrazilDate("2022-12-31"),
			LeadInsurerCode:     pointerOf("string"),
			LeadInsurerPolicyID: pointerOf("string"),
			MaxLMG: insurer.AmountDetails{
				Amount:         "100.00",
				UnitType:       "PORCENTAGEM",
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        "R$",
					Description: "BRL",
				},
			},
			ProposalID: "string",
			Insureds: []auto.Insured{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					BirthDate:                mustParseBrazilDate("1990-06-12"),
					PostCode:                 "10000000",
					Email:                    pointerOf("KdFp4feT>D@c9l_\\b6^}gR.NE&_]K1Og:'8u5\\\"H`jz-5'~%:IY4W@zY$j.+seQawdVy/5Nw</556-+&`et'Wxk^LePe*aa'a3/FJ-0Ur-[4z/8uxYow}KZ@$U\\((EP,[</@+15oeGikTVlS)c96I\\"),
					City:                     "string",
					State:                    "AC",
					Country:                  "BRA",
					Address:                  "string",
				},
			},
			Beneficiaries: &[]auto.Beneficiary{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
				},
			},
			Principals: &[]auto.Principal{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 "10000000",
					Email:                    pointerOf("RCklX-EaoeoO:ZV=S%J]LZOcA!xuPok[r_2#at0f9zl13Vr0pTILkyA-!.Mix7mtnAN54OqYH0w>j2+/p4\"sw\\'5*67RU@DXt.k]t:M8P%mr.uk&I!(5qxbDt#yV;H;KjR1oc%?@Ausmm3_hiQ4NxuB^V62DR$e|CSIaf0hy~{.W*=pWJ8c5rd"),
					City:                     "string",
					State:                    "AC",
					Country:                  "BRA",
					Address:                  "string",
					AddressAdditionalInfo:    pointerOf("string"),
				},
			},
			Intermediaries: &[]auto.Intermediary{
				{
					Type:                     auto.IntermediaryTypeRepresentative,
					TypeOthers:               pointerOf("string"),
					Identification:           pointerOf("12345678900"),
					BrokerID:                 pointerOf("644587421"),
					IdentificationType:       pointerOf(insurer.IdentificationTypeCPF),
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 pointerOf("10000000"),
					City:                     pointerOf("string"),
					State:                    pointerOf("AC"),
					Country:                  pointerOf("BRA"),
					Address:                  pointerOf("string"),
				},
			},
			InsuredObjects: []auto.InsuredObject{
				{
					Identification:                   "string",
					IdentificationType:               auto.AUTOMOVEL,
					IdentificationTypeAdditionalInfo: pointerOf("string"),
					Description:                      "string",
					HasExactVehicleIdentification:    pointerOf(true),
					Modality:                         pointerOf(auto.InsuredObjectModalityValueDetermined),
					ModalityOthers:                   pointerOf("string"),
					AmountReferenceTable:             pointerOf(auto.AmountReferenceTableMolicar),
					AmountReferenceTableOthers:       pointerOf("string"),
					Model:                            pointerOf("string"),
					Year:                             pointerOf("2024"),
					FareCategory:                     pointerOf(auto.FareCategory10),
					RiskPostCode:                     pointerOf("10000000"),
					VehicleUsage:                     pointerOf(auto.VehicleUsageLeisure),
					VehicleUsageOthers:               pointerOf("string"),
					FrequentDestinationPostCode:      pointerOf("10000000"),
					OvernightPostCode:                pointerOf("10000000"),
					Coverages: []auto.InsuredObjectCoverage{
						{
							Branch:             "0111",
							Code:               "CASCO_COMPREENSIVA",
							Description:        pointerOf("string"),
							InternalCode:       pointerOf("string"),
							SusepProcessNumber: "string",
							LMI: insurer.AmountDetails{
								Amount:         "5486585.13",
								UnitType:       "PORCENTAGEM",
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        "R$",
									Description: "BRL",
								},
							},
							TermStartDate:             mustParseBrazilDate("2022-12-31"),
							TermEndDate:               mustParseBrazilDate("2022-12-31"),
							IsMainCoverage:            true,
							Feature:                   auto.CoverageFeatureMass,
							Type:                      auto.CoverageTypeParametric,
							GracePeriod:               pointerOf(0),
							GracePeriodicity:          pointerOf(insurer.PeriodicityDay),
							GracePeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
							GracePeriodStartDate:      pointerOf(mustParseBrazilDate("2022-12-31")),
							GracePeriodEndDate:        pointerOf(mustParseBrazilDate("2022-12-31")),
							AdjustmentRate:            pointerOf("10.00"),
							PremiumAmount: insurer.AmountDetails{
								Amount:         "829276",
								UnitType:       "PORCENTAGEM",
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        "R$",
									Description: "BRL",
								},
							},
							PremiumPeriodicity:            auto.PremiumPeriodicityMonthly,
							PremiumPeriodicityOthers:      pointerOf("string"),
							CompensationType:              pointerOf(auto.CoverageCompensationTypePartial),
							CompensationTypeOthers:        pointerOf("string"),
							PartialCompensationPercentage: pointerOf("10.00"),
							PercentageOverLMI:             pointerOf("10.00"),
							DaysForTotalCompensation:      pointerOf(0),
							BoundCoverage:                 pointerOf(auto.BoundCoverageVehicle),
							BoundCoverageOthers:           pointerOf("string"),
						},
					},
				},
			},
			Coverages: []auto.Coverage{
				{
					Branch:      pointerOf("0111"),
					Code:        auto.CoverageCodeComprehensive,
					Description: pointerOf("string"),
					Deductible: &auto.CoverageDeductible{
						Type:       auto.CoverageDeductibleTypeReduced,
						TypeOthers: pointerOf("string"),
						Amount: &insurer.AmountDetails{
							Amount:         "19.72",
							UnitType:       "PORCENTAGEM",
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        "R$",
								Description: "BRL",
							},
						},
						Period:                             pointerOf(10),
						Periodicity:                        pointerOf(insurer.PeriodicityDay),
						PeriodCountingMethod:               pointerOf(insurer.PeriodCountingMethodBusinessDays),
						PeriodStartDate:                    pointerOf(mustParseBrazilDate("2022-05-16")),
						PeriodEndDate:                      pointerOf(mustParseBrazilDate("2022-05-17")),
						Description:                        pointerOf("Franquia de exemplo"),
						HasDeductibleOverTotalCompensation: pointerOf(true),
					},
					POS: &auto.CoveragePOS{
						ApplicationType: insurer.ValueTypeValue,
						Description:     pointerOf("string"),
						MinValue: &insurer.AmountDetails{
							Amount:         "7488583.06",
							UnitType:       "PORCENTAGEM",
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        "R$",
								Description: "BRL",
							},
						},
						MaxValue: &insurer.AmountDetails{
							Amount:         "14.48",
							UnitType:       "PORCENTAGEM",
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        "R$",
								Description: "BRL",
							},
						},
						Percentage: &insurer.AmountDetails{
							Amount:         "99384484480.01",
							UnitType:       "PORCENTAGEM",
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        "R$",
								Description: "BRL",
							},
						},
						ValueOthers: &insurer.AmountDetails{
							Amount:         "57794",
							UnitType:       "PORCENTAGEM",
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        "R$",
								Description: "BRL",
							},
						},
					},
				},
			},
			CoinsuranceRetainedPercentage: pointerOf("10.00"),
			Coinurers: &[]auto.Coinsurer{
				{
					Identification:  "string",
					CededPercentage: "10.00",
				},
			},
			RepairNetwork:               auto.RepairNetworkReferred,
			RepairNetworkOthers:         pointerOf("string"),
			RepairedPartsUsageType:      auto.RepairedPartsUsageTypeNewAndUsed,
			RepairedPartsClassification: auto.RepairedPartsClassificationOriginalAndCompatible,
			RepairedPartsNationality:    auto.RepairedPartsNationalityNationalAndImported,
			ValidityType:                insurer.ValidityTypeSemestralIntermittent,
			ValidateTypeOthers:          pointerOf("string"),
			OtherCompensations:          pointerOf("string"),
			OtherBenefits:               pointerOf(auto.OtherBenefitsDiscounts),
			AssistancePackages:          pointerOf(auto.AssistancePackagesUpTo10Services),
			IsExpiredRiskPolicy:         pointerOf(true),
			BonusDiscountRate:           pointerOf("100.000"),
			BonusClass:                  pointerOf("string"),
			Drivers: &[]auto.Driver{
				{
					Identification:     pointerOf("12345678900"),
					Sex:                pointerOf(auto.SexFemale),
					SexOthers:          pointerOf("string"),
					BirthDate:          pointerOf(mustParseBrazilDate("2022-12-31")),
					LicensedExperience: pointerOf(4),
				},
			},
			Premium: auto.PremiumData{
				PaymentsQuantity: "4",
				Amount: insurer.AmountDetails{
					Amount:         "87381.35",
					UnitType:       "PORCENTAGEM",
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        "R$",
						Description: "BRL",
					},
				},
				Coverages: []auto.PremiumCoverage{
					{
						Branch:      "0111",
						Code:        auto.CoverageCodeComprehensive,
						Description: pointerOf("string"),
						PremiumAmount: insurer.AmountDetails{
							Amount:         "734876.20",
							UnitType:       "PORCENTAGEM",
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        "R$",
								Description: "BRL",
							},
						},
					},
				},
				Payments: []auto.Payment{
					{
						MovementDate:           mustParseBrazilDate("2022-12-31"),
						MovementType:           auto.PaymentMovementTypePremiumLiquidation,
						MovementOrigin:         pointerOf(auto.PaymentMovementOriginDirectIssuance),
						MovementPaymentsNumber: 0,
						Amount: insurer.AmountDetails{
							Amount:         "57468.28",
							UnitType:       "PORCENTAGEM",
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        "R$",
								Description: "BRL",
							},
						},
						MaturityDate:             mustParseBrazilDate("2022-12-31"),
						TellerID:                 pointerOf("string"),
						TellerIDType:             pointerOf(insurer.IdentificationTypeCPF),
						TellerIDOthers:           pointerOf("RNE"),
						TellerName:               pointerOf("string"),
						FinancialInstitutionCode: pointerOf("string"),
						PaymentType:              pointerOf(auto.PaymentTypeBankSlip),
						PaymentTypeOthers:        pointerOf("string"),
					},
				},
			},
		},
		OwnerID:   testUser.ID,
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testAutoPolicy).Error; err != nil {
		return fmt.Errorf("failed to create test auto policy: %w", err)
	}

	testCapitalizationTitlePlan := &capitalizationtitle.Plan{
		ID:      uuid.MustParse("da3b6fc7-b47b-4cc0-bf1b-0cd1a50f585b"),
		OwnerID: testUser.ID,
		Data: capitalizationtitle.PlanData{
			Series: []capitalizationtitle.PlanSeries{
				{
					ID:                 "string",
					Modality:           capitalizationtitle.ModalityTraditional,
					SusepProcessNumber: "15414622222222222",
					CommercialName:     pointerOf("Denominação comercial do produto"),
					SerieSize:          5000000,
					Quotas: []capitalizationtitle.Quota{
						{
							Number:              10,
							CapitalizationQuota: "0.000002",
							RaffleQuota:         "0.000002",
							ChargingQuota:       "0.000002",
						},
					},
					GracePeriodRedemption:        pointerOf(48),
					GracePeriodForFullRedemption: 48,
					UpdateIndex:                  capitalizationtitle.IndexIPCA,
					UpdateIndexOthers:            pointerOf("Índice de atualização Outros"),
					ReadjustmentIndex:            capitalizationtitle.IndexIPCA,
					ReadjustmentIndexOthers:      pointerOf("Índice de reajuste Outros"),
					BonusClause:                  false,
					Frequency:                    capitalizationtitle.FrequencyMonthly,
					FrequencyDescription:         pointerOf("string"),
					InterestRate:                 "10.00",
					Brokers: &[]capitalizationtitle.Broker{
						{
							SusepBrokerCode:   "123123123",
							BrokerDescription: "string",
						},
					},
					Titles: []capitalizationtitle.Title{
						{
							ID:               "string",
							RegistrationForm: "string",
							IssueTitleDate:   mustParseBrazilDate("2023-01-30"),
							TermStartDate:    mustParseBrazilDate("2023-01-30"),
							TermEndDate:      mustParseBrazilDate("2023-01-30"),
							RafflePremiumAmount: insurer.AmountDetails{
								Amount:         "38062",
								UnitType:       insurer.UnitTypePercentage,
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.UnitDescriptionBRL,
								},
							},
							ContributionAmount: insurer.AmountDetails{
								Amount:         "100.00",
								UnitType:       insurer.UnitTypePercentage,
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.UnitDescriptionBRL,
								},
							},
							Subscribers: []capitalizationtitle.Subscriber{
								{
									Name:               "Nome do Subscritor",
									DocumentType:       capitalizationtitle.DocumentTypeOthers,
									DocumentTypeOthers: pointerOf("string"),
									DocumentNumber:     "string",
									Phones: &[]capitalizationtitle.Phone{
										{
											CountryCallingCode: pointerOf("55"),
											AreaCode:           pointerOf(insurer.PhoneAreaCode11),
											Number:             pointerOf("29875132"),
										},
									},
									Address:               "Av Naburo Ykesaki, 1270",
									AddressAdditionalInfo: pointerOf("Fundos"),
									TownName:              "Rio de Janeiro",
									CountrySubDivision:    insurer.CountrySubDivisionRJ,
									CountryCode:           string(insurer.CountryCodeBrazil),
									PostCode:              "17500001",
									Holders: &[]capitalizationtitle.Holder{
										{
											Name:               "Nome do Titular",
											DocumentType:       capitalizationtitle.DocumentTypeOthers,
											DocumentTypeOthers: pointerOf("string"),
											DocumentNumber:     "string",
											Phones: &[]capitalizationtitle.Phone{
												{
													CountryCallingCode: pointerOf("55"),
													AreaCode:           pointerOf(insurer.PhoneAreaCode11),
													Number:             pointerOf("29875132"),
												},
											},
											Address:               "Av Naburo Ykesaki, 1270",
											AddressAdditionalInfo: pointerOf("Fundos"),
											TownName:              "Rio de Janeiro",
											CountrySubDivision:    insurer.CountrySubDivisionRJ,
											CountryCode:           string(insurer.CountryCodeBrazil),
											PostCode:              "17500001",
											Redemption:            false,
											Raffle:                false,
										},
									},
								},
							},
							TechnicalProvisions: []capitalizationtitle.TechnicalProvision{
								{
									PMCAmount: insurer.AmountDetails{
										Amount:         "24038817.64",
										UnitType:       insurer.UnitTypePercentage,
										UnitTypeOthers: pointerOf("Horas"),
										Unit: &insurer.Unit{
											Code:        insurer.UnitCodeReal,
											Description: insurer.UnitDescriptionBRL,
										},
									},
									PDBAmount: insurer.AmountDetails{
										Amount:         "727561",
										UnitType:       insurer.UnitTypePercentage,
										UnitTypeOthers: pointerOf("Horas"),
										Unit: &insurer.Unit{
											Code:        insurer.UnitCodeReal,
											Description: insurer.UnitDescriptionBRL,
										},
									},
									PRAmount: insurer.AmountDetails{
										Amount:         "55",
										UnitType:       insurer.UnitTypePercentage,
										UnitTypeOthers: pointerOf("Horas"),
										Unit: &insurer.Unit{
											Code:        insurer.UnitCodeReal,
											Description: insurer.UnitDescriptionBRL,
										},
									},
									PSPAmount: insurer.AmountDetails{
										Amount:         "033610",
										UnitType:       insurer.UnitTypePercentage,
										UnitTypeOthers: pointerOf("Horas"),
										Unit: &insurer.Unit{
											Code:        insurer.UnitCodeReal,
											Description: insurer.UnitDescriptionBRL,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testCapitalizationTitlePlan).Error; err != nil {
		return fmt.Errorf("failed to create test capitalization title plan: %w", err)
	}

	testCapitalizationTitleEvent := &capitalizationtitle.Event{
		ID:     uuid.MustParse("8983016f-2a82-4d70-8bcc-71dd4e28cd60"),
		PlanID: testCapitalizationTitlePlan.ID,
		Data: capitalizationtitle.EventData{
			TitleID: pointerOf("string"),
			Type:    pointerOf(capitalizationtitle.EventTypeRaffle),
			Raffle: &capitalizationtitle.Raffle{
				Amount: insurer.AmountDetails{
					Amount:         "100.00",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.UnitDescriptionBRL,
					},
				},
				Date:           mustParseBrazilDate("2023-01-30"),
				SettlementDate: mustParseBrazilDate("2023-01-30"),
			},
			Redemption: &capitalizationtitle.Redemption{
				Type: capitalizationtitle.RedemptionTypePartialAnticipation,
				Amount: insurer.AmountDetails{
					Amount:         "94.69",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.UnitDescriptionBRL,
					},
				},
				BonusAmount: insurer.AmountDetails{
					Amount:         "100.00",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.UnitDescriptionBRL,
					},
				},
				UnreturnedAmount: pointerOf(insurer.AmountDetails{
					Amount:         "100.00",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.UnitDescriptionBRL,
					},
				}),
				Date:           mustParseBrazilDate("2023-01-30"),
				SettlementDate: mustParseBrazilDate("2023-01-30"),
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testCapitalizationTitleEvent).Error; err != nil {
		return fmt.Errorf("failed to create test capitalization title event: %w", err)
	}

	testCapitalizationTitleSettlement := &capitalizationtitle.Settlement{
		ID:     uuid.MustParse("7c74f797-a7aa-433c-bdb7-230eb0421f1e"),
		PlanID: testCapitalizationTitlePlan.ID,
		Data: capitalizationtitle.SettlementData{
			FinancialAmount: insurer.AmountDetails{
				Amount:         "521601042873331.15",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.UnitDescriptionBRL,
				},
			},
			PaymentDate: mustParseBrazilDate("2023-01-30"),
			DueDate:     mustParseBrazilDate("2023-01-30"),
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testCapitalizationTitleSettlement).Error; err != nil {
		return fmt.Errorf("failed to create test capitalization title settlement: %w", err)
	}

	return nil
}

func mustParseDateTime(s string) timeutil.DateTime {
	t, _ := time.Parse("2006-01-02T15:04:05Z", s)
	return timeutil.NewDateTime(t)
}
