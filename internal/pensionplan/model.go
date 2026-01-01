package pensionplan

import (
	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/resource"
	"github.com/luikyv/mock-insurer/internal/strutil"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

var (
	Scope = goidc.NewScope("insurance-pension-plan")
)

type Contract struct {
	ID        string `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      ContractData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (Contract) TableName() string {
	return "insurance_pension_plan_contracts"
}

func (c *Contract) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = strutil.Random(60)
	}
	return nil
}

type ConsentContract struct {
	ConsentID  uuid.UUID
	ContractID string
	OwnerID    uuid.UUID
	Status     resource.Status
	Contract   *Contract
	OrgID      string
	CreatedAt  timeutil.DateTime
	UpdatedAt  timeutil.DateTime
}

func (ConsentContract) TableName() string {
	return "consent_insurance_pension_plan_contracts"
}

type Portability struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	ContractID string
	Data       PortabilityData `gorm:"serializer:json"`
	OrgID      string
	CrossOrg   bool
	CreatedAt  timeutil.DateTime
	UpdatedAt  timeutil.DateTime
}

func (Portability) TableName() string {
	return "insurance_pension_plan_portabilities"
}

func (p *Portability) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Withdrawal struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	ContractID string
	Data       WithdrawalData `gorm:"serializer:json"`
	OrgID      string
	CrossOrg   bool
	CreatedAt  timeutil.DateTime
	UpdatedAt  timeutil.DateTime
}

func (Withdrawal) TableName() string {
	return "insurance_pension_plan_withdrawals"
}

func (w *Withdrawal) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

type Claim struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	ContractID string
	Data       ClaimData `gorm:"serializer:json"`
	OrgID      string
	CrossOrg   bool
	CreatedAt  timeutil.DateTime
	UpdatedAt  timeutil.DateTime
}

func (Claim) TableName() string {
	return "insurance_pension_plan_claims"
}

func (c *Claim) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type ContractData struct {
	ContractingType       ContractingType        `json:"contractingType"`
	PlanType              *PlanType              `json:"planType,omitempty"`
	Documents             []Document             `json:"documents"`
	MovementContributions []MovementContribution `json:"movementContributions"`
	MovementBenefits      []MovementBenefit      `json:"movementBenefits"`
}

type Document struct {
	CertificateID      string              `json:"certificateId"`
	ContractID         *string             `json:"contractId,omitempty"`
	ProposalID         string              `json:"proposalId"`
	EffectiveDateStart timeutil.BrazilDate `json:"effectiveDateStart"`
	EffectiveDateEnd   timeutil.BrazilDate `json:"effectiveDateEnd"`
	Insureds           []Insured           `json:"insureds"`
	Intermediary       *[]Intermediary     `json:"intermediary,omitempty"`
	Beneficiary        *[]Beneficiary      `json:"beneficiary,omitempty"`
	Plans              Plans               `json:"plans"`
}

type ContractingType string

const (
	ContractingTypeCollective ContractingType = "COLETIVO"
	ContractingTypeIndividual ContractingType = "INDIVIDUAL"
)

type PlanType string

const (
	PlanTypeApproved                   PlanType = "AVERBADO"
	PlanTypeEstablishedContributory    PlanType = "INSTITUIDO_CONTRIBUTARIO"
	PlanTypeEstablishedNonContributory PlanType = "INSTITUIDO_NAO_CONTRIBUTARIO"
)

type Periodicity string

const (
	PeriodicityMonthly       Periodicity = "MENSAL"
	PeriodicityBimonthly     Periodicity = "BIMESTRAL"
	PeriodicityQuarterly     Periodicity = "TRIMESTRAL"
	PeriodicityQuadrimestral Periodicity = "QUADRIMESTRAL"
	PeriodicitySemiannual    Periodicity = "SEMESTRAL"
	PeriodicityAnnual        Periodicity = "ANUAL" //nolint:misspell
	PeriodicitySporadic      Periodicity = "ESPORADICA"
	PeriodicityOneTime       Periodicity = "PAGAMENTO_UNICO"
	PeriodicityOthers        Periodicity = "OUTROS"
)

type TaxRegime string

const (
	TaxRegimeProgressive TaxRegime = "PROGRESSIVO" //nolint:misspell
	TaxRegimeRegressive  TaxRegime = "REGRESSIVO"  //nolint:misspell
)

type Insured struct {
	DocumentType          DocumentType               `json:"documentType"`
	DocumentTypeOthers    *string                    `json:"documentTypeOthers,omitempty"`
	DocumentNumber        string                     `json:"documentNumber"`
	Name                  string                     `json:"name"`
	BirthDate             timeutil.BrazilDate        `json:"birthDate"`
	Gender                Gender                     `json:"gender"`
	PostCode              string                     `json:"postCode"`
	Email                 *string                    `json:"email,omitempty"`
	TownName              string                     `json:"townName"`
	CountrySubDivision    insurer.CountrySubDivision `json:"countrySubDivision"`
	CountryCode           insurer.CountryCode        `json:"countryCode"`
	Address               string                     `json:"address"`
	AddressAdditionalInfo *string                    `json:"addressAdditionalInfo,omitempty"`
}

type DocumentType string

const (
	DocumentTypeCPF DocumentType = "CPF"
	DocumentTypeRG  DocumentType = "RG"
)

type Gender string

const (
	GenderMale    Gender = "MASCULINO"
	GenderFemale  Gender = "FEMININO"
	GenderUnknown Gender = "NAO_INFORMADO"
)

type Intermediary struct {
	Type               IntermediaryType            `json:"type"`
	TypeOthers         *string                     `json:"typeOthers,omitempty"`
	DocumentNumber     *string                     `json:"documentNumber,omitempty"`
	IntermediaryID     *string                     `json:"intermediaryId,omitempty"`
	DocumentType       *insurer.IdentificationType `json:"documentType,omitempty"`
	DocumentTypeOthers *string                     `json:"documentTypeOthers,omitempty"`
	Name               *string                     `json:"name,omitempty"`
	PostCode           *string                     `json:"postCode,omitempty"`
	TownName           *string                     `json:"townName,omitempty"`
	CountrySubDivision *insurer.CountrySubDivision `json:"countrySubDivision,omitempty"`
	CountryCode        *insurer.CountryCode        `json:"countryCode,omitempty"`
	Address            *string                     `json:"address,omitempty"`
	AdditionalInfo     *string                     `json:"additionalInfo,omitempty"`
}

type IntermediaryType string

const (
	IntermediaryTypeBroker                               IntermediaryType = "CORRETOR"
	IntermediaryTypeRepresentative                       IntermediaryType = "REPRESENTANTE"
	IntermediaryTypeInsuredEstablisherApprovedInstituted IntermediaryType = "ESTIPULANTE_AVERBADOR_INSTITUIDOR"
	IntermediaryTypeCorrespondent                        IntermediaryType = "CORRESPONDENTE"
	IntermediaryTypeMicroinsuranceAgent                  IntermediaryType = "AGENTE_DE_MICROSSEGUROS"
	IntermediaryTypeOthers                               IntermediaryType = "OUTROS"
)

type Beneficiary struct {
	DocumentNumber          string       `json:"documentNumber"`
	DocumentType            DocumentType `json:"documentType"`
	DocumentTypeOthers      *string      `json:"documentTypeOthers,omitempty"`
	Name                    string       `json:"name"`
	ParticipationPercentage string       `json:"participationPercentage"`
}

type Kinship string

const (
	KinshipSpouse     Kinship = "CONJUGE"
	KinshipParent     Kinship = "PAIS"
	KinshipSibling    Kinship = "IRMAOS"
	KinshipChild      Kinship = "FILHOS"
	KinshipGrandchild Kinship = "NETOS"
	KinshipOther      Kinship = "OUTROS"
)

type Plans struct {
	Coverages []Coverage `json:"coverages"`
	Grace     *[]Grace   `json:"grace,omitempty"`
}

type Coverage struct {
	CoverageCode             string                 `json:"coverageCode"`
	CoverageName             string                 `json:"coverageName"`
	SusepProcessNumber       string                 `json:"susepProcessNumber"`
	StructureModality        StructureModality      `json:"structureModality"`
	BenefitAmount            insurer.AmountDetails  `json:"benefitAmount"`
	BenefitPaymentMethod     BenefitPaymentMethod   `json:"benefitPaymentMethod"`
	ChargedAmount            insurer.AmountDetails  `json:"chargedAmount"`
	ContributionAmount       insurer.AmountDetails  `json:"contributionAmount"`
	FinancialRegime          FinancialRegime        `json:"financialRegime"`
	PricingMethod            PricingMethod          `json:"pricingMethod"`
	PricingMethodDescription *string                `json:"pricingMethodDescription,omitempty"`
	Periodicity              Periodicity            `json:"periodicity"`
	PeriodicityOthers        *string                `json:"periodicityOthers,omitempty"`
	LockedPlan               bool                   `json:"lockedPlan"`
	BiometricTable           *BiometricTable        `json:"biometricTable,omitempty"`
	RentsInterestRate        *string                `json:"rentsInterestRate,omitempty"`
	TermStartDate            timeutil.BrazilDate    `json:"termStartDate"`
	TermEndDate              timeutil.BrazilDate    `json:"termEndDate"`
	UpdateIndex              UpdateIndex            `json:"updateIndex"`
	UpdateIndexDescription   *string                `json:"updateIndexDescription,omitempty"`
	UpdateIndexLagging       int                    `json:"updateIndexLagging"`
	UpdatePeriodicity        *string                `json:"updatePeriodicity,omitempty"`
	UpdatePeriodicityUnit    *UpdatePeriodicityUnit `json:"updatePeriodicityUnit,omitempty"`
	Events                   *[]Event               `json:"events,omitempty"`
}

type StructureModality string

const (
	StructureModalityDefinedBenefit       StructureModality = "BENEFICIO_DEFINIDO"
	StructureModalityVariableContribution StructureModality = "CONTRIBUICAO_VARIAVEL"
)

type FinancialRegime string

const (
	FinancialRegimeCapitalization                FinancialRegime = "CAPITALIZACAO"
	FinancialRegimeRepartitionByCoverageCapitals FinancialRegime = "REPARTICAO_POR_CAPITAIS_DE_COBERTURA"
	FinancialRegimeSimpleRepartition             FinancialRegime = "REPARTICAO_SIMPLES"
)

type PricingMethod string

const (
	PricingMethodAgeRange    PricingMethod = "FAIXA_ETARIA"
	PricingMethodOthers      PricingMethod = "OUTROS"
	PricingMethodByAge       PricingMethod = "POR_IDADE"
	PricingMethodAverageRate PricingMethod = "TAXA_MEDIA"
)

type UpdateIndex string

const (
	UpdateIndexIGPDIFGV UpdateIndex = "IGP-DI-FGV"
	UpdateIndexIGPMFGV  UpdateIndex = "IGPM-FGV"
	UpdateIndexINPCIBGE UpdateIndex = "INPC-IBGE"
	UpdateIndexIPCAIBGE UpdateIndex = "IPCA-IBGE"
	UpdateIndexIPCFGV   UpdateIndex = "IPC-FGV"
	UpdateIndexTR       UpdateIndex = "TR"
	UpdateIndexOthers   UpdateIndex = "OUTROS"
)

type UpdatePeriodicityUnit string

const (
	UpdatePeriodicityUnitYear  UpdatePeriodicityUnit = "ANO"
	UpdatePeriodicityUnitDay   UpdatePeriodicityUnit = "DIA"
	UpdatePeriodicityUnitMonth UpdatePeriodicityUnit = "MES"
)

type Event struct {
	EventType       EventType `json:"eventType"`
	EventTypeOthers *string   `json:"eventTypeOthers,omitempty"`
}

type EventType string

const (
	EventTypeDisability EventType = "INVALIDEZ"
	EventTypeDeath      EventType = "MORTE"
	EventTypeOthers     EventType = "OUTROS"
)

type BiometricTable string

const (
	BiometricTableAT49M BiometricTable = "AT49_M"
	BiometricTableAT49F BiometricTable = "AT49_F"
	BiometricTableAT50M BiometricTable = "AT50_M"
	BiometricTableAT50F BiometricTable = "AT50_F"
	BiometricTableAT55M BiometricTable = "AT55_M"
	BiometricTableAT55F BiometricTable = "AT55_F"
	BiometricTableAT71M BiometricTable = "AT71_M"
	BiometricTableAT71F BiometricTable = "AT71_F"
	BiometricTableAT83M BiometricTable = "AT83_M"
	BiometricTableAT83F BiometricTable = "AT83_F"
)

type BenefitPaymentMethod string

const (
	BenefitPaymentMethodOnce   BenefitPaymentMethod = "UNICO"
	BenefitPaymentMethodIncome BenefitPaymentMethod = "RENDA"
)

type PriceIndex string

const (
	PriceIndexIPCFGV   PriceIndex = "IPC-FGV"
	PriceIndexIGPDIFGV PriceIndex = "IGP-DI-FGV"
	PriceIndexIPCAIBGE PriceIndex = "IPCA-IBGE"
	PriceIndexIGPMFGV  PriceIndex = "IGPM-FGV"
	PriceIndexINPCIBGE PriceIndex = "INPC-IBGE"
	PriceIndexTR       PriceIndex = "TR"
	PriceIndexOthers   PriceIndex = "OUTROS"
)

type Grace struct {
	GraceType              *GraceType              `json:"graceType,omitempty"`
	GracePeriod            *int                    `json:"gracePeriod,omitempty"`
	GracePeriodicity       *GracePeriodicity       `json:"gracePeriodicity,omitempty"`
	DayIndicator           *DayIndicator           `json:"dayIndicator,omitempty"`
	GracePeriodStart       *timeutil.BrazilDate    `json:"gracePeriodStart,omitempty"`
	GracePeriodEnd         *timeutil.BrazilDate    `json:"gracePeriodEnd,omitempty"`
	GracePeriodBetween     *int                    `json:"gracePeriodBetween,omitempty"`
	GracePeriodBetweenType *GracePeriodBetweenType `json:"gracePeriodBetweenType,omitempty"`
}

type GracePeriodicity string

const (
	GracePeriodicityYear  GracePeriodicity = "ANO"
	GracePeriodicityDay   GracePeriodicity = "DIA"
	GracePeriodicityMonth GracePeriodicity = "MES"
)

type GracePeriodBetweenType string

const (
	GracePeriodBetweenTypeYear  GracePeriodBetweenType = "ANO"
	GracePeriodBetweenTypeDay   GracePeriodBetweenType = "DIA"
	GracePeriodBetweenTypeMonth GracePeriodBetweenType = "MES"
)

type GraceType string

const (
	GraceTypePortability GraceType = "PORTABILIDADE"
	GraceTypeWithdrawal  GraceType = "RESGATE"
)

type DayIndicator string

const (
	DayIndicatorUsual    DayIndicator = "CORRIDOS"
	DayIndicatorBusiness DayIndicator = "UTEIS"
)

type MovementBenefit struct {
	BenefitAmount      *insurer.AmountDetails `json:"benefitAmount,omitempty"`
	BenefitPaymentDate *timeutil.BrazilDate   `json:"benefitPaymentDate,omitempty"`
}

type MovementContribution struct {
	ContributionAmount         insurer.AmountDetails `json:"contributionAmount"`
	ContributionPaymentDate    timeutil.BrazilDate   `json:"contributionPaymentDate"`
	ContributionExpirationDate timeutil.BrazilDate   `json:"contributionExpirationDate"`
	ChargedInAdvanceAmount     insurer.AmountDetails `json:"chargedInAdvanceAmount"`
	Periodicity                MovementPeriodicity   `json:"periodicity"`
	PeriodicityOthers          *string               `json:"periodicityOthers,omitempty"`
}

type MovementPeriodicity string

const (
	MovementPeriodicityMensal         MovementPeriodicity = "MENSAL"
	MovementPeriodicityBimestral      MovementPeriodicity = "BIMESTRAL"
	MovementPeriodicityTrimestral     MovementPeriodicity = "TRIMESTRAL"
	MovementPeriodicityQuadrimestral  MovementPeriodicity = "QUADRIMESTRAL"
	MovementPeriodicitySemestral      MovementPeriodicity = "SEMESTRAL"
	MovementPeriodicityAnual          MovementPeriodicity = "ANUAL" //nolint:misspell
	MovementPeriodicityEsporadica     MovementPeriodicity = "ESPORADICA"
	MovementPeriodicityPagamentoUnico MovementPeriodicity = "PAGAMENTO_UNICO"
	MovementPeriodicityOutros         MovementPeriodicity = "OUTROS"
)

type MovementType string

const (
	MovementTypeContribution MovementType = "APORTE"
	MovementTypeWithdrawal   MovementType = "RESGATE"
	MovementTypeTransfer     MovementType = "TRANSFERENCIA"
	MovementTypeIncome       MovementType = "RENDIMENTO"
	MovementTypeFee          MovementType = "TAXA"
	MovementTypeOthers       MovementType = "OUTROS"
)

type PortabilityData struct {
	Direction       PortabilityDirection  `json:"direction"`
	Type            *PortabilityType      `json:"type,omitempty"`
	Amount          insurer.AmountDetails `json:"amount"`
	RequestDate     timeutil.DateTime     `json:"requestDate"`
	LiquidationDate timeutil.DateTime     `json:"liquidationDate"`
	ChargingValue   insurer.AmountDetails `json:"chargingValue"`
	SourceEntity    *string               `json:"sourceEntity,omitempty"`
	TargetEntity    *string               `json:"targetEntity,omitempty"`
	SusepProcess    *string               `json:"susepProcess,omitempty"`
}

type PortabilityDirection string

const (
	PortabilityDirectionEntry PortabilityDirection = "ENTRADA"
	PortabilityDirectionExit  PortabilityDirection = "SAIDA"
)

type PortabilityType string

const (
	PortabilityTypeTotal   PortabilityType = "TOTAL"
	PortabilityTypePartial PortabilityType = "PARCIAL"
)

type WithdrawalData struct {
	WithdrawalOccurence bool                   `json:"withdrawalOccurence"`
	Type                *WithdrawalType        `json:"type,omitempty"`
	RequestDate         *timeutil.DateTime     `json:"requestDate,omitempty"`
	Amount              *insurer.AmountDetails `json:"amount,omitempty"`
	LiquidationDate     *timeutil.DateTime     `json:"liquidationDate,omitempty"`
	PostedChargedAmount *insurer.AmountDetails `json:"postedChargedAmount,omitempty"`
	Nature              *WithdrawalNature      `json:"nature,omitempty"`
}

type WithdrawalType string

const (
	WithdrawalTypeTotal   WithdrawalType = "TOTAL"
	WithdrawalTypePartial WithdrawalType = "PARCIAL"
)

type WithdrawalNature string

const (
	WithdrawalNatureRegularWithdrawal                 WithdrawalNature = "RESGATE_REGULAR"
	WithdrawalNatureDeath                             WithdrawalNature = "MORTE"
	WithdrawalNatureDisability                        WithdrawalNature = "INVALIDEZ"
	WithdrawalNatureScheduledFinancialPayment         WithdrawalNature = "PAGAMENTO_FINANCEIRO_PROGRAMADO"
	WithdrawalNatureRiskCoverageCostInConjugatedPlans WithdrawalNature = "CUSTEIO_DE_COBERTURA_DE_RISCO_EM_PLANOS_CONJUGADOS"
	WithdrawalNatureFinancialAssistance               WithdrawalNature = "ASSISTENCIA_FINANCEIRA"
)

type ClaimData struct {
	EventInfo  EventInfo   `json:"eventInfo"`
	IncomeInfo *IncomeInfo `json:"incomeInfo,omitempty"`
}

type EventInfo struct {
	EventAlertDate    timeutil.BrazilDate `json:"eventAlertDate"`
	EventRegisterDate timeutil.BrazilDate `json:"eventRegisterDate"`
	EventStatus       EventStatus         `json:"eventStatus"`
}

type EventStatus string

const (
	EventStatusOpen                           EventStatus = "ABERTO"
	EventStatusInitialAssessment              EventStatus = "AVALIACAO_INICIAL"
	EventStatusCancelledDueToOperationalError EventStatus = "CANCELADO_POR_ERRO_OPERACIONAL"
	EventStatusClosedWithIncomeBenefitGrant   EventStatus = "ENCERRADO_COM_CONCESSAO_DE_RENDA_BENEFICIO"
	EventStatusClosedWithCompensation         EventStatus = "ENCERRADO_COM_INDENIZACAO"
	EventStatusClosedWithSinglePaymentBenefit EventStatus = "ENCERRADO_COM_PAGAMENTO_UNICO_BENEFICIO"
	EventStatusClosedDeniedBenefit            EventStatus = "ENCERRADO_INDEFERIDO_BENEFICIO"
	EventStatusClosedWithoutCompensation      EventStatus = "ENCERRADO_SEM_INDENIZACAO"
	EventStatusReopened                       EventStatus = "REABERTO"
)

type IncomeInfo struct {
	BeneficiaryDocument      string                `json:"beneficiaryDocument"`
	BeneficiaryDocumentType  DocumentType          `json:"beneficiaryDocumentType"`
	BeneficiaryDocTypeOthers *string               `json:"beneficiaryDocTypeOthers,omitempty"`
	BeneficiaryName          string                `json:"beneficiaryName"`
	BeneficiaryCategory      BeneficiaryCategory   `json:"beneficiaryCategory"`
	BeneficiaryBirthDate     timeutil.BrazilDate   `json:"beneficiaryBirthDate"`
	IncomeType               IncomeType            `json:"incomeType"`
	IncomeTypeDetails        *string               `json:"incomeTypeDetails,omitempty"`
	ReversedIncome           *bool                 `json:"reversedIncome,omitempty"`
	IncomeAmount             insurer.AmountDetails `json:"incomeAmount"`
	PaymentTerms             *string               `json:"paymentTerms,omitempty"`
	BenefitAmount            int                   `json:"benefitAmount"`
	GrantedDate              timeutil.BrazilDate   `json:"grantedDate"`
	MonetaryUpdateIndex      MonetaryUpdateIndex   `json:"monetaryUpdateIndex"`
	MonetaryUpdIndexOthers   *string               `json:"monetaryUpdIndexOthers,omitempty"`
	LastUpdateDate           timeutil.BrazilDate   `json:"lastUpdateDate"`
}

type MonetaryUpdateIndex string

const (
	MonetaryUpdateIndexIGPDIFGV MonetaryUpdateIndex = "IGP-DI-FGV"
	MonetaryUpdateIndexIGPMFGV  MonetaryUpdateIndex = "IGPM-FGV"
	MonetaryUpdateIndexINPCIBGE MonetaryUpdateIndex = "INPC-IBGE"
	MonetaryUpdateIndexIPCAIBGE MonetaryUpdateIndex = "IPCA-IBGE"
	MonetaryUpdateIndexIPCFGV   MonetaryUpdateIndex = "IPC-FGV"
	MonetaryUpdateIndexTR       MonetaryUpdateIndex = "TR"
	MonetaryUpdateIndexOthers   MonetaryUpdateIndex = "OUTROS"
)

type BeneficiaryCategory string

const (
	BeneficiaryCategoryInsured              BeneficiaryCategory = "SEGURADO"
	BeneficiaryCategorySpouse               BeneficiaryCategory = "CONJUGE"
	BeneficiaryCategoryMinorChild           BeneficiaryCategory = "FILHO_MENOR_DE_IDADE"
	BeneficiaryCategoryIndicatedBeneficiary BeneficiaryCategory = "BENEFICIARIO_INDICADO"
)

type IncomeType string

const (
	IncomeTypeSinglePayment                                          IncomeType = "PAGAMENTO_UNICO"
	IncomeTypeIncomeForFixedTerm                                     IncomeType = "RENDA_POR_PRAZO_CERTO"
	IncomeTypeTemporaryIncome                                        IncomeType = "RENDA_TEMPORARIA"
	IncomeTypeReversibleTemporaryIncome                              IncomeType = "RENDA_TEMPORARIA_REVERSIVEL"
	IncomeTypeTemporaryIncomeWithMinimumGuaranteedTerm               IncomeType = "RENDA_TEMPORARIA_COM_PRAZO_MINIMO_GARANTIDO"
	IncomeTypeReversibleTemporaryIncomeWithMinimumGuaranteedTerm     IncomeType = "RENDA_TEMPORARIA_REVERSIVEL_COM_PRAZO_MINIMO_GARANTIDO"
	IncomeTypeLifetimeIncome                                         IncomeType = "RENDA_VITALICIA"
	IncomeTypeLifetimeIncomeReversibleToIndicatedBeneficiary         IncomeType = "RENDA_VITALICIA_REVERSIVEL_AO_BENEFICIARIO_INDICADO"
	IncomeTypeLifetimeIncomeReversibleToSpouseWithContinuityToMinors IncomeType = "RENDA_VITALICIA_REVERSIVEL_AO_CONJUGE_COM_CONTINUIDADE_AOS_MENORES"
	IncomeTypeLifetimeIncomeWithMinimumGuaranteedTerm                IncomeType = "RENDA_VITALICIA_COM_PRAZO_MINIMO_GARANTIDO"
	IncomeTypeReversibleLifetimeIncomeWithMinimumGuaranteedTerm      IncomeType = "RENDA_VITALICIA_REVERSIVEL_COM_PRAZO_MINIMO_GARANTIDO"
	IncomeTypeLifetimeIncomeReversibleToSpouse                       IncomeType = "RENDA_VITALICIA_REVERSIVEL_AO_CONJUGE"
	IncomeTypeOthers                                                 IncomeType = "OUTROS"
)
