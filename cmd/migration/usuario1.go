package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/acceptancebranchesabroad"
	"github.com/luikyv/mock-insurer/internal/auto"
	"github.com/luikyv/mock-insurer/internal/capitalizationtitle"
	"github.com/luikyv/mock-insurer/internal/customer"
	"github.com/luikyv/mock-insurer/internal/financialassistance"
	"github.com/luikyv/mock-insurer/internal/financialrisk"
	"github.com/luikyv/mock-insurer/internal/housing"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/lifepension"
	"github.com/luikyv/mock-insurer/internal/patrimonial"
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
					Email:                    pointerOf("test@test.com"),
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
							PremiumPeriodicity:            insurer.PremiumPeriodicityMonthly,
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
						MovementType:           insurer.PaymentMovementTypePremiumLiquidation,
						MovementOrigin:         pointerOf(insurer.PaymentMovementOriginDirectIssuance),
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
						PaymentType:              pointerOf(insurer.PaymentTypeBankSlip),
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
									Description: insurer.CurrencyBRL,
								},
							},
							ContributionAmount: insurer.AmountDetails{
								Amount:         "100.00",
								UnitType:       insurer.UnitTypePercentage,
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.CurrencyBRL,
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
											Description: insurer.CurrencyBRL,
										},
									},
									PDBAmount: insurer.AmountDetails{
										Amount:         "727561",
										UnitType:       insurer.UnitTypePercentage,
										UnitTypeOthers: pointerOf("Horas"),
										Unit: &insurer.Unit{
											Code:        insurer.UnitCodeReal,
											Description: insurer.CurrencyBRL,
										},
									},
									PRAmount: insurer.AmountDetails{
										Amount:         "55",
										UnitType:       insurer.UnitTypePercentage,
										UnitTypeOthers: pointerOf("Horas"),
										Unit: &insurer.Unit{
											Code:        insurer.UnitCodeReal,
											Description: insurer.CurrencyBRL,
										},
									},
									PSPAmount: insurer.AmountDetails{
										Amount:         "033610",
										UnitType:       insurer.UnitTypePercentage,
										UnitTypeOthers: pointerOf("Horas"),
										Unit: &insurer.Unit{
											Code:        insurer.UnitCodeReal,
											Description: insurer.CurrencyBRL,
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
						Description: insurer.CurrencyBRL,
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
						Description: insurer.CurrencyBRL,
					},
				},
				BonusAmount: insurer.AmountDetails{
					Amount:         "100.00",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
					},
				},
				UnreturnedAmount: pointerOf(insurer.AmountDetails{
					Amount:         "100.00",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
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
					Description: insurer.CurrencyBRL,
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

	testFinancialAssistanceContract := &financialassistance.Contract{
		ID:      "financial-assistance-contract-001",
		OwnerID: testUser.ID,
		Data: financialassistance.ContractData{
			CertificateID:      "42",
			GroupContractID:    pointerOf("42"),
			SusepProcessNumber: "12345",
			Insureds: []financialassistance.Insured{
				{
					DocumentType:       financialassistance.DocumentTypeCPF,
					DocumentTypeOthers: pointerOf("string"),
					DocumentNumber:     "12345678910",
					Name:               "Juan Kaique Cláudio Fernandes",
				},
			},
			ConceivedCreditValue: insurer.AmountDetails{
				Amount:         "654599.53",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			CreditedLiquidValue: insurer.AmountDetails{
				Amount:         "70732",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			CounterInstallments: financialassistance.CounterInstallment{
				Value: insurer.AmountDetails{
					Amount:         "250044238268.53",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
					},
				},
				Periodicity: financialassistance.CounterInstallmentPeriodicityMonthly,
				Quantity:    4,
				FirstDate:   mustParseBrazilDate("2021-05-21"),
				LastDate:    mustParseBrazilDate("2021-09-21"),
			},
			InterestRate: insurer.AmountDetails{
				Amount:         "85820115752.46",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			EffectiveCostRate: insurer.AmountDetails{
				Amount:         "69377099807020.86",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			AmortizationPeriod: 4,
			AcquittanceValue: pointerOf(insurer.AmountDetails{
				Amount:         "4991183909.95",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			}),
			AcquittanceDate: pointerOf(mustParseBrazilDate("2021-09-21")),
			TaxesValue: insurer.AmountDetails{
				Amount:         "3.69",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			ExpensesValue: pointerOf(insurer.AmountDetails{
				Amount:         "171719391175072.92",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			}),
			FinesValue: pointerOf(insurer.AmountDetails{
				Amount:         "100.00",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			}),
			MonetaryUpdatesValue: pointerOf(insurer.AmountDetails{
				Amount:         "4.61",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			}),
			AdministrativeFeesValue: insurer.AmountDetails{
				Amount:         "100.00",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			InterestValue: insurer.AmountDetails{
				Amount:         "808",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testFinancialAssistanceContract).Error; err != nil {
		return fmt.Errorf("failed to create test financial assistance contract: %w", err)
	}

	testFinancialAssistanceMovement := &financialassistance.Movement{
		ID:         uuid.MustParse("7c74f797-a7aa-433c-bdb7-230eb0421f1e"),
		ContractID: testFinancialAssistanceContract.ID,
		MovementData: financialassistance.MovementData{
			UpdatedDebitAmount: insurer.AmountDetails{
				Amount:         "100.00",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			RemainingCounterInstallmentsQuantity:       4,
			RemainingUnpaidCounterInstallmentsQuantity: 4,
			LifePensionPMBACAmount: pointerOf(insurer.AmountDetails{
				Amount:         "692256726733238.17",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			}),
			PensionPlanPMBACAmount: pointerOf(insurer.AmountDetails{
				Amount:         "21744585058654.54",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			}),
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testFinancialAssistanceMovement).Error; err != nil {
		return fmt.Errorf("failed to create test financial assistance movement: %w", err)
	}

	testAcceptanceAndBranchesAbroadPolicy := &acceptancebranchesabroad.Policy{
		ID:      "acceptance-branches-abroad-policy-001",
		OwnerID: testUser.ID,
		Data: acceptancebranchesabroad.PolicyData{
			ProductName:         "Acceptance and Branches Abroad Policy",
			DocumentType:        acceptancebranchesabroad.DocumentTypeIndividualPolicy,
			SusepProcessNumber:  pointerOf("string"),
			GroupCertificateID:  pointerOf("string"),
			IssuanceType:        acceptancebranchesabroad.IssuanceTypeOwn,
			IssuanceDate:        mustParseBrazilDate("2022-05-21"),
			TermStartDate:       mustParseBrazilDate("2022-05-21"),
			TermEndDate:         mustParseBrazilDate("2022-05-21"),
			LeadInsurerCode:     pointerOf("string"),
			LeadInsurerPolicyID: pointerOf("string"),
			MaxLMG: insurer.AmountDetails{
				Amount:         "78389",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			ProposalID: "string",
			Insureds: []acceptancebranchesabroad.Insured{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 "10000000",
					BirthDate:                mustParseBrazilDate("1999-06-12"),
					Email:                    pointerOf("email@gmail.com"),
					City:                     "string",
					State:                    "SP",
					Country:                  insurer.CountryCodeBrazil,
					Address:                  "string",
					AddressAdditionalInfo:    pointerOf("string"),
				},
			},
			Beneficiaries: pointerOf([]acceptancebranchesabroad.Beneficiary{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
				},
			}),
			Principals: pointerOf([]acceptancebranchesabroad.Principal{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 "10000000",
					Email:                    pointerOf("email@gmail.com"),
					City:                     "string",
					State:                    "SP",
					Country:                  insurer.CountryCodeBrazil,
					Address:                  "string",
					AddressAdditionalInfo:    pointerOf("string"),
				},
			}),
			Intermediaries: pointerOf([]acceptancebranchesabroad.Intermediary{
				{
					Type:                     acceptancebranchesabroad.IntermediaryTypeRepresentative,
					TypeOthers:               pointerOf("string"),
					Identification:           pointerOf("12345678900"),
					BrokerID:                 pointerOf("073158995"),
					IdentificationType:       pointerOf(insurer.IdentificationTypeCPF),
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 pointerOf("10000000"),
					City:                     pointerOf("string"),
					State:                    pointerOf("SP"),
					Country:                  pointerOf("BRA"),
					Address:                  pointerOf("string"),
				},
			}),
			InsuredObjects: []acceptancebranchesabroad.InsuredObject{
				{
					Identification:     "string",
					Type:               acceptancebranchesabroad.InsuredObjectTypeContract,
					TypeAdditionalInfo: pointerOf("string"),
					Description:        "string",
					Amount: pointerOf(insurer.AmountDetails{
						Amount:         "100.00",
						UnitType:       insurer.UnitTypePercentage,
						UnitTypeOthers: pointerOf("Horas"),
						Unit: &insurer.Unit{
							Code:        insurer.UnitCodeReal,
							Description: insurer.CurrencyBRL,
						},
					}),
					Coverages: []acceptancebranchesabroad.InsuredObjectCoverage{
						{
							Branch:             "0111",
							Code:               "OUTRAS",
							Description:        pointerOf("string"),
							InternalCode:       pointerOf("string"),
							SusepProcessNumber: "string",
							LMI: insurer.AmountDetails{
								Amount:         "100.00",
								UnitType:       insurer.UnitTypePercentage,
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.CurrencyBRL,
								},
							},
							TermStartDate:             mustParseBrazilDate("2022-05-21"),
							TermEndDate:               mustParseBrazilDate("2022-05-21"),
							IsMainCoverage:            pointerOf(true),
							Feature:                   acceptancebranchesabroad.CoverageFeatureMass,
							Type:                      acceptancebranchesabroad.CoverageTypeParametric,
							GracePeriod:               pointerOf(0),
							GracePeriodicity:          pointerOf(insurer.PeriodicityDay),
							GracePeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
							GracePeriodStartDate:      pointerOf(mustParseBrazilDate("2022-05-21")),
							GracePeriodEndDate:        pointerOf(mustParseBrazilDate("2022-05-21")),
							PremiumPeriodicity:        insurer.PremiumPeriodicityMonthly,
							PremiumPeriodicityOthers:  pointerOf("string"),
						},
					},
				},
			},
			Coverages: pointerOf([]acceptancebranchesabroad.Coverage{
				{
					Branch:      "0111",
					Code:        "OUTRAS",
					Description: pointerOf("string"),
					Deductible: pointerOf(acceptancebranchesabroad.CoverageDeductible{
						Type:               acceptancebranchesabroad.CoverageDeductibleTypeDeductible,
						TypeAdditionalInfo: pointerOf("string"),
						Amount: insurer.AmountDetails{
							Amount:         "100.00",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
						Period:               10,
						Periodicity:          insurer.PeriodicityDay,
						PeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
						PeriodStartDate:      mustParseBrazilDate("2022-05-16"),
						PeriodEndDate:        mustParseBrazilDate("2022-05-17"),
						Description:          "Franquia de exemplo",
					}),
					POS: pointerOf(acceptancebranchesabroad.CoveragePOS{
						ApplicationType: insurer.ValueTypeValue,
						Description:     "string",
						MinValue: pointerOf(insurer.AmountDetails{
							Amount:         "5.68",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						MaxValue: pointerOf(insurer.AmountDetails{
							Amount:         "102886.81",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						Percentage: pointerOf(insurer.AmountDetails{
							Amount:         "2",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						ValueOthers: pointerOf(insurer.AmountDetails{
							Amount:         "3171",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
					}),
				},
			}),
			CoinsuranceRetainedPercentage: pointerOf("10.00"),
			Coinsurers: pointerOf([]acceptancebranchesabroad.Coinsurer{
				{
					Identification:  pointerOf("string"),
					CededPercentage: pointerOf("10.00"),
				},
			}),
			BranchInfo: acceptancebranchesabroad.BranchInfo{
				RiskCountry:      insurer.CountryCodeBrazil,
				HasForum:         true,
				ForumDescription: pointerOf("string"),
				TransferorID:     "12345678912",
				TransferorName:   "Nome Sobrenome",
				GroupBranches:    []string{"0111"},
			},
			Premium: acceptancebranchesabroad.Premium{
				PaymentsQuantity: "4",
				Amount: insurer.AmountDetails{
					Amount:         "67415467082.34",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
					},
				},
				Coverages: []acceptancebranchesabroad.PremiumCoverage{
					{
						Branch:      "0111",
						Code:        "OUTRAS",
						Description: pointerOf("string"),
						PremiumAmount: insurer.AmountDetails{
							Amount:         "100.00",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
					},
				},
				Payments: []acceptancebranchesabroad.Payment{
					{
						MovementDate:           mustParseBrazilDate("2022-05-21"),
						MovementType:           insurer.PaymentMovementTypePremiumLiquidation,
						MovementOrigin:         pointerOf(insurer.PaymentMovementOriginDirectIssuance),
						MovementPaymentsNumber: "1",
						Amount: insurer.AmountDetails{
							Amount:         "1.39",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
						MaturityDate:             mustParseBrazilDate("2022-05-21"),
						TellerID:                 pointerOf("string"),
						TellerIDType:             pointerOf(insurer.IdentificationTypeCPF),
						TellerIDOthers:           pointerOf("RNE"),
						TellerName:               pointerOf("string"),
						FinancialInstitutionCode: pointerOf("string"),
						PaymentType:              pointerOf(insurer.PaymentTypeBankSlip),
						PaymentTypeOthers:        pointerOf("string"),
					},
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testAcceptanceAndBranchesAbroadPolicy).Error; err != nil {
		return fmt.Errorf("failed to create test acceptance and branches abroad policy: %w", err)
	}

	testAcceptanceAndBranchesAbroadClaim := &acceptancebranchesabroad.Claim{
		ID:       uuid.MustParse("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
		PolicyID: testAcceptanceAndBranchesAbroadPolicy.ID,
		Data: acceptancebranchesabroad.ClaimData{
			Identification:       "string",
			DocumentDeliveryDate: pointerOf(mustParseBrazilDate("2022-05-21")),
			Status:               acceptancebranchesabroad.ClaimStatusOpen,
			StatusAlterationDate: pointerOf(mustParseBrazilDate("2022-05-21")),
			OccurrenceDate:       mustParseBrazilDate("2022-05-21"),
			WarningDate:          mustParseBrazilDate("2022-05-21"),
			ThirdPartyClaimDate:  pointerOf(mustParseBrazilDate("2022-05-21")),
			Amount: insurer.AmountDetails{
				Amount:         "8798779606.95",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			DenialJustification:            pointerOf(acceptancebranchesabroad.DenialJustificationExcludedRisk),
			DenialJustificationDescription: pointerOf("string"),
			Coverages: []acceptancebranchesabroad.ClaimCoverage{
				{
					InsuredObjectID:     pointerOf("string"),
					Branch:              "0111",
					Code:                "OUTRAS",
					Description:         pointerOf("string"),
					WarningDate:         pointerOf(mustParseBrazilDate("2022-12-31")),
					ThirdPartyClaimDate: pointerOf(mustParseBrazilDate("2022-12-31")),
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testAcceptanceAndBranchesAbroadClaim).Error; err != nil {
		return fmt.Errorf("failed to create test acceptance and branches abroad claim: %w", err)
	}

	testFinancialRiskPolicy := &financialrisk.Policy{
		ID:      "financial-risk-policy-001",
		OwnerID: testUser.ID,
		Data: financialrisk.PolicyData{
			ProductName:         "Financial Risk Policy",
			DocumentType:        financialrisk.DocumentTypeIndividualPolicy,
			SusepProcessNumber:  pointerOf("string"),
			GroupCertificateID:  pointerOf("string"),
			IssuanceType:        financialrisk.IssuanceTypeOwn,
			IssuanceDate:        mustParseBrazilDate("2022-12-31"),
			TermStartDate:       mustParseBrazilDate("2022-12-31"),
			TermEndDate:         mustParseBrazilDate("2022-12-31"),
			LeadInsurerCode:     pointerOf("string"),
			LeadInsurerPolicyID: pointerOf("string"),
			MaxLMG: pointerOf(insurer.AmountDetails{
				Amount:         "01",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			}),
			ProposalID: "string",
			Insureds: []financialrisk.Insured{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 "10000000",
					BirthDate:                mustParseBrazilDate("1999-06-12"),
					Email:                    pointerOf("]!|8y5j-RE>]s/Wr0$vd{Z@>p8j<6|M%v%'(n@pE~q(/LT.dHnk>1X*&(.=}M`&VTQ:domV+'k~y+KGfK:=%#nl-|8?[z.yWua)pgLvI.iz{YZjyA=CLA>nPv#V'~IqL<<CU8b>Po"),
					City:                     "string",
					State:                    "AC",
					Country:                  "BRA",
					Address:                  "string",
				},
			},
			Beneficiaries: pointerOf([]financialrisk.Beneficiary{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
				},
			}),
			Principals: pointerOf([]financialrisk.Principal{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 "10000000",
					Email:                    pointerOf("]%g-L@*M'=s?6=/6VD\"s^7~1_[\\7Pv:bH\"6X2FY@t(kn+.W^SH<%4$I["),
					City:                     "string",
					State:                    "AC",
					Country:                  "BRA",
					Address:                  "string",
					AddressAdditionalInfo:    pointerOf("Fundos"),
				},
			}),
			Intermediaries: pointerOf([]financialrisk.Intermediary{
				{
					Type:                     financialrisk.IntermediaryTypeRepresentative,
					TypeOthers:               pointerOf("string"),
					Identification:           pointerOf("12345678900"),
					BrokerID:                 pointerOf("395257613"),
					IdentificationType:       pointerOf(insurer.IdentificationTypeCPF),
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 pointerOf("10000000"),
					City:                     pointerOf("string"),
					State:                    pointerOf("string"),
					Country:                  pointerOf("BRA"),
					Address:                  pointerOf("string"),
				},
			}),
			InsuredObjects: []financialrisk.InsuredObject{
				{
					Identification:     pointerOf("string"),
					Type:               financialrisk.InsuredObjectTypeContract,
					TypeAdditionalInfo: pointerOf("string"),
					Description:        "string",
					Amount: pointerOf(insurer.AmountDetails{
						Amount:         "07.33",
						UnitType:       insurer.UnitTypePercentage,
						UnitTypeOthers: pointerOf("Horas"),
						Unit: &insurer.Unit{
							Code:        insurer.UnitCodeReal,
							Description: insurer.CurrencyBRL,
						},
					}),
					Coverages: []financialrisk.InsuredObjectCoverage{
						{
							Branch:             "0111",
							Code:               "PROTECAO_DE_BENS",
							Description:        pointerOf("string"),
							InternalCode:       pointerOf("string"),
							SusepProcessNumber: "string",
							LMI: insurer.AmountDetails{
								Amount:         "8553221718627.79",
								UnitType:       insurer.UnitTypePercentage,
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.CurrencyBRL,
								},
							},
							TermStartDate:             mustParseBrazilDate("2022-12-31"),
							TermEndDate:               mustParseBrazilDate("2022-12-31"),
							IsMainCoverage:            pointerOf(true),
							Feature:                   financialrisk.CoverageFeatureMass,
							Type:                      financialrisk.CoverageTypeParametric,
							GracePeriod:               pointerOf(0),
							GracePeriodicity:          pointerOf(insurer.PeriodicityDay),
							GracePeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
							GracePeriodStartDate:      pointerOf(mustParseBrazilDate("2022-12-31")),
							GracePeriodEndDate:        pointerOf(mustParseBrazilDate("2022-12-31")),
							IsLMISublimit:             pointerOf(true),
							PremiumPeriodicity:        insurer.PremiumPeriodicityMonthly,
							PremiumPeriodicityOthers:  pointerOf("string"),
						},
					},
				},
			},
			Coverages: pointerOf([]financialrisk.Coverage{
				{
					Branch:      "0111",
					Code:        financialrisk.CoverageCodeProtectionOfAssets,
					Description: pointerOf("string"),
					Deductible: pointerOf(financialrisk.CoverageDeductible{
						Type:               financialrisk.CoverageDeductibleTypeDeductible,
						TypeAdditionalInfo: pointerOf("string"),
						Amount: insurer.AmountDetails{
							Amount:         "44670",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
						Period:               10,
						Periodicity:          insurer.PeriodicityDay,
						PeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
						PeriodStartDate:      mustParseBrazilDate("2022-05-16"),
						PeriodEndDate:        mustParseBrazilDate("2022-05-17"),
						Description:          "Franquia de exemplo",
					}),
					POS: pointerOf(financialrisk.CoveragePOS{
						ApplicationType: insurer.ValueTypeValue,
						Description:     "string",
						MinValue: pointerOf(insurer.AmountDetails{
							Amount:         "100.00",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						MaxValue: pointerOf(insurer.AmountDetails{
							Amount:         "5.84",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						Percentage: pointerOf(insurer.AmountDetails{
							Amount:         "80015.20",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						ValueOthers: pointerOf(insurer.AmountDetails{
							Amount:         "373482",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
					}),
				},
			}),
			CoinsuranceRetainedPercentage: pointerOf("10.00"),
			Coinsurers: pointerOf([]financialrisk.Coinsurer{
				{
					Identification:  pointerOf("string"),
					CededPercentage: "10.00",
				},
			}),
			BranchInfo: pointerOf(financialrisk.BranchInfo{
				Identification:   "string",
				UserGroup:        "string",
				TechnicalSurplus: "10.00",
			}),
			Premium: financialrisk.Premium{
				PaymentsQuantity: 4,
				Amount: insurer.AmountDetails{
					Amount:         "49.89",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
					},
				},
				Coverages: []financialrisk.PremiumCoverage{
					{
						Branch:      "0111",
						Code:        "PROTECAO_DE_BENS",
						Description: pointerOf("string"),
						PremiumAmount: insurer.AmountDetails{
							Amount:         "75017",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
					},
				},
				Payments: []financialrisk.Payment{
					{
						MovementDate:           mustParseBrazilDate("2022-12-31"),
						MovementType:           insurer.PaymentMovementTypePremiumLiquidation,
						MovementOrigin:         pointerOf(insurer.PaymentMovementOriginDirectIssuance),
						MovementPaymentsNumber: "1",
						Amount: insurer.AmountDetails{
							Amount:         "6.11",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
						MaturityDate:             mustParseBrazilDate("2022-12-31"),
						TellerID:                 pointerOf("string"),
						TellerIDType:             pointerOf(insurer.IdentificationTypeCPF),
						TellerIDOthers:           pointerOf("RNE"),
						TellerName:               pointerOf("string"),
						FinancialInstitutionCode: pointerOf("string"),
						PaymentType:              pointerOf(insurer.PaymentTypeBankSlip),
						PaymentTypeOthers:        pointerOf("string"),
					},
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testFinancialRiskPolicy).Error; err != nil {
		return fmt.Errorf("failed to create test financial risk policy: %w", err)
	}

	testFinancialRiskClaim := &financialrisk.Claim{
		ID:       uuid.MustParse("b2c3d4e5-f6a7-8901-bcde-f12345678901"),
		PolicyID: testFinancialRiskPolicy.ID,
		Data: financialrisk.ClaimData{
			Identification:            "string",
			DocumentationDeliveryDate: pointerOf(mustParseBrazilDate("2022-12-31")),
			Status:                    financialrisk.ClaimStatusOpen,
			StatusAlterationDate:      mustParseBrazilDate("2022-12-31"),
			OccurrenceDate:            mustParseBrazilDate("2022-12-31"),
			WarningDate:               mustParseBrazilDate("2022-12-31"),
			ThirdPartyClaimDate:       pointerOf(mustParseBrazilDate("2022-12-31")),
			Amount: insurer.AmountDetails{
				Amount:         "3618",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			DenialJustification:            pointerOf(financialrisk.DenialJustificationExcludedRisk),
			DenialJustificationDescription: pointerOf("string"),
			Coverages: []financialrisk.ClaimCoverage{
				{
					InsuredObjectID:     pointerOf("string"),
					Branch:              "0111",
					Code:                financialrisk.CoverageCodeProtectionOfAssets,
					Description:         pointerOf("string"),
					WarningDate:         pointerOf(mustParseBrazilDate("2022-12-31")),
					ThirdPartyClaimDate: pointerOf(mustParseBrazilDate("2022-12-31")),
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testFinancialRiskClaim).Error; err != nil {
		return fmt.Errorf("failed to create test financial risk claim: %w", err)
	}

	testHousingPolicy := &housing.Policy{
		ID:      "housing-policy-001",
		OwnerID: testUser.ID,
		Data: housing.PolicyData{
			ProductName:         "Housing Policy",
			DocumentType:        housing.DocumentTypeIndividualPolicy,
			SusepProcessNumber:  pointerOf("string"),
			GroupCertificateID:  pointerOf("string"),
			IssuanceType:        housing.IssuanceTypeOwn,
			IssuanceDate:        mustParseBrazilDate("2022-12-31"),
			TermStartDate:       mustParseBrazilDate("2022-12-31"),
			TermEndDate:         mustParseBrazilDate("2022-12-31"),
			LeadInsurerCode:     pointerOf("string"),
			LeadInsurerPolicyID: pointerOf("string"),
			MaxLMG: insurer.AmountDetails{
				Amount:         "34.71",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			ProposalID: "string",
			Insureds: []housing.Insured{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 "10000000",
					BirthDate:                pointerOf(mustParseBrazilDate("1999-06-12")),
					Email:                    pointerOf("&BS\\CBR2<v)gG{@IkCM~t623e80!6</dolSoO6;x35T-[nNxs&=-6r~2p>Xg0@hQdF<`bp24s'w,]}})Yjupi|hh)]'2klhnbz$WJ>A;N-R41z=5B#D?3Sf=z,*.A[@y?tHa9&p/PXliits&kFuj7Q+Q{^nw%ABZ?WTH$_CIal:7YV&1`bN%"),
					City:                     "string",
					State:                    "AC",
					Country:                  "BRA",
					Address:                  "string",
				},
			},
			Beneficiaries: pointerOf([]housing.Beneficiary{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
				},
			}),
			Principals: pointerOf([]housing.Principal{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 "10000000",
					Email:                    pointerOf("}cw@'/v{i/CN.iZ3VT4,[%g`bd!WQBQF@x<ZoML7?hozTP:N>~Qa^DT?u}))_KeNL@.t\"}S*:/x7uc.!vz2$+(gBb.{ZbJ"),
					City:                     "string",
					State:                    "AC",
					Country:                  "BRA",
					Address:                  "string",
					AddressAdditionalInfo:    pointerOf("Fundos"),
				},
			}),
			Intermediaries: pointerOf([]housing.Intermediary{
				{
					Type:                     housing.IntermediaryTypeRepresentative,
					TypeOthers:               pointerOf("string"),
					Identification:           pointerOf("12345678900"),
					BrokerID:                 pointerOf("210917424"),
					IdentificationType:       pointerOf(insurer.IdentificationTypeCPF),
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 pointerOf("10000000"),
					City:                     pointerOf("string"),
					State:                    pointerOf("string"),
					Country:                  pointerOf("BRA"),
					Address:                  pointerOf("string"),
				},
			}),
			InsuredObjects: []housing.InsuredObject{
				{
					Identification:     pointerOf("string"),
					Type:               housing.InsuredObjectTypeContract,
					TypeAdditionalInfo: pointerOf("string"),
					Description:        "string",
					Amount: pointerOf(insurer.AmountDetails{
						Amount:         "100.00",
						UnitType:       insurer.UnitTypePercentage,
						UnitTypeOthers: pointerOf("Horas"),
						Unit: &insurer.Unit{
							Code:        insurer.UnitCodeReal,
							Description: insurer.CurrencyBRL,
						},
					}),
					Coverages: []housing.InsuredObjectCoverage{
						{
							Branch:             "0111",
							Code:               "DANOS_ELETRICOS",
							Description:        pointerOf("string"),
							InternalCode:       pointerOf("string"),
							SusepProcessNumber: "string",
							LMI: insurer.AmountDetails{
								Amount:         "100.00",
								UnitType:       insurer.UnitTypePercentage,
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        insurer.UnitCodeReal,
									Description: insurer.CurrencyBRL,
								},
							},
							TermStartDate:             mustParseBrazilDate("2022-12-31"),
							TermEndDate:               mustParseBrazilDate("2022-12-31"),
							IsMainCoverage:            pointerOf(true),
							Feature:                   housing.CoverageFeatureMass,
							Type:                      housing.CoverageTypeParametric,
							GracePeriod:               pointerOf(0),
							GracePeriodicity:          pointerOf(insurer.PeriodicityDay),
							GracePeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
							GracePeriodStartDate:      pointerOf(mustParseBrazilDate("2022-12-31")),
							GracePeriodEndDate:        pointerOf(mustParseBrazilDate("2022-12-31")),
							PremiumPeriodicity:        insurer.PremiumPeriodicityMonthly,
							PremiumPeriodicityOthers:  pointerOf("string"),
						},
					},
				},
			},
			Coverages: pointerOf([]housing.Coverage{
				{
					Branch:      "0111",
					Code:        housing.CoverageCodeElectricalDamage,
					Description: pointerOf("string"),
					Deductible: pointerOf(housing.CoverageDeductible{
						Type:               housing.CoverageDeductibleTypeDeductible,
						TypeAdditionalInfo: pointerOf("string"),
						Amount: insurer.AmountDetails{
							Amount:         "1975251011677.32",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
						Period:               10,
						Periodicity:          insurer.PeriodicityDay,
						PeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
						PeriodStartDate:      mustParseBrazilDate("2022-05-16"),
						PeriodEndDate:        mustParseBrazilDate("2022-05-17"),
						Description:          "Franquia de exemplo",
					}),
					POS: pointerOf(housing.CoveragePOS{
						ApplicationType: insurer.ValueTypeValue,
						Description:     pointerOf("Descrição de exemplo"),
						MinValue: pointerOf(insurer.AmountDetails{
							Amount:         "9.65",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						MaxValue: pointerOf(insurer.AmountDetails{
							Amount:         "881095",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						Percentage: pointerOf(insurer.AmountDetails{
							Amount:         "1",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
						ValueOthers: pointerOf(insurer.AmountDetails{
							Amount:         "100.00",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						}),
					}),
				},
			}),
			CoinsuranceRetainedPercentage: pointerOf("10.00"),
			Coinsurers: pointerOf([]housing.Coinsurer{
				{
					Identification:  "string",
					CededPercentage: "10.00",
				},
			}),
			BranchInfo: housing.BranchInfo{
				InsuredObjects: []housing.SpecificInsuredObject{
					{
						Identification:             "string",
						PropertyType:               housing.PropertyTypeHouse,
						PropertyTypeAdditionalInfo: pointerOf("string"),
						PostCode:                   "10000000",
						InterestRate:               "10.00",
						CostRate:                   "10.00",
						UpdateIndex:                housing.UpdateIndex("IPCA_IBGE"),
						UpdateIndexOthers:          pointerOf("Índice de atualização"),
						Lenders: []housing.Lender{
							{
								CompanyName: "string",
								CnpjNumber:  "12345678901234",
							},
						},
					},
				},
				Insureds: []housing.SpecificInsured{
					{
						Identification:           "12345678900",
						IdentificationType:       insurer.IdentificationTypeCPF,
						IdentificationTypeOthers: pointerOf("RNE"),
						BirthDate:                pointerOf(mustParseBrazilDate("2022-12-31")),
					},
				},
			},
			Premium: housing.Premium{
				PaymentsQuantity: 4,
				Amount: insurer.AmountDetails{
					Amount:         "14465685570.04",
					UnitType:       insurer.UnitTypePercentage,
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
					},
				},
				Coverages: []housing.PremiumCoverage{
					{
						Branch:      "0111",
						Code:        "DANOS_ELETRICOS",
						Description: pointerOf("string"),
						PremiumAmount: insurer.AmountDetails{
							Amount:         "9.09",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
					},
				},
				Payments: []housing.Payment{
					{
						MovementDate:           mustParseBrazilDate("2022-12-31"),
						MovementType:           insurer.PaymentMovementTypePremiumLiquidation,
						MovementOrigin:         pointerOf(insurer.PaymentMovementOriginDirectIssuance),
						MovementPaymentsNumber: "str",
						Amount: insurer.AmountDetails{
							Amount:         "25.58",
							UnitType:       insurer.UnitTypePercentage,
							UnitTypeOthers: pointerOf("Horas"),
							Unit: &insurer.Unit{
								Code:        insurer.UnitCodeReal,
								Description: insurer.CurrencyBRL,
							},
						},
						MaturityDate:             mustParseBrazilDate("2022-12-31"),
						TellerID:                 pointerOf("string"),
						TellerIDType:             pointerOf(insurer.IdentificationTypeCPF),
						TellerIDOthers:           pointerOf("RNE"),
						TellerName:               pointerOf("string"),
						FinancialInstitutionCode: pointerOf("string"),
						PaymentType:              pointerOf(insurer.PaymentTypeBankSlip),
						PaymentTypeOthers:        pointerOf("string"),
					},
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testHousingPolicy).Error; err != nil {
		return fmt.Errorf("failed to create test housing policy: %w", err)
	}

	testHousingClaim := &housing.Claim{
		ID:       uuid.MustParse("c3d4e5f6-a7b8-9012-cdef-123456789012"),
		PolicyID: testHousingPolicy.ID,
		Data: housing.ClaimData{
			Identification:            "string",
			DocumentationDeliveryDate: pointerOf(mustParseBrazilDate("2022-12-31")),
			Status:                    housing.ClaimStatusOpen,
			StatusAlterationDate:      mustParseBrazilDate("2022-12-31"),
			OccurrenceDate:            mustParseBrazilDate("2022-12-31"),
			WarningDate:               mustParseBrazilDate("2022-12-31"),
			ThirdPartyClaimDate:       pointerOf(mustParseBrazilDate("2022-12-31")),
			Amount: insurer.AmountDetails{
				Amount:         "258392",
				UnitType:       insurer.UnitTypePercentage,
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        insurer.UnitCodeReal,
					Description: insurer.CurrencyBRL,
				},
			},
			DenialJustification:            pointerOf(housing.DenialJustificationExcludedRisk),
			DenialJustificationDescription: pointerOf("string"),
			Coverages: []housing.ClaimCoverage{
				{
					InsuredObjectID:     pointerOf("string"),
					Branch:              "0111",
					Code:                housing.CoverageCodeElectricalDamage,
					Description:         pointerOf("string"),
					WarningDate:         pointerOf(mustParseBrazilDate("2022-12-31")),
					ThirdPartyClaimDate: pointerOf(mustParseBrazilDate("2022-12-31")),
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testHousingClaim).Error; err != nil {
		return fmt.Errorf("failed to create test housing claim: %w", err)
	}

	testLifePensionContract := &lifepension.Contract{
		ID:      "life-pension-contract-001",
		OwnerID: testUser.ID,
		Data: lifepension.ContractData{
			ProductCode:        "1234",
			ProductName:        "Life Pension Product",
			ProposalID:         "987",
			ContractID:         pointerOf("681"),
			ContractingType:    lifepension.ContractingTypeIndividual,
			EffectiveDateStart: mustParseBrazilDate("2021-05-21"),
			EffectiveDateEnd:   mustParseBrazilDate("2021-05-21"),
			CertificateActive:  true,
			ConjugatedPlan:     true,
			PlanType:           pointerOf(lifepension.PlanTypeAverbado),
			Periodicity:        lifepension.PeriodicityMensal,
			PeriodicityOthers:  pointerOf("string"),
			TaxRegime:          lifepension.TaxRegimeProgressivo,
			Insured: lifepension.Insured{
				DocumentType:          lifepension.DocumentTypeCPF,
				DocumentTypeOthers:    pointerOf("OUTROS"),
				DocumentNumber:        "12345678910",
				Name:                  "Juan Kaique Cláudio Fernandes",
				BirthDate:             mustParseBrazilDate("2021-05-21"),
				Gender:                lifepension.GenderMasculino,
				PostCode:              "17500001",
				TownName:              "Rio de Janeiro",
				CountrySubDivision:    "RJ",
				CountryCode:           "BRA",
				Address:               "Av Naburo Ykesaki, 1270",
				AddressAdditionalInfo: pointerOf("Fundos"),
				Email:                 pointerOf("nome@br.net"),
			},
			Beneficiaries: &[]lifepension.Beneficiary{
				{
					DocumentNumber:          "12345678910",
					DocumentType:            lifepension.DocumentTypeCPF,
					DocumentTypeOthers:      pointerOf("OUTROS"),
					Name:                    "Juan Kaique Cláudio Fernandes",
					BirthDate:               pointerOf(mustParseBrazilDate("2022-12-31")),
					Kinship:                 pointerOf("PAIS"),
					KinshipOthers:           pointerOf("string"),
					ParticipationPercentage: "10.00",
				},
			},
			Intermediary: &lifepension.Intermediary{
				Type:               lifepension.IntermediaryTypeCorretor,
				TypeOthers:         pointerOf("string"),
				DocumentNumber:     pointerOf("12345678910"),
				IntermediaryID:     pointerOf("12097"),
				DocumentType:       pointerOf(lifepension.DocumentTypeCPF),
				DocumentTypeOthers: pointerOf("OUTROS"),
				Name:               pointerOf("Empresa A"),
				PostCode:           pointerOf("17500001"),
				TownName:           pointerOf("Rio de Janeiro"),
				CountrySubDivision: pointerOf("RJ"),
				CountryCode:        pointerOf("BRA"),
				Address:            pointerOf("Av Naburo Ykesaki, 1270"),
				AdditionalInfo:     pointerOf("Fundos"),
			},
			Suseps: []lifepension.Suseps{
				{
					CoverageCode:                      "1999",
					SusepProcessNumber:                "12345",
					StructureModality:                 lifepension.StructureModalityBeneficioDefinido,
					Type:                              lifepension.SusepsTypePGBL,
					TypeDetails:                       pointerOf("Descrição do Tipo de Plano"),
					LockedPlan:                        false,
					QualifiedProposer:                 false,
					BenefitPaymentMethod:              lifepension.BenefitPaymentMethodRenda,
					FinancialResultReversal:           false,
					FinancialResultReversalPercentage: pointerOf("10.00"),
					CalculationBasis:                  lifepension.CalculationBasisMensal,
					BenefitAmount: &insurer.AmountDetails{
						Amount:         "100.00",
						UnitType:       "PORCENTAGEM",
						UnitTypeOthers: pointerOf("Horas"),
						Unit: &insurer.Unit{
							Code:        "R$",
							Description: "BRL",
						},
					},
					RentsInterestRate: pointerOf("10.00"),
					Grace: &[]lifepension.Grace{
						{
							GraceType:              pointerOf(lifepension.GraceTypeResgate),
							GracePeriod:            pointerOf(4),
							GracePeriodicity:       pointerOf(lifepension.GracePeriodicityAno),
							DayIndicator:           pointerOf(lifepension.DayIndicatorUteis),
							GracePeriodStart:       pointerOf(mustParseBrazilDate("2021-05-21")),
							GracePeriodEnd:         pointerOf(mustParseBrazilDate("2021-05-21")),
							GracePeriodBetween:     pointerOf(6),
							GracePeriodBetweenType: pointerOf(lifepension.GracePeriodBetweenTypeDia),
						},
					},
					BiometricTable:                     pointerOf("AT50_M"),
					PmbacInterestRate:                  pointerOf("10.00"),
					PmbacGuaranteePriceIndex:           pointerOf("IPC-FGV"),
					PmbacGuaranteePriceOthers:          pointerOf("string"),
					PmbacIndexLagging:                  pointerOf(1),
					PdrOrVdrminimalGuaranteeIndex:      pointerOf("IPC-FGV"),
					PdrOrVdrminimalGuaranteeOthers:     pointerOf("string"),
					PdrOrVdrminimalGuaranteePercentage: pointerOf("10.00"),
					FIE: []lifepension.FIE{
						{
							FIECNPJ:      "12345678901234",
							FIEName:      "RAZÃO SOCIAL",
							FIETradeName: "NOME FANTASIA",
							PmbacAmount: insurer.AmountDetails{
								Amount:         "32.29",
								UnitType:       "PORCENTAGEM",
								UnitTypeOthers: pointerOf("Horas"),
								Unit: &insurer.Unit{
									Code:        "R$",
									Description: "BRL",
								},
							},
							ProvisionSurplusAmount: insurer.AmountDetails{
								Amount:         "100.00",
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
			},
			MovementContributions: []lifepension.MovementContribution{
				{
					Amount: insurer.AmountDetails{
						Amount:         "27059",
						UnitType:       "PORCENTAGEM",
						UnitTypeOthers: pointerOf("Horas"),
						Unit: &insurer.Unit{
							Code:        "R$",
							Description: "BRL",
						},
					},
					PaymentDate:    mustParseBrazilDate("2021-05-21"),
					ExpirationDate: mustParseBrazilDate("2021-05-21"),
					ChargedInAdvanceAmount: insurer.AmountDetails{
						Amount:         "6.10",
						UnitType:       "PORCENTAGEM",
						UnitTypeOthers: pointerOf("Horas"),
						Unit: &insurer.Unit{
							Code:        "R$",
							Description: "BRL",
						},
					},
					Periodicity:       lifepension.MovementPeriodicityMensal,
					PeriodicityOthers: pointerOf("string"),
				},
			},
			MovementBenefits: []lifepension.MovementBenefit{
				{
					Amount: insurer.AmountDetails{
						Amount:         "294852.44",
						UnitType:       "PORCENTAGEM",
						UnitTypeOthers: pointerOf("Horas"),
						Unit: &insurer.Unit{
							Code:        "R$",
							Description: "BRL",
						},
					},
					PaymentDate: mustParseBrazilDate("2021-05-21"),
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testLifePensionContract).Error; err != nil {
		return fmt.Errorf("failed to create test life pension contract: %w", err)
	}

	testLifePensionPortability := &lifepension.Portability{
		ID:         uuid.MustParse("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
		ContractID: testLifePensionContract.ID,
		Data: lifepension.PortabilityData{
			PortabilityDate:        mustParseBrazilDate("2022-05-20"),
			SourceInstitution:      "string",
			DestinationInstitution: "string",
			PortabilityAmount: insurer.AmountDetails{
				Amount:         "99.50",
				UnitType:       "PORCENTAGEM",
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        "R$",
					Description: "BRL",
				},
			},
			Status:     lifepension.PortabilityStatusCompleted,
			StatusDate: mustParseBrazilDate("2022-05-20"),
			Direction:  lifepension.PortabilityDirectionEntrada,
			PostedChargedAmount: insurer.AmountDetails{
				Amount:         "1.65",
				UnitType:       "PORCENTAGEM",
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        "R$",
					Description: "BRL",
				},
			},
			SusepProcess: pointerOf("12345"),
			TaxRegime:    pointerOf(lifepension.PortabilityTaxRegimeProgressivo),
			Type:         pointerOf(lifepension.PortabilityTypeParcial),
			FIE: &[]lifepension.PortabilityFIE{
				{
					FIECNPJ:      "12345678901234",
					FIEName:      "RAZÃO SOCIAL",
					FIETradeName: "NOME FANTASIA",
					PortedType:   lifepension.PortabilityFIEPortedTypeOrigin,
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testLifePensionPortability).Error; err != nil {
		return fmt.Errorf("failed to create test life pension portability: %w", err)
	}

	testLifePensionWithdrawal := &lifepension.Withdrawal{
		ID:         uuid.MustParse("b2c3d4e5-f6a7-8901-bcde-f12345678901"),
		ContractID: testLifePensionContract.ID,
		Data: lifepension.WithdrawalData{
			WithdrawalOccurence: true,
			Type:                pointerOf(lifepension.WithdrawalTypePartial),
			RequestDate:         pointerOf(mustParseDateTime("2022-05-20T08:30:00Z")),
			Amount: &insurer.AmountDetails{
				Amount:         "3",
				UnitType:       "PORCENTAGEM",
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        "R$",
					Description: "BRL",
				},
			},
			LiquidationDate: pointerOf(mustParseDateTime("2022-05-20T08:30:00Z")),
			PostedChargedAmount: &insurer.AmountDetails{
				Amount:         "100.00",
				UnitType:       "PORCENTAGEM",
				UnitTypeOthers: pointerOf("Horas"),
				Unit: &insurer.Unit{
					Code:        "R$",
					Description: "BRL",
				},
			},
			Nature: pointerOf(lifepension.WithdrawalNatureRegularWithdrawal),
			FIE: &[]lifepension.WithdrawalFIE{
				{
					FIECNPJ:      pointerOf("12345678901234"),
					FIEName:      pointerOf("RAZÃO SOCIAL"),
					FIETradeName: pointerOf("NOME FANTASIA"),
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testLifePensionWithdrawal).Error; err != nil {
		return fmt.Errorf("failed to create test life pension withdrawal: %w", err)
	}

	testLifePensionClaim := &lifepension.Claim{
		ID:         uuid.MustParse("c3d4e5f6-a7b8-9012-cdef-123456789012"),
		ContractID: testLifePensionContract.ID,
		Data: lifepension.ClaimData{
			EventInfo: lifepension.EventInfo{
				EventAlertDate:    mustParseBrazilDate("2021-05-21"),
				EventRegisterDate: mustParseBrazilDate("2021-05-21"),
				EventStatus:       lifepension.EventStatusOpen,
			},
			IncomeInfo: &lifepension.IncomeInfo{
				BeneficiaryDocument:      "12345678910",
				BeneficiaryDocumentType:  lifepension.DocumentTypeCPF,
				BeneficiaryDocTypeOthers: pointerOf("string"),
				BeneficiaryName:          "NOME BENEFICIARIO",
				BeneficiaryCategory:      lifepension.BeneficiaryCategoryInsured,
				BeneficiaryBirthDate:     mustParseBrazilDate("2021-05-21"),
				IncomeType:               lifepension.IncomeTypeSinglePayment,
				IncomeTypeDetails:        pointerOf("Descrição do Tipo de Renda"),
				ReversedIncome:           pointerOf(false),
				IncomeAmount: insurer.AmountDetails{
					Amount:         "100.00",
					UnitType:       "PORCENTAGEM",
					UnitTypeOthers: pointerOf("Horas"),
					Unit: &insurer.Unit{
						Code:        "R$",
						Description: "BRL",
					},
				},
				PaymentTerms:           pointerOf("PRAZO"),
				BenefitAmount:          2,
				GrantedDate:            mustParseBrazilDate("2021-05-21"),
				MonetaryUpdateIndex:    lifepension.MonetaryUpdateIndexOutros, // "IPC-FGV" mapped to OUTROS
				MonetaryUpdIndexOthers: pointerOf("string"),
				LastUpdateDate:         mustParseBrazilDate("2021-05-21"),
				DefermentDueDate:       pointerOf(mustParseBrazilDate("2025-12-31")),
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testLifePensionClaim).Error; err != nil {
		return fmt.Errorf("failed to create test life pension claim: %w", err)
	}

	testPatrimonialPolicy := &patrimonial.Policy{
		ID:      "patrimonial-policy-001",
		OwnerID: testUser.ID,
		Data: patrimonial.PolicyData{
			ProductName:         "Patrimonial Policy",
			DocumentType:        patrimonial.DocumentTypeIndividualPolicy,
			SusepProcessNumber:  "string",
			GroupCertificateID:  "string",
			IssuanceType:        patrimonial.IssuanceTypeOwn,
			IssuanceDate:        mustParseBrazilDate("2022-12-31"),
			TermStartDate:       mustParseBrazilDate("2022-12-31"),
			TermEndDate:         mustParseBrazilDate("2022-12-31"),
			LeadInsurerCode:     pointerOf("string"),
			LeadInsurerPolicyID: pointerOf("string"),
			MaxLMG: patrimonial.AmountDetails{
				Amount:   "2000.00",
				Currency: insurer.CurrencyBRL,
			},
			ProposalID: "string",
			Insureds: []patrimonial.Insured{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					BirthDate:                mustParseBrazilDate("1999-06-12"),
					PostCode:                 "10000000",
					Email:                    pointerOf("1Q0\\/@$-!a?}i#P|-Q$fW-`=\\+~WoleUG#i@3]w,Mwt*V!yTV9lHFK0.`gU;s?(x61rpudQ\\~aZ,taJB'y^C\"i9M?l.9>"),
					City:                     "string",
					State:                    "AC",
					Country:                  insurer.CountryCodeBrazil,
					Address:                  "string",
					AddressAdditionalInfo:    pointerOf("string"),
				},
			},
			Beneficiaries: &[]patrimonial.Beneficiary{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
				},
			},
			Principals: &[]patrimonial.Principal{
				{
					Identification:           "12345678900",
					IdentificationType:       insurer.IdentificationTypeCPF,
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 "10000000",
					Email:                    pointerOf("1|3|MoV#hPq+|h@>y;r1H}](9jjDSR]zsq|w6ygCmWJuhadnq`01f`MBvki4.$7uBli*_6=Pkr=f=&p8?iyB5rzsuB1x,i):w%Y7*\\\"kwD):F?glC}u]#V3X`3)oQv2.w"),
					City:                     "string",
					State:                    "AC",
					Country:                  insurer.CountryCodeBrazil,
					Address:                  "string",
					AddressAdditionalInfo:    pointerOf("Fundos"),
				},
			},
			Intermediaries: &[]patrimonial.Intermediary{
				{
					Type:                     patrimonial.IntermediaryTypeRepresentative,
					TypeOthers:               pointerOf("string"),
					Identification:           pointerOf("12345678900"),
					BrokerID:                 pointerOf("033948600"),
					IdentificationType:       pointerOf(insurer.IdentificationTypeCPF),
					IdentificationTypeOthers: pointerOf("RNE"),
					Name:                     "Nome Sobrenome",
					PostCode:                 pointerOf("10000000"),
					City:                     pointerOf("string"),
					State:                    pointerOf("string"),
					Country:                  pointerOf(insurer.CountryCodeBrazil),
					Address:                  pointerOf("string"),
				},
			},
			InsuredObjects: []patrimonial.InsuredObject{
				{
					Identification:     pointerOf("string"),
					Type:               patrimonial.InsuredObjectTypeContract,
					TypeAdditionalInfo: pointerOf("string"),
					Description:        "string",
					Amount: &patrimonial.AmountDetails{
						Amount:   "2000.00",
						Currency: insurer.CurrencyBRL,
					},
					Coverages: []patrimonial.InsuredObjectCoverage{
						{
							Branch:             "0111",
							Code:               patrimonial.CoverageCode("IMOVEL_BASICA"),
							Description:        pointerOf("string"),
							InternalCode:       pointerOf("string"),
							SusepProcessNumber: "string",
							LMI: patrimonial.AmountDetails{
								Amount:   "2000.00",
								Currency: insurer.CurrencyBRL,
							},
							TermStartDate:             mustParseBrazilDate("2022-12-31"),
							TermEndDate:               mustParseBrazilDate("2022-12-31"),
							IsMainCoverage:            pointerOf(true),
							Feature:                   patrimonial.CoverageFeatureMass,
							Type:                      patrimonial.CoverageTypeParametric,
							GracePeriod:               pointerOf(0),
							GracePeriodicity:          pointerOf(insurer.PeriodicityDay),
							GracePeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
							GracePeriodStartDate:      pointerOf(mustParseBrazilDate("2022-12-31")),
							GracePeriodEndDate:        pointerOf(mustParseBrazilDate("2022-12-31")),
							PremiumPeriodicity:        insurer.PremiumPeriodicityMonthly,
							PremiumPeriodicityOthers:  pointerOf("string"),
						},
						{
							Branch:             "0114",
							Code:               patrimonial.CoverageCode("IMOVEL_BASICA"),
							Description:        pointerOf("string"),
							InternalCode:       pointerOf("string"),
							SusepProcessNumber: "string",
							LMI: patrimonial.AmountDetails{
								Amount:   "2000.00",
								Currency: insurer.CurrencyBRL,
							},
							TermStartDate:             mustParseBrazilDate("2022-12-31"),
							TermEndDate:               mustParseBrazilDate("2022-12-31"),
							IsMainCoverage:            pointerOf(true),
							Feature:                   patrimonial.CoverageFeatureMass,
							Type:                      patrimonial.CoverageTypeParametric,
							GracePeriod:               pointerOf(0),
							GracePeriodicity:          pointerOf(insurer.PeriodicityDay),
							GracePeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
							GracePeriodStartDate:      pointerOf(mustParseBrazilDate("2022-12-31")),
							GracePeriodEndDate:        pointerOf(mustParseBrazilDate("2022-12-31")),
							PremiumPeriodicity:        insurer.PremiumPeriodicityMonthly,
							PremiumPeriodicityOthers:  pointerOf("string"),
						},
					},
				},
			},
			Coverages: &[]patrimonial.Coverage{
				{
					Branch:      "0111",
					Code:        patrimonial.CoverageCode("IMOVEL_BASICA"),
					Description: pointerOf("string"),
					Deductible: &patrimonial.CoverageDeductible{
						Type:               patrimonial.CoverageDeductibleTypeDeductible,
						TypeAdditionalInfo: pointerOf("string"),
						Amount: patrimonial.AmountDetails{
							Amount:   "2000.00",
							Currency: insurer.CurrencyBRL,
						},
						Period:               10,
						Periodicity:          insurer.PeriodicityDay,
						PeriodCountingMethod: pointerOf(insurer.PeriodCountingMethodBusinessDays),
						PeriodStartDate:      mustParseBrazilDate("2022-05-16"),
						PeriodEndDate:        mustParseBrazilDate("2022-05-17"),
						Description:          "Franquia de exemplo",
					},
					POS: &patrimonial.CoveragePOS{
						ApplicationType: insurer.ValueTypeValue,
						Description:     pointerOf("string"),
						MinValue: &patrimonial.AmountDetails{
							Amount:   "2000.00",
							Currency: insurer.CurrencyBRL,
						},
						MaxValue: &patrimonial.AmountDetails{
							Amount:   "2000.00",
							Currency: insurer.CurrencyBRL,
						},
						Percentage: pointerOf("10.00"),
						ValueOthers: &patrimonial.AmountDetails{
							Amount:   "2000.00",
							Currency: insurer.CurrencyBRL,
						},
					},
				},
			},
			CoinsuranceRetainedPercentage: pointerOf("10.00"),
			Coinsurers: &[]patrimonial.Coinsurer{
				{
					Identification:  "string",
					CededPercentage: "10.00",
				},
			},
			BranchInfo: &patrimonial.BranchInfo{
				BasicCoverageIndex: &patrimonial.BasicCoverageIndex{
					Index: "SIMPLES",
				},
				InsuredObjects: []patrimonial.SpecificInsuredObject{
					{
						Identification:   "string",
						PropertyType:     pointerOf(patrimonial.PropertyTypeHouse),
						StructuringType:  pointerOf(patrimonial.StructuringTypeCondominiumVertical),
						PostCode:         pointerOf("10000000"),
						BusinessActivity: pointerOf("1234567"),
					},
				},
			},
			Premium: patrimonial.Premium{
				PaymentsQuantity: 4,
				Amount: patrimonial.AmountDetails{
					Amount:   "2000.00",
					Currency: insurer.CurrencyBRL,
				},
				Coverages: []patrimonial.PremiumCoverage{
					{
						Branch:      "0111",
						Code:        patrimonial.CoverageCode("IMOVEL_BASICA"),
						Description: pointerOf("string"),
						PremiumAmount: patrimonial.AmountDetails{
							Amount:   "2000.00",
							Currency: insurer.CurrencyBRL,
						},
					},
				},
				Payments: []patrimonial.Payment{
					{
						MovementDate:           mustParseBrazilDate("2022-12-31"),
						MovementType:           insurer.PaymentMovementTypePremiumLiquidation,
						MovementOrigin:         pointerOf(insurer.PaymentMovementOriginDirectIssuance),
						MovementPaymentsNumber: "str",
						Amount: patrimonial.AmountDetails{
							Amount:   "2000.00",
							Currency: insurer.CurrencyBRL,
						},
						MaturityDate:             mustParseBrazilDate("2022-12-31"),
						TellerID:                 pointerOf("string"),
						TellerIDType:             pointerOf(insurer.IdentificationTypeCPF),
						TellerIDOthers:           pointerOf("string"),
						TellerName:               pointerOf("string"),
						FinancialInstitutionCode: pointerOf("string"),
						PaymentType:              pointerOf(insurer.PaymentTypeBankSlip),
						PaymentTypeOthers:        pointerOf("string"),
					},
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testPatrimonialPolicy).Error; err != nil {
		return fmt.Errorf("failed to create test patrimonial policy: %w", err)
	}

	testPatrimonialPolicyClaim := &patrimonial.Claim{
		ID:       uuid.MustParse("d4e5f6a7-b8c9-0123-def4-567890abcdef"),
		PolicyID: testPatrimonialPolicy.ID,
		Data: patrimonial.ClaimData{
			Identification:            "string",
			DocumentationDeliveryDate: pointerOf(mustParseBrazilDate("2022-12-31")),
			Status:                    patrimonial.ClaimStatusOpen,
			StatusAlterationDate:      mustParseBrazilDate("2022-12-31"),
			OccurrenceDate:            mustParseBrazilDate("2022-12-31"),
			WarningDate:               mustParseBrazilDate("2022-12-31"),
			ThirdPartyClaimDate:       pointerOf(mustParseBrazilDate("2022-12-31")),
			Amount: patrimonial.AmountDetails{
				Amount:   "2000.00",
				Currency: insurer.CurrencyBRL,
			},
			DenialJustification:            pointerOf(patrimonial.DenialJustificationExcludedRisk),
			DenialJustificationDescription: pointerOf("string"),
			Coverages: []patrimonial.ClaimCoverage{
				{
					InsuredObjectID:     pointerOf("string"),
					Branch:              "0111",
					Code:                patrimonial.CoverageCode("IMOVEL_BASICA"),
					Description:         pointerOf("string"),
					WarningDate:         pointerOf(mustParseBrazilDate("2022-12-31")),
					ThirdPartyClaimDate: pointerOf(mustParseBrazilDate("2022-12-31")),
				},
			},
		},
		CrossOrg:  true,
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testPatrimonialPolicyClaim).Error; err != nil {
		return fmt.Errorf("failed to create test patrimonial policy claim: %w", err)
	}

	return nil
}

func mustParseDateTime(s string) timeutil.DateTime {
	t, _ := time.Parse("2006-01-02T15:04:05Z", s)
	return timeutil.NewDateTime(t)
}
