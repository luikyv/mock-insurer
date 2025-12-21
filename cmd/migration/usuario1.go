package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/auto"
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
							GracePeriodicity:          pointerOf(auto.PeriodicityDay),
							GracePeriodCountingMethod: pointerOf(auto.PeriodCountingMethodBusinessDays),
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
						Periodicity:                        pointerOf(auto.PeriodicityDay),
						PeriodCountingMethod:               pointerOf(auto.PeriodCountingMethodBusinessDays),
						PeriodStartDate:                    pointerOf(mustParseBrazilDate("2022-05-16")),
						PeriodEndDate:                      pointerOf(mustParseBrazilDate("2022-05-17")),
						Description:                        pointerOf("Franquia de exemplo"),
						HasDeductibleOverTotalCompensation: pointerOf(true),
					},
					POS: &auto.CoveragePOS{
						ApplicationType: auto.CoveragePOSApplicationTypeValue,
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

	return nil
}
