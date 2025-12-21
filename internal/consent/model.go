package consent

import (
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

const (
	URNPrefix = "urn:mockinsurer:consent:"
)

var (
	ScopeID = goidc.NewDynamicScope("consent", func(requestedScope string) bool {
		return strings.HasPrefix(requestedScope, "consent:")
	})
	Scope = goidc.NewScope("consents")
)

type Consent struct {
	ID                     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Status                 Status
	Permissions            PermissionGroup `gorm:"serializer:json"`
	StatusUpdatedAt        timeutil.DateTime
	ExpiresAt              timeutil.DateTime
	UserIdentification     string
	UserRel                Relation
	OwnerID                *uuid.UUID
	BusinessIdentification *string
	BusinessRel            *Relation
	ClientID               string
	Rejection              *Rejection `gorm:"serializer:json"`

	OrgID     string
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (Consent) TableName() string {
	return "consents"
}

func (c *Consent) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (c Consent) URN() string {
	return URN(c.ID)
}

func (c Consent) HasPermissions(permissions []Permission) bool {
	for _, p := range permissions {
		if !slices.Contains(c.Permissions, p) {
			return false
		}
	}

	return true
}

type Status string

const (
	StatusAwaitingAuthorization Status = "AWAITING_AUTHORISATION"
	StatusAuthorized            Status = "AUTHORISED"
	StatusRejected              Status = "REJECTED"
	StatusConsumed              Status = "CONSUMED"
)

type Permission string

const (
	PermissionCapitalizationTitleEventsRead                             Permission = "CAPITALIZATION_TITLE_EVENTS_READ"
	PermissionCapitalizationTitlePlanInfoRead                           Permission = "CAPITALIZATION_TITLE_PLANINFO_READ"
	PermissionCapitalizationTitleRead                                   Permission = "CAPITALIZATION_TITLE_READ"
	PermissionCapitalizationTitleSettlementsRead                        Permission = "CAPITALIZATION_TITLE_SETTLEMENTS_READ"
	PermissionCapitalizationTitleWithdrawalCreate                       Permission = "CAPITALIZATION_TITLE_WITHDRAWAL_CREATE"
	PermissionClaimNotificationRequestDamageCreate                      Permission = "CLAIM_NOTIFICATION_REQUEST_DAMAGE_CREATE"
	PermissionClaimNotificationRequestPersonCreate                      Permission = "CLAIM_NOTIFICATION_REQUEST_PERSON_CREATE"
	PermissionContractLifePensionCreate                                 Permission = "CONTRACT_LIFE_PENSION_CREATE"
	PermissionContractLifePensionLeadCreate                             Permission = "CONTRACT_LIFE_PENSION_LEAD_CREATE"
	PermissionContractLifePensionLeadPortabilityCreate                  Permission = "CONTRACT_LIFE_PENSION_LEAD_PORTABILITY_CREATE"
	PermissionContractLifePensionLeadPortabilityUpdate                  Permission = "CONTRACT_LIFE_PENSION_LEAD_PORTABILITY_UPDATE"
	PermissionContractLifePensionLeadUpdate                             Permission = "CONTRACT_LIFE_PENSION_LEAD_UPDATE"
	PermissionContractLifePensionRead                                   Permission = "CONTRACT_LIFE_PENSION_READ"
	PermissionContractLifePensionUpdate                                 Permission = "CONTRACT_LIFE_PENSION_UPDATE"
	PermissionContractPensionPlanLeadCreate                             Permission = "CONTRACT_PENSION_PLAN_LEAD_CREATE"
	PermissionContractPensionPlanLeadPortabilityCreate                  Permission = "CONTRACT_PENSION_PLAN_LEAD_PORTABILITY_CREATE"
	PermissionContractPensionPlanLeadPortabilityUpdate                  Permission = "CONTRACT_PENSION_PLAN_LEAD_PORTABILITY_UPDATE"
	PermissionContractPensionPlanLeadUpdate                             Permission = "CONTRACT_PENSION_PLAN_LEAD_UPDATE"
	PermissionCustomersBusinessAdditionalInfoRead                       Permission = "CUSTOMERS_BUSINESS_ADDITIONALINFO_READ"
	PermissionCustomersBusinessIdentificationsRead                      Permission = "CUSTOMERS_BUSINESS_IDENTIFICATIONS_READ"
	PermissionCustomersBusinessQualificationRead                        Permission = "CUSTOMERS_BUSINESS_QUALIFICATION_READ"
	PermissionCustomersPersonalAdditionalInfoRead                       Permission = "CUSTOMERS_PERSONAL_ADDITIONALINFO_READ"
	PermissionCustomersPersonalIdentificationsRead                      Permission = "CUSTOMERS_PERSONAL_IDENTIFICATIONS_READ"
	PermissionCustomersPersonalQualificationRead                        Permission = "CUSTOMERS_PERSONAL_QUALIFICATION_READ"
	PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadClaimRead      Permission = "DAMAGES_AND_PEOPLE_ACCEPTANCE_AND_BRANCHES_ABROAD_CLAIM_READ"
	PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadPolicyInfoRead Permission = "DAMAGES_AND_PEOPLE_ACCEPTANCE_AND_BRANCHES_ABROAD_POLICYINFO_READ"
	PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadPremiumRead    Permission = "DAMAGES_AND_PEOPLE_ACCEPTANCE_AND_BRANCHES_ABROAD_PREMIUM_READ"
	PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadRead           Permission = "DAMAGES_AND_PEOPLE_ACCEPTANCE_AND_BRANCHES_ABROAD_READ"
	PermissionDamagesAndPeopleAutoClaimRead                             Permission = "DAMAGES_AND_PEOPLE_AUTO_CLAIM_READ"
	PermissionDamagesAndPeopleAutoPolicyInfoRead                        Permission = "DAMAGES_AND_PEOPLE_AUTO_POLICYINFO_READ"
	PermissionDamagesAndPeopleAutoPremiumRead                           Permission = "DAMAGES_AND_PEOPLE_AUTO_PREMIUM_READ"
	PermissionDamagesAndPeopleAutoRead                                  Permission = "DAMAGES_AND_PEOPLE_AUTO_READ"
	PermissionDamagesAndPeopleFinancialRisksClaimRead                   Permission = "DAMAGES_AND_PEOPLE_FINANCIAL_RISKS_CLAIM_READ"
	PermissionDamagesAndPeopleFinancialRisksPolicyInfoRead              Permission = "DAMAGES_AND_PEOPLE_FINANCIAL_RISKS_POLICYINFO_READ"
	PermissionDamagesAndPeopleFinancialRisksPremiumRead                 Permission = "DAMAGES_AND_PEOPLE_FINANCIAL_RISKS_PREMIUM_READ"
	PermissionDamagesAndPeopleFinancialRisksRead                        Permission = "DAMAGES_AND_PEOPLE_FINANCIAL_RISKS_READ"
	PermissionDamagesAndPeopleHousingClaimRead                          Permission = "DAMAGES_AND_PEOPLE_HOUSING_CLAIM_READ"
	PermissionDamagesAndPeopleHousingPolicyInfoRead                     Permission = "DAMAGES_AND_PEOPLE_HOUSING_POLICYINFO_READ"
	PermissionDamagesAndPeopleHousingPremiumRead                        Permission = "DAMAGES_AND_PEOPLE_HOUSING_PREMIUM_READ"
	PermissionDamagesAndPeopleHousingRead                               Permission = "DAMAGES_AND_PEOPLE_HOUSING_READ"
	PermissionDamagesAndPeoplePatrimonialClaimRead                      Permission = "DAMAGES_AND_PEOPLE_PATRIMONIAL_CLAIM_READ"
	PermissionDamagesAndPeoplePatrimonialPolicyInfoRead                 Permission = "DAMAGES_AND_PEOPLE_PATRIMONIAL_POLICYINFO_READ"
	PermissionDamagesAndPeoplePatrimonialPremiumRead                    Permission = "DAMAGES_AND_PEOPLE_PATRIMONIAL_PREMIUM_READ"
	PermissionDamagesAndPeoplePatrimonialRead                           Permission = "DAMAGES_AND_PEOPLE_PATRIMONIAL_READ"
	PermissionDamagesAndPeoplePersonClaimRead                           Permission = "DAMAGES_AND_PEOPLE_PERSON_CLAIM_READ"
	PermissionDamagesAndPeoplePersonPolicyInfoRead                      Permission = "DAMAGES_AND_PEOPLE_PERSON_POLICYINFO_READ"
	PermissionDamagesAndPeoplePersonPremiumRead                         Permission = "DAMAGES_AND_PEOPLE_PERSON_PREMIUM_READ"
	PermissionDamagesAndPeoplePersonRead                                Permission = "DAMAGES_AND_PEOPLE_PERSON_READ"
	PermissionDamagesAndPeopleResponsibilityClaimRead                   Permission = "DAMAGES_AND_PEOPLE_RESPONSIBILITY_CLAIM_READ"
	PermissionDamagesAndPeopleResponsibilityPolicyInfoRead              Permission = "DAMAGES_AND_PEOPLE_RESPONSIBILITY_POLICYINFO_READ"
	PermissionDamagesAndPeopleResponsibilityPremiumRead                 Permission = "DAMAGES_AND_PEOPLE_RESPONSIBILITY_PREMIUM_READ"
	PermissionDamagesAndPeopleResponsibilityRead                        Permission = "DAMAGES_AND_PEOPLE_RESPONSIBILITY_READ"
	PermissionDamagesAndPeopleRuralClaimRead                            Permission = "DAMAGES_AND_PEOPLE_RURAL_CLAIM_READ"
	PermissionDamagesAndPeopleRuralPolicyInfoRead                       Permission = "DAMAGES_AND_PEOPLE_RURAL_POLICYINFO_READ"
	PermissionDamagesAndPeopleRuralPremiumRead                          Permission = "DAMAGES_AND_PEOPLE_RURAL_PREMIUM_READ"
	PermissionDamagesAndPeopleRuralRead                                 Permission = "DAMAGES_AND_PEOPLE_RURAL_READ"
	PermissionDamagesAndPeopleTransportClaimRead                        Permission = "DAMAGES_AND_PEOPLE_TRANSPORT_CLAIM_READ"
	PermissionDamagesAndPeopleTransportPolicyInfoRead                   Permission = "DAMAGES_AND_PEOPLE_TRANSPORT_POLICYINFO_READ"
	PermissionDamagesAndPeopleTransportPremiumRead                      Permission = "DAMAGES_AND_PEOPLE_TRANSPORT_PREMIUM_READ"
	PermissionDamagesAndPeopleTransportRead                             Permission = "DAMAGES_AND_PEOPLE_TRANSPORT_READ"
	PermissionEndorsementRequestCreate                                  Permission = "ENDORSEMENT_REQUEST_CREATE"
	PermissionFinancialAssistanceContractInfoRead                       Permission = "FINANCIAL_ASSISTANCE_CONTRACTINFO_READ"
	PermissionFinancialAssistanceMovementsRead                          Permission = "FINANCIAL_ASSISTANCE_MOVEMENTS_READ"
	PermissionFinancialAssistanceRead                                   Permission = "FINANCIAL_ASSISTANCE_READ"
	PermissionLifePensionClaim                                          Permission = "LIFE_PENSION_CLAIM"
	PermissionLifePensionContractInfoRead                               Permission = "LIFE_PENSION_CONTRACTINFO_READ"
	PermissionLifePensionMovementsRead                                  Permission = "LIFE_PENSION_MOVEMENTS_READ"
	PermissionLifePensionPortabilitiesRead                              Permission = "LIFE_PENSION_PORTABILITIES_READ"
	PermissionLifePensionRead                                           Permission = "LIFE_PENSION_READ"
	PermissionLifePensionWithdrawalsRead                                Permission = "LIFE_PENSION_WITHDRAWALS_READ"
	PermissionPensionPlanClaim                                          Permission = "PENSION_PLAN_CLAIM"
	PermissionPensionPlanContractInfoRead                               Permission = "PENSION_PLAN_CONTRACTINFO_READ"
	PermissionPensionPlanMovementsRead                                  Permission = "PENSION_PLAN_MOVEMENTS_READ"
	PermissionPensionPlanPortabilitiesRead                              Permission = "PENSION_PLAN_PORTABILITIES_READ"
	PermissionPensionPlanRead                                           Permission = "PENSION_PLAN_READ"
	PermissionPensionPlanWithdrawalsRead                                Permission = "PENSION_PLAN_WITHDRAWALS_READ"
	PermissionPensionWithdrawalCreate                                   Permission = "PENSION_WITHDRAWAL_CREATE"
	PermissionPensionWithdrawalLeadCreate                               Permission = "PENSION_WITHDRAWAL_LEAD_CREATE"
	PermissionPersonWithdrawalCreate                                    Permission = "PERSON_WITHDRAWAL_CREATE"
	PermissionQuoteAcceptanceAndBranchesAbroadLeadCreate                Permission = "QUOTE_ACCEPTANCE_AND_BRANCHES_ABROAD_LEAD_CREATE"
	PermissionQuoteAcceptanceAndBranchesAbroadLeadUpdate                Permission = "QUOTE_ACCEPTANCE_AND_BRANCHES_ABROAD_LEAD_UPDATE"
	PermissionQuoteAutoCreate                                           Permission = "QUOTE_AUTO_CREATE"
	PermissionQuoteAutoLeadCreate                                       Permission = "QUOTE_AUTO_LEAD_CREATE"
	PermissionQuoteAutoLeadUpdate                                       Permission = "QUOTE_AUTO_LEAD_UPDATE"
	PermissionQuoteAutoRead                                             Permission = "QUOTE_AUTO_READ"
	PermissionQuoteAutoUpdate                                           Permission = "QUOTE_AUTO_UPDATE"
	PermissionQuoteCapitalizationTitleCreate                            Permission = "QUOTE_CAPITALIZATION_TITLE_CREATE"
	PermissionQuoteCapitalizationTitleLeadCreate                        Permission = "QUOTE_CAPITALIZATION_TITLE_LEAD_CREATE"
	PermissionQuoteCapitalizationTitleLeadUpdate                        Permission = "QUOTE_CAPITALIZATION_TITLE_LEAD_UPDATE"
	PermissionQuoteCapitalizationTitleRaffleCreate                      Permission = "QUOTE_CAPITALIZATION_TITLE_RAFFLE_CREATE"
	PermissionQuoteCapitalizationTitleRead                              Permission = "QUOTE_CAPITALIZATION_TITLE_READ"
	PermissionQuoteCapitalizationTitleUpdate                            Permission = "QUOTE_CAPITALIZATION_TITLE_UPDATE"
	PermissionQuoteFinancialRiskLeadCreate                              Permission = "QUOTE_FINANCIAL_RISK_LEAD_CREATE"
	PermissionQuoteFinancialRiskLeadUpdate                              Permission = "QUOTE_FINANCIAL_RISK_LEAD_UPDATE"
	PermissionQuoteHousingLeadCreate                                    Permission = "QUOTE_HOUSING_LEAD_CREATE"
	PermissionQuoteHousingLeadUpdate                                    Permission = "QUOTE_HOUSING_LEAD_UPDATE"
	PermissionQuotePatrimonialBusinessCreate                            Permission = "QUOTE_PATRIMONIAL_BUSINESS_CREATE"
	PermissionQuotePatrimonialBusinessRead                              Permission = "QUOTE_PATRIMONIAL_BUSINESS_READ"
	PermissionQuotePatrimonialBusinessUpdate                            Permission = "QUOTE_PATRIMONIAL_BUSINESS_UPDATE"
	PermissionQuotePatrimonialCondominiumCreate                         Permission = "QUOTE_PATRIMONIAL_CONDOMINIUM_CREATE"
	PermissionQuotePatrimonialCondominiumRead                           Permission = "QUOTE_PATRIMONIAL_CONDOMINIUM_READ"
	PermissionQuotePatrimonialCondominiumUpdate                         Permission = "QUOTE_PATRIMONIAL_CONDOMINIUM_UPDATE"
	PermissionQuotePatrimonialDiverseRisksCreate                        Permission = "QUOTE_PATRIMONIAL_DIVERSE_RISKS_CREATE"
	PermissionQuotePatrimonialDiverseRisksRead                          Permission = "QUOTE_PATRIMONIAL_DIVERSE_RISKS_READ"
	PermissionQuotePatrimonialDiverseRisksUpdate                        Permission = "QUOTE_PATRIMONIAL_DIVERSE_RISKS_UPDATE"
	PermissionQuotePatrimonialHomeCreate                                Permission = "QUOTE_PATRIMONIAL_HOME_CREATE"
	PermissionQuotePatrimonialHomeRead                                  Permission = "QUOTE_PATRIMONIAL_HOME_READ"
	PermissionQuotePatrimonialHomeUpdate                                Permission = "QUOTE_PATRIMONIAL_HOME_UPDATE"
	PermissionQuotePatrimonialLeadCreate                                Permission = "QUOTE_PATRIMONIAL_LEAD_CREATE"
	PermissionQuotePatrimonialLeadUpdate                                Permission = "QUOTE_PATRIMONIAL_LEAD_UPDATE"
	PermissionQuotePersonLeadCreate                                     Permission = "QUOTE_PERSON_LEAD_CREATE"
	PermissionQuotePersonLeadUpdate                                     Permission = "QUOTE_PERSON_LEAD_UPDATE"
	PermissionQuotePersonLifeCreate                                     Permission = "QUOTE_PERSON_LIFE_CREATE"
	PermissionQuotePersonLifeRead                                       Permission = "QUOTE_PERSON_LIFE_READ"
	PermissionQuotePersonLifeUpdate                                     Permission = "QUOTE_PERSON_LIFE_UPDATE"
	PermissionQuotePersonTravelCreate                                   Permission = "QUOTE_PERSON_TRAVEL_CREATE"
	PermissionQuotePersonTravelRead                                     Permission = "QUOTE_PERSON_TRAVEL_READ"
	PermissionQuotePersonTravelUpdate                                   Permission = "QUOTE_PERSON_TRAVEL_UPDATE"
	PermissionQuoteResponsibilityLeadCreate                             Permission = "QUOTE_RESPONSIBILITY_LEAD_CREATE"
	PermissionQuoteResponsibilityLeadUpdate                             Permission = "QUOTE_RESPONSIBILITY_LEAD_UPDATE"
	PermissionQuoteRuralLeadCreate                                      Permission = "QUOTE_RURAL_LEAD_CREATE"
	PermissionQuoteRuralLeadUpdate                                      Permission = "QUOTE_RURAL_LEAD_UPDATE"
	PermissionQuoteTransportLeadCreate                                  Permission = "QUOTE_TRANSPORT_LEAD_CREATE"
	PermissionQuoteTransportLeadUpdate                                  Permission = "QUOTE_TRANSPORT_LEAD_UPDATE"
	PermissionResourcesRead                                             Permission = "RESOURCES_READ"
)

type PermissionGroup []Permission

func (p PermissionGroup) HasAutoPermissions() bool {
	return slices.ContainsFunc(p, func(permission Permission) bool {
		return strings.HasPrefix(string(permission), "DAMAGES_AND_PEOPLE_AUTO")
	})
}

var (
	// Fase 2: Cadastro Pessoa Física
	PermissionGroupPersonalRegistrationData PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionCustomersPersonalIdentificationsRead,
		PermissionCustomersPersonalQualificationRead,
		PermissionCustomersPersonalAdditionalInfoRead,
	}

	// Fase 2: Cadastro Pessoa Jurídica
	PermissionGroupBusinessRegistrationData PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionCustomersBusinessIdentificationsRead,
		PermissionCustomersBusinessQualificationRead,
		PermissionCustomersBusinessAdditionalInfoRead,
	}

	// Fase 2: Títulos de Capitalização
	PermissionGroupCapitalizationTitle PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionCapitalizationTitleRead,
		PermissionCapitalizationTitlePlanInfoRead,
		PermissionCapitalizationTitleEventsRead,
		PermissionCapitalizationTitleSettlementsRead,
	}

	// Fase 2: Previdência Risco
	PermissionGroupPensionPlan PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionPensionPlanRead,
		PermissionPensionPlanContractInfoRead,
		PermissionPensionPlanMovementsRead,
		PermissionPensionPlanPortabilitiesRead,
		PermissionPensionPlanWithdrawalsRead,
		PermissionPensionPlanClaim,
	}

	// Fase 2: Previdência e Pessoas Sobrevivência
	PermissionGroupLifePension PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionLifePensionRead,
		PermissionLifePensionContractInfoRead,
		PermissionLifePensionMovementsRead,
		PermissionLifePensionPortabilitiesRead,
		PermissionLifePensionWithdrawalsRead,
		PermissionLifePensionClaim,
	}

	// Fase 2: Assistência Financeira
	PermissionGroupFinancialAssistance PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionFinancialAssistanceRead,
		PermissionFinancialAssistanceContractInfoRead,
		PermissionFinancialAssistanceMovementsRead,
	}

	// Fase 2: Danos e Pessoas - Patrimonial
	PermissionGroupDamagesAndPeoplePatrimonial PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeoplePatrimonialRead,
		PermissionDamagesAndPeoplePatrimonialPolicyInfoRead,
		PermissionDamagesAndPeoplePatrimonialPremiumRead,
		PermissionDamagesAndPeoplePatrimonialClaimRead,
	}

	// Fase 2: Danos e Pessoas - Responsabilidade
	PermissionGroupDamagesAndPeopleResponsibility PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeopleResponsibilityRead,
		PermissionDamagesAndPeopleResponsibilityPolicyInfoRead,
		PermissionDamagesAndPeopleResponsibilityPremiumRead,
		PermissionDamagesAndPeopleResponsibilityClaimRead,
	}

	// Fase 2: Danos e Pessoas - Transportes
	PermissionGroupDamagesAndPeopleTransport PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeopleTransportRead,
		PermissionDamagesAndPeopleTransportPolicyInfoRead,
		PermissionDamagesAndPeopleTransportPremiumRead,
		PermissionDamagesAndPeopleTransportClaimRead,
	}

	// Fase 2: Danos e Pessoas - Riscos Financeiros
	PermissionGroupDamagesAndPeopleFinancialRisks PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeopleFinancialRisksRead,
		PermissionDamagesAndPeopleFinancialRisksPolicyInfoRead,
		PermissionDamagesAndPeopleFinancialRisksPremiumRead,
		PermissionDamagesAndPeopleFinancialRisksClaimRead,
	}

	// Fase 2: Danos e Pessoas - Rural
	PermissionGroupDamagesAndPeopleRural PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeopleRuralRead,
		PermissionDamagesAndPeopleRuralPolicyInfoRead,
		PermissionDamagesAndPeopleRuralPremiumRead,
		PermissionDamagesAndPeopleRuralClaimRead,
	}

	// Fase 2: Danos e Pessoas - Automóveis
	PermissionGroupDamagesAndPeopleAuto PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeopleAutoRead,
		PermissionDamagesAndPeopleAutoPolicyInfoRead,
		PermissionDamagesAndPeopleAutoPremiumRead,
		PermissionDamagesAndPeopleAutoClaimRead,
	}

	// Fase 2: Danos e Pessoas - Habitacional
	PermissionGroupDamagesAndPeopleHousing PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeopleHousingRead,
		PermissionDamagesAndPeopleHousingPolicyInfoRead,
		PermissionDamagesAndPeopleHousingPremiumRead,
		PermissionDamagesAndPeopleHousingClaimRead,
	}

	// Fase 2: Danos e Pessoas - Aceitação e Sucursal no exterior
	PermissionGroupDamagesAndPeopleAcceptanceAndBranchesAbroad PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadRead,
		PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadPolicyInfoRead,
		PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadPremiumRead,
		PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadClaimRead,
	}

	// Fase 2: Danos e Pessoas - Pessoas
	PermissionGroupDamagesAndPeoplePerson PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionDamagesAndPeoplePersonRead,
		PermissionDamagesAndPeoplePersonPolicyInfoRead,
		PermissionDamagesAndPeoplePersonPremiumRead,
		PermissionDamagesAndPeoplePersonClaimRead,
	}

	// Fase 3: Aviso de Sinistro - Danos
	PermissionGroupClaimNotificationRequestDamage PermissionGroup = []Permission{
		PermissionClaimNotificationRequestDamageCreate,
	}

	// Fase 3: Aviso de Sinistro - Pessoas
	PermissionGroupClaimNotificationRequestPerson PermissionGroup = []Permission{
		PermissionClaimNotificationRequestPersonCreate,
	}

	// Fase 3: Endosso
	PermissionGroupEndorsementRequest PermissionGroup = []Permission{
		PermissionEndorsementRequestCreate,
	}

	// Fase 3: Cotação Patrimonial Lead
	PermissionGroupQuotePatrimonialLead PermissionGroup = []Permission{
		PermissionQuotePatrimonialLeadCreate,
		PermissionQuotePatrimonialLeadUpdate,
	}

	// Fase 3: Cotação Patrimonial Home
	PermissionGroupQuotePatrimonialHome PermissionGroup = []Permission{
		PermissionQuotePatrimonialHomeRead,
		PermissionQuotePatrimonialHomeCreate,
		PermissionQuotePatrimonialHomeUpdate,
	}

	// Fase 3: Cotação Patrimonial Condominium
	PermissionGroupQuotePatrimonialCondominium PermissionGroup = []Permission{
		PermissionQuotePatrimonialCondominiumRead,
		PermissionQuotePatrimonialCondominiumCreate,
		PermissionQuotePatrimonialCondominiumUpdate,
	}

	// Fase 3: Cotação Patrimonial Business
	PermissionGroupQuotePatrimonialBusiness PermissionGroup = []Permission{
		PermissionQuotePatrimonialBusinessRead,
		PermissionQuotePatrimonialBusinessCreate,
		PermissionQuotePatrimonialBusinessUpdate,
	}

	// Fase 3: Cotação Patrimonial Diverse Risks
	PermissionGroupQuotePatrimonialDiverseRisks PermissionGroup = []Permission{
		PermissionQuotePatrimonialDiverseRisksRead,
		PermissionQuotePatrimonialDiverseRisksCreate,
		PermissionQuotePatrimonialDiverseRisksUpdate,
	}

	// Fase 3: Cotação Aceitação e Sucursal no exterior
	PermissionGroupQuoteAcceptanceAndBranchesAbroadLead PermissionGroup = []Permission{
		PermissionQuoteAcceptanceAndBranchesAbroadLeadCreate,
		PermissionQuoteAcceptanceAndBranchesAbroadLeadUpdate,
	}

	// Fase 3: Cotação Auto Lead
	PermissionGroupQuoteAutoLead PermissionGroup = []Permission{
		PermissionQuoteAutoLeadCreate,
		PermissionQuoteAutoLeadUpdate,
	}

	// Fase 3: Cotação Auto
	PermissionGroupQuoteAuto PermissionGroup = []Permission{
		PermissionQuoteAutoRead,
		PermissionQuoteAutoCreate,
		PermissionQuoteAutoUpdate,
	}

	// Fase 3: Cotação Riscos Financeiros Lead
	PermissionGroupQuoteFinancialRiskLead PermissionGroup = []Permission{
		PermissionQuoteFinancialRiskLeadCreate,
		PermissionQuoteFinancialRiskLeadUpdate,
	}

	// Fase 3: Cotação Habitacional Lead
	PermissionGroupQuoteHousingLead PermissionGroup = []Permission{
		PermissionQuoteHousingLeadCreate,
		PermissionQuoteHousingLeadUpdate,
	}

	// Fase 3: Cotação Responsabilidade Lead
	PermissionGroupQuoteResponsibilityLead PermissionGroup = []Permission{
		PermissionQuoteResponsibilityLeadCreate,
		PermissionQuoteResponsibilityLeadUpdate,
	}

	// Fase 3: Cotação Rural Lead
	PermissionGroupQuoteRuralLead PermissionGroup = []Permission{
		PermissionQuoteRuralLeadCreate,
		PermissionQuoteRuralLeadUpdate,
	}

	// Fase 3: Cotação Transportes Lead
	PermissionGroupQuoteTransportLead PermissionGroup = []Permission{
		PermissionQuoteTransportLeadCreate,
		PermissionQuoteTransportLeadUpdate,
	}

	// Fase 3: Cotação Pessoas Lead
	PermissionGroupQuotePersonLead PermissionGroup = []Permission{
		PermissionQuotePersonLeadCreate,
		PermissionQuotePersonLeadUpdate,
	}

	// Fase 3: Cotação Pessoas Life
	PermissionGroupQuotePersonLife PermissionGroup = []Permission{
		PermissionQuotePersonLifeRead,
		PermissionQuotePersonLifeCreate,
		PermissionQuotePersonLifeUpdate,
	}

	// Fase 3: Cotação Pessoas Travel
	PermissionGroupQuotePersonTravel PermissionGroup = []Permission{
		PermissionQuotePersonTravelRead,
		PermissionQuotePersonTravelCreate,
		PermissionQuotePersonTravelUpdate,
	}

	// Fase 3: Cotação Títulos de Capitalização Lead
	PermissionGroupQuoteCapitalizationTitleLead PermissionGroup = []Permission{
		PermissionQuoteCapitalizationTitleLeadCreate,
		PermissionQuoteCapitalizationTitleLeadUpdate,
	}

	// Fase 3: Cotação Títulos de Capitalização
	PermissionGroupQuoteCapitalizationTitle PermissionGroup = []Permission{
		PermissionQuoteCapitalizationTitleRead,
		PermissionQuoteCapitalizationTitleCreate,
		PermissionQuoteCapitalizationTitleUpdate,
	}

	// Fase 3: Sorteio Títulos de Capitalização
	PermissionGroupQuoteCapitalizationTitleRaffle PermissionGroup = []Permission{
		PermissionQuoteCapitalizationTitleRaffleCreate,
	}

	// Fase 3: Contratação Previdência Risco Lead
	PermissionGroupContractPensionPlanLead PermissionGroup = []Permission{
		PermissionContractPensionPlanLeadCreate,
		PermissionContractPensionPlanLeadUpdate,
	}

	// Fase 3: Portabilidade Previdência Risco
	PermissionGroupContractPensionPlanLeadPortability PermissionGroup = []Permission{
		PermissionContractPensionPlanLeadPortabilityCreate,
		PermissionContractPensionPlanLeadPortabilityUpdate,
	}

	// Fase 3: Contratação Previdência Sobrevivência Lead
	PermissionGroupContractLifePensionLead PermissionGroup = []Permission{
		PermissionContractLifePensionLeadCreate,
		PermissionContractLifePensionLeadUpdate,
	}

	// Fase 3: Contratação Previdência Sobrevivência
	PermissionGroupContractLifePension PermissionGroup = []Permission{
		PermissionContractLifePensionCreate,
		PermissionContractLifePensionUpdate,
		PermissionContractLifePensionRead,
	}

	// Fase 3: Portabilidade Previdência Sobrevivência
	PermissionGroupContractLifePensionLeadPortability PermissionGroup = []Permission{
		PermissionContractLifePensionLeadPortabilityCreate,
		PermissionContractLifePensionLeadPortabilityUpdate,
	}

	// Fase 3: Resgate Previdência
	PermissionGroupPensionWithdrawal PermissionGroup = []Permission{
		PermissionPensionWithdrawalCreate,
	}

	// Fase 3: Resgate Previdência Lead
	PermissionGroupPensionWithdrawalLead PermissionGroup = []Permission{
		PermissionPensionWithdrawalLeadCreate,
	}

	// Fase 3: Resgate Capitalização
	PermissionGroupCapitalizationTitleWithdrawal PermissionGroup = []Permission{
		PermissionCapitalizationTitleWithdrawalCreate,
	}

	// Fase 3: Resgate Pessoas
	PermissionGroupPersonWithdrawal PermissionGroup = []Permission{
		PermissionPersonWithdrawalCreate,
	}

	PermissionGroupPhase2 PermissionGroup = []Permission{
		PermissionResourcesRead,
		PermissionCustomersPersonalIdentificationsRead,
		PermissionCustomersBusinessIdentificationsRead,
		PermissionCustomersPersonalAdditionalInfoRead,
		PermissionCustomersBusinessAdditionalInfoRead,
		PermissionCustomersPersonalQualificationRead,
		PermissionCustomersBusinessQualificationRead,
		PermissionCapitalizationTitleRead,
		PermissionCapitalizationTitlePlanInfoRead,
		PermissionCapitalizationTitleEventsRead,
		PermissionCapitalizationTitleSettlementsRead,
		PermissionPensionPlanRead,
		PermissionPensionPlanContractInfoRead,
		PermissionPensionPlanMovementsRead,
		PermissionPensionPlanPortabilitiesRead,
		PermissionPensionPlanWithdrawalsRead,
		PermissionPensionPlanClaim,
		PermissionLifePensionRead,
		PermissionLifePensionContractInfoRead,
		PermissionLifePensionMovementsRead,
		PermissionLifePensionPortabilitiesRead,
		PermissionLifePensionWithdrawalsRead,
		PermissionLifePensionClaim,
		PermissionFinancialAssistanceRead,
		PermissionFinancialAssistanceContractInfoRead,
		PermissionFinancialAssistanceMovementsRead,
		PermissionDamagesAndPeoplePatrimonialRead,
		PermissionDamagesAndPeoplePatrimonialPolicyInfoRead,
		PermissionDamagesAndPeoplePatrimonialPremiumRead,
		PermissionDamagesAndPeoplePatrimonialClaimRead,
		PermissionDamagesAndPeopleResponsibilityRead,
		PermissionDamagesAndPeopleResponsibilityPolicyInfoRead,
		PermissionDamagesAndPeopleResponsibilityPremiumRead,
		PermissionDamagesAndPeopleResponsibilityClaimRead,
		PermissionDamagesAndPeopleTransportRead,
		PermissionDamagesAndPeopleTransportPolicyInfoRead,
		PermissionDamagesAndPeopleTransportPremiumRead,
		PermissionDamagesAndPeopleTransportClaimRead,
		PermissionDamagesAndPeopleFinancialRisksRead,
		PermissionDamagesAndPeopleFinancialRisksPolicyInfoRead,
		PermissionDamagesAndPeopleFinancialRisksPremiumRead,
		PermissionDamagesAndPeopleFinancialRisksClaimRead,
		PermissionDamagesAndPeopleRuralRead,
		PermissionDamagesAndPeopleRuralPolicyInfoRead,
		PermissionDamagesAndPeopleRuralPremiumRead,
		PermissionDamagesAndPeopleRuralClaimRead,
		PermissionDamagesAndPeopleAutoRead,
		PermissionDamagesAndPeopleAutoPolicyInfoRead,
		PermissionDamagesAndPeopleAutoPremiumRead,
		PermissionDamagesAndPeopleAutoClaimRead,
		PermissionDamagesAndPeopleHousingRead,
		PermissionDamagesAndPeopleHousingPolicyInfoRead,
		PermissionDamagesAndPeopleHousingPremiumRead,
		PermissionDamagesAndPeopleHousingClaimRead,
		PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadRead,
		PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadPolicyInfoRead,
		PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadPremiumRead,
		PermissionDamagesAndPeopleAcceptanceAndBranchesAbroadClaimRead,
		PermissionDamagesAndPeoplePersonRead,
		PermissionDamagesAndPeoplePersonPolicyInfoRead,
		PermissionDamagesAndPeoplePersonPremiumRead,
		PermissionDamagesAndPeoplePersonClaimRead,
	}

	PermissionGroupPhase3 PermissionGroup = []Permission{
		PermissionClaimNotificationRequestDamageCreate,
		PermissionClaimNotificationRequestPersonCreate,
		PermissionEndorsementRequestCreate,
		PermissionQuotePatrimonialLeadCreate,
		PermissionQuotePatrimonialLeadUpdate,
		PermissionQuotePatrimonialHomeRead,
		PermissionQuotePatrimonialHomeCreate,
		PermissionQuotePatrimonialHomeUpdate,
		PermissionQuotePatrimonialCondominiumRead,
		PermissionQuotePatrimonialCondominiumCreate,
		PermissionQuotePatrimonialCondominiumUpdate,
		PermissionQuotePatrimonialBusinessRead,
		PermissionQuotePatrimonialBusinessCreate,
		PermissionQuotePatrimonialBusinessUpdate,
		PermissionQuotePatrimonialDiverseRisksRead,
		PermissionQuotePatrimonialDiverseRisksCreate,
		PermissionQuotePatrimonialDiverseRisksUpdate,
		PermissionQuoteAcceptanceAndBranchesAbroadLeadCreate,
		PermissionQuoteAcceptanceAndBranchesAbroadLeadUpdate,
		PermissionQuoteAutoLeadCreate,
		PermissionQuoteAutoLeadUpdate,
		PermissionQuoteAutoRead,
		PermissionQuoteAutoCreate,
		PermissionQuoteAutoUpdate,
		PermissionQuoteFinancialRiskLeadCreate,
		PermissionQuoteFinancialRiskLeadUpdate,
		PermissionQuoteHousingLeadCreate,
		PermissionQuoteHousingLeadUpdate,
		PermissionQuoteResponsibilityLeadCreate,
		PermissionQuoteResponsibilityLeadUpdate,
		PermissionQuoteRuralLeadCreate,
		PermissionQuoteRuralLeadUpdate,
		PermissionQuoteTransportLeadCreate,
		PermissionQuoteTransportLeadUpdate,
		PermissionQuotePersonLeadCreate,
		PermissionQuotePersonLeadUpdate,
		PermissionQuotePersonLifeRead,
		PermissionQuotePersonLifeCreate,
		PermissionQuotePersonLifeUpdate,
		PermissionQuotePersonTravelRead,
		PermissionQuotePersonTravelCreate,
		PermissionQuotePersonTravelUpdate,
		PermissionQuoteCapitalizationTitleLeadCreate,
		PermissionQuoteCapitalizationTitleLeadUpdate,
		PermissionQuoteCapitalizationTitleRead,
		PermissionQuoteCapitalizationTitleCreate,
		PermissionQuoteCapitalizationTitleUpdate,
		PermissionQuoteCapitalizationTitleRaffleCreate,
		PermissionContractPensionPlanLeadCreate,
		PermissionContractPensionPlanLeadUpdate,
		PermissionContractPensionPlanLeadPortabilityCreate,
		PermissionContractPensionPlanLeadPortabilityUpdate,
		PermissionContractLifePensionLeadCreate,
		PermissionContractLifePensionLeadUpdate,
		PermissionContractLifePensionCreate,
		PermissionContractLifePensionUpdate,
		PermissionContractLifePensionRead,
		PermissionContractLifePensionLeadPortabilityCreate,
		PermissionContractLifePensionLeadPortabilityUpdate,
		PermissionPensionWithdrawalCreate,
		PermissionPensionWithdrawalLeadCreate,
		PermissionCapitalizationTitleWithdrawalCreate,
		PermissionPersonWithdrawalCreate,
	}
)

var PermissionGroups = []PermissionGroup{
	// Fase 2: Groups
	PermissionGroupPersonalRegistrationData,
	PermissionGroupBusinessRegistrationData,
	PermissionGroupCapitalizationTitle,
	PermissionGroupPensionPlan,
	PermissionGroupLifePension,
	PermissionGroupFinancialAssistance,
	PermissionGroupDamagesAndPeoplePatrimonial,
	PermissionGroupDamagesAndPeopleResponsibility,
	PermissionGroupDamagesAndPeopleTransport,
	PermissionGroupDamagesAndPeopleFinancialRisks,
	PermissionGroupDamagesAndPeopleRural,
	PermissionGroupDamagesAndPeopleAuto,
	PermissionGroupDamagesAndPeopleHousing,
	PermissionGroupDamagesAndPeopleAcceptanceAndBranchesAbroad,
	PermissionGroupDamagesAndPeoplePerson,
	// Fase 3: Groups
	PermissionGroupClaimNotificationRequestDamage,
	PermissionGroupClaimNotificationRequestPerson,
	PermissionGroupEndorsementRequest,
	PermissionGroupQuotePatrimonialLead,
	PermissionGroupQuotePatrimonialHome,
	PermissionGroupQuotePatrimonialCondominium,
	PermissionGroupQuotePatrimonialBusiness,
	PermissionGroupQuotePatrimonialDiverseRisks,
	PermissionGroupQuoteAcceptanceAndBranchesAbroadLead,
	PermissionGroupQuoteAutoLead,
	PermissionGroupQuoteAuto,
	PermissionGroupQuoteFinancialRiskLead,
	PermissionGroupQuoteHousingLead,
	PermissionGroupQuoteResponsibilityLead,
	PermissionGroupQuoteRuralLead,
	PermissionGroupQuoteTransportLead,
	PermissionGroupQuotePersonLead,
	PermissionGroupQuotePersonLife,
	PermissionGroupQuotePersonTravel,
	PermissionGroupQuoteCapitalizationTitleLead,
	PermissionGroupQuoteCapitalizationTitle,
	PermissionGroupQuoteCapitalizationTitleRaffle,
	PermissionGroupContractPensionPlanLead,
	PermissionGroupContractPensionPlanLeadPortability,
	PermissionGroupContractLifePensionLead,
	PermissionGroupContractLifePension,
	PermissionGroupContractLifePensionLeadPortability,
	PermissionGroupPensionWithdrawal,
	PermissionGroupPensionWithdrawalLead,
	PermissionGroupCapitalizationTitleWithdrawal,
	PermissionGroupPersonWithdrawal,
}

type Rejection struct {
	By                   RejectedBy          `json:"rejectedBy"`
	ReasonCode           RejectionReasonCode `json:"reasonCode"`
	ReasonAdditionalInfo *string             `json:"reasonAdditionalInfo,omitempty"`
}

type RejectedBy string

const (
	RejectedByUser  RejectedBy = "USER"
	RejectedByASPSP RejectedBy = "ASPSP"
	RejectedByTPP   RejectedBy = "TPP"
)

type RejectionReasonCode string

const (
	RejectionReasonCodeConsentExpired           RejectionReasonCode = "CONSENT_EXPIRED"
	RejectionReasonCodeCustomerManuallyRejected RejectionReasonCode = "CUSTOMER_MANUALLY_REJECTED"
	RejectionReasonCodeCustomerManuallyRevoked  RejectionReasonCode = "CUSTOMER_MANUALLY_REVOKED"
	RejectionReasonCodeConsentMaxDateReached    RejectionReasonCode = "CONSENT_MAX_DATE_REACHED"
	RejectionReasonCodeConsentTechnicalIssue    RejectionReasonCode = "CONSENT_TECHNICAL_ISSUE"
	RejectionReasonCodeInternalSecurityReason   RejectionReasonCode = "INTERNAL_SECURITY_REASON"
)

type Document struct {
	Identification string   `json:"identification"`
	Rel            Relation `json:"rel"`
}

type Relation string

const (
	RelationCPF  Relation = "CPF"
	RelationCNPJ Relation = "CNPJ"
)

type Filter struct {
	OwnerID string
}
