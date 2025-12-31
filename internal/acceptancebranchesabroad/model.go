package acceptancebranchesabroad

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
	Scope = goidc.NewScope("insurance-acceptance-and-branches-abroad")
)

type Policy struct {
	ID        string `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      PolicyData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (Policy) TableName() string {
	return "insurance_acceptance_and_branches_abroad_policies"
}

func (p *Policy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = strutil.Random(60)
	}
	return nil
}

type ConsentPolicy struct {
	ConsentID uuid.UUID
	PolicyID  string
	OwnerID   uuid.UUID
	Status    resource.Status
	Policy    *Policy
	OrgID     string
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (ConsentPolicy) TableName() string {
	return "consent_insurance_acceptance_and_branches_abroad_policies"
}

type Claim struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	PolicyID  string
	Data      ClaimData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (Claim) TableName() string {
	return "insurance_acceptance_and_branches_abroad_claims"
}

func (c *Claim) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type PolicyData struct {
	ProductName                   string                `json:"productName"`
	DocumentType                  DocumentType          `json:"documentType"`
	SusepProcessNumber            *string               `json:"susepProcessNumber,omitempty"`
	GroupCertificateID            *string               `json:"groupCertificateID,omitempty"`
	IssuanceType                  IssuanceType          `json:"issuanceType"`
	IssuanceDate                  timeutil.BrazilDate   `json:"issuanceDate"`
	TermStartDate                 timeutil.BrazilDate   `json:"termStartDate"`
	TermEndDate                   timeutil.BrazilDate   `json:"termEndDate"`
	LeadInsurerCode               *string               `json:"leadInsurerCode,omitempty"`
	LeadInsurerPolicyID           *string               `json:"leadInsurerPolicyID,omitempty"`
	MaxLMG                        insurer.AmountDetails `json:"maxLMG"`
	ProposalID                    string                `json:"proposalID"`
	Insureds                      []Insured             `json:"insureds"`
	Beneficiaries                 *[]Beneficiary        `json:"beneficiaries,omitempty"`
	Principals                    *[]Principal          `json:"principals,omitempty"`
	Intermediaries                *[]Intermediary       `json:"intermediaries,omitempty"`
	InsuredObjects                []InsuredObject       `json:"insuredObjects"`
	Coverages                     *[]Coverage           `json:"coverages,omitempty"`
	CoinsuranceRetainedPercentage *string               `json:"coinsuranceRetainedPercentage,omitempty"`
	Coinsurers                    *[]Coinsurer          `json:"coinsurers,omitempty"`
	BranchInfo                    BranchInfo            `json:"branchInfo"`
	Premium                       Premium               `json:"premium"`
}

type ClaimData struct {
	Identification                 string                `json:"identification"`
	DocumentDeliveryDate           *timeutil.BrazilDate  `json:"documentDeliveryDate,omitempty"`
	Status                         ClaimStatus           `json:"status"`
	StatusAlterationDate           *timeutil.BrazilDate  `json:"statusAlterationDate,omitempty"`
	OccurrenceDate                 timeutil.BrazilDate   `json:"occurrenceDate"`
	WarningDate                    timeutil.BrazilDate   `json:"warningDate"`
	ThirdPartyClaimDate            *timeutil.BrazilDate  `json:"thirdPartyClaimDate,omitempty"`
	Amount                         insurer.AmountDetails `json:"amount"`
	DenialJustification            *DenialJustification  `json:"denialJustification,omitempty"`
	DenialJustificationDescription *string               `json:"denialJustificationDescription,omitempty"`
	Coverages                      []ClaimCoverage       `json:"coverages"`
}

type DocumentType string

const (
	DocumentTypeIndividualPolicy     DocumentType = "APOLICE_INDIVIDUAL"
	DocumentTypeTicket               DocumentType = "BILHETE"
	DocumentTypeCertificate          DocumentType = "CERTIFICADO"
	DocumentTypeIndividualAutoPolicy DocumentType = "APOLICE_INDIVIDUAL_AUTOMOVEL"
	DocumentTypeFleetAutoPolicy      DocumentType = "APOLICE_FROTA_AUTOMOVEL"
	DocumentTypeAutoCertificate      DocumentType = "CERTIFICADO_AUTOMOVEL"
)

type IssuanceType string

const (
	IssuanceTypeOwn      IssuanceType = "EMISSAO_PROPRIA"
	IssuanceTypeAccepted IssuanceType = "COSSEGURO_ACEITO"
)

type Insured struct {
	Identification           string                     `json:"identification"`
	IdentificationType       insurer.IdentificationType `json:"identificationType"`
	IdentificationTypeOthers *string                    `json:"identificationTypeOthers,omitempty"`
	Name                     string                     `json:"name"`
	PostCode                 string                     `json:"postCode"`
	BirthDate                timeutil.BrazilDate        `json:"birthDate"`
	Email                    *string                    `json:"email,omitempty"`
	City                     string                     `json:"city"`
	State                    string                     `json:"state"`
	Country                  insurer.CountryCode        `json:"country"`
	Address                  string                     `json:"address"`
	AddressAdditionalInfo    *string                    `json:"addressAdditionalInfo,omitempty"`
}

type Beneficiary struct {
	Identification           string                     `json:"identification"`
	IdentificationType       insurer.IdentificationType `json:"identificationType"`
	IdentificationTypeOthers *string                    `json:"identificationTypeOthers,omitempty"`
	Name                     string                     `json:"name"`
}

type Principal struct {
	Identification           string                     `json:"identification"`
	IdentificationType       insurer.IdentificationType `json:"identificationType"`
	IdentificationTypeOthers *string                    `json:"identificationTypeOthers,omitempty"`
	Name                     string                     `json:"name"`
	PostCode                 string                     `json:"postCode"`
	Email                    *string                    `json:"email,omitempty"`
	City                     string                     `json:"city"`
	State                    string                     `json:"state"`
	Country                  insurer.CountryCode        `json:"country"`
	Address                  string                     `json:"address"`
	AddressAdditionalInfo    *string                    `json:"addressAdditionalInfo,omitempty"`
}

type Intermediary struct {
	Type                     IntermediaryType            `json:"type"`
	TypeOthers               *string                     `json:"typeOthers,omitempty"`
	Identification           *string                     `json:"identification,omitempty"`
	BrokerID                 *string                     `json:"brokerId,omitempty"`
	IdentificationType       *insurer.IdentificationType `json:"identificationType,omitempty"`
	IdentificationTypeOthers *string                     `json:"identificationTypeOthers,omitempty"`
	Name                     string                      `json:"name"`
	PostCode                 *string                     `json:"postCode,omitempty"`
	City                     *string                     `json:"city,omitempty"`
	State                    *string                     `json:"state,omitempty"`
	Country                  *string                     `json:"country,omitempty"`
	Address                  *string                     `json:"address,omitempty"`
}

type IntermediaryType string

const (
	IntermediaryTypeBroker                       IntermediaryType = "CORRETOR"
	IntermediaryTypeRepresentative               IntermediaryType = "REPRESENTANTE"
	IntermediaryTypeStipulatorEndorserInstitutor IntermediaryType = "ESTIPULANTE_AVERBADOR_INSTITUIDOR"
	IntermediaryTypeCorrespondent                IntermediaryType = "CORRESPONDENTE" //nolint:misspell
	IntermediaryTypeMicroinsuranceAgent          IntermediaryType = "AGENTE_DE_MICROSSEGUROS"
	IntermediaryTypeOthers                       IntermediaryType = "OUTROS"
)

type InsuredObject struct {
	Identification     string                  `json:"identification"`
	Type               InsuredObjectType       `json:"type"`
	TypeAdditionalInfo *string                 `json:"typeAdditionalInfo,omitempty"`
	Description        string                  `json:"description"`
	Amount             *insurer.AmountDetails  `json:"amount,omitempty"`
	Coverages          []InsuredObjectCoverage `json:"coverages"`
}

type InsuredObjectType string

const (
	InsuredObjectTypeContract              InsuredObjectType = "CONTRATO"
	InsuredObjectTypeAdministrativeProcess InsuredObjectType = "PROCESSO_ADMINISTRATIVO" //nolint:misspell
	InsuredObjectTypeJudicialProcess       InsuredObjectType = "PROCESSO_JUDICIAL"
	InsuredObjectTypeAutomobile            InsuredObjectType = "AUTOMOVEL"
	InsuredObjectTypeDriver                InsuredObjectType = "CONDUTOR"
	InsuredObjectTypeFleet                 InsuredObjectType = "FROTA"
	InsuredObjectTypePerson                InsuredObjectType = "PESSOA"
	InsuredObjectTypeOthers                InsuredObjectType = "OUTROS"
)

type InsuredObjectCoverage struct {
	Branch                    string                        `json:"branch"`
	Code                      string                        `json:"code"`
	Description               *string                       `json:"description,omitempty"`
	InternalCode              *string                       `json:"internalCode,omitempty"`
	SusepProcessNumber        string                        `json:"susepProcessNumber"`
	LMI                       insurer.AmountDetails         `json:"LMI,omitempty"`
	TermStartDate             timeutil.BrazilDate           `json:"termStartDate"`
	TermEndDate               timeutil.BrazilDate           `json:"termEndDate"`
	IsMainCoverage            *bool                         `json:"isMainCoverage,omitempty"`
	Feature                   CoverageFeature               `json:"feature"`
	Type                      CoverageType                  `json:"type"`
	GracePeriod               *int                          `json:"gracePeriod,omitempty"`
	GracePeriodicity          *insurer.Periodicity          `json:"gracePeriodicity,omitempty"`
	GracePeriodCountingMethod *insurer.PeriodCountingMethod `json:"gracePeriodCountingMethod,omitempty"`
	GracePeriodStartDate      *timeutil.BrazilDate          `json:"gracePeriodStartDate,omitempty"`
	GracePeriodEndDate        *timeutil.BrazilDate          `json:"gracePeriodEndDate,omitempty"`
	PremiumPeriodicity        insurer.PremiumPeriodicity    `json:"premiumPeriodicity"`
	PremiumPeriodicityOthers  *string                       `json:"premiumPeriodicityOthers,omitempty"`
}

type CoverageFeature string

const (
	CoverageFeatureMass               CoverageFeature = "MASSIFICADOS"
	CoverageFeatureMassMicroinsurance CoverageFeature = "MASSIFICADOS_MICROSEGUROS"
	CoverageFeatureLargeRisks         CoverageFeature = "GRANDES_RISCOS"
)

type CoverageType string

const (
	CoverageTypeParametric                CoverageType = "PARAMETRICO"
	CoverageTypeIntermittent              CoverageType = "INTERMITENTE"
	CoverageTypeRegularCommon             CoverageType = "REGULAR_COMUM"
	CoverageTypeCapitalGlobal             CoverageType = "CAPITAL_GLOBAL"
	CoverageTypeParametricAndIntermittent CoverageType = "PARAMETRICO_E_INTERMITENTE"
)

type Coverage struct {
	Branch      string              `json:"branch"`
	Code        string              `json:"code"`
	Description *string             `json:"description,omitempty"`
	Deductible  *CoverageDeductible `json:"deductible,omitempty"`
	POS         *CoveragePOS        `json:"POS,omitempty"`
}

type CoverageDeductible struct {
	Type                 CoverageDeductibleType        `json:"type"`
	TypeAdditionalInfo   *string                       `json:"typeAdditionalInfo,omitempty"`
	Amount               insurer.AmountDetails         `json:"amount"`
	Period               int                           `json:"period"`
	Periodicity          insurer.Periodicity           `json:"periodicity"`
	PeriodCountingMethod *insurer.PeriodCountingMethod `json:"periodCountingMethod,omitempty"`
	PeriodStartDate      timeutil.BrazilDate           `json:"periodStartDate"`
	PeriodEndDate        timeutil.BrazilDate           `json:"periodEndDate"`
	Description          string                        `json:"description"`
}

type CoverageDeductibleType string

const (
	CoverageDeductibleTypeReduced    CoverageDeductibleType = "REDUZIDA"
	CoverageDeductibleTypeNormal     CoverageDeductibleType = "NORMAL"
	CoverageDeductibleTypeIncreased  CoverageDeductibleType = "MAJORADA"
	CoverageDeductibleTypeDeductible CoverageDeductibleType = "DEDUTIVEL"
	CoverageDeductibleTypeOthers     CoverageDeductibleType = "OUTROS"
)

type CoveragePOS struct {
	ApplicationType insurer.ValueType      `json:"applicationType"`
	Description     string                 `json:"description"`
	MinValue        *insurer.AmountDetails `json:"minValue,omitempty"`
	MaxValue        *insurer.AmountDetails `json:"maxValue,omitempty"`
	Percentage      *insurer.AmountDetails `json:"percentage,omitempty"`
	ValueOthers     *insurer.AmountDetails `json:"valueOthers,omitempty"`
}

type Coinsurer struct {
	Identification  *string `json:"identification,omitempty"`
	CededPercentage *string `json:"cededPercentage,omitempty"`
}

type BranchInfo struct {
	RiskCountry      insurer.CountryCode `json:"riskCountry"`
	HasForum         bool                `json:"hasForum"`
	ForumDescription *string             `json:"forumDescription,omitempty"`
	TransferorID     string              `json:"transferorId"`
	TransferorName   string              `json:"transferorName"`
	GroupBranches    []string            `json:"groupBranches"`
}

type Premium struct {
	PaymentsQuantity string                `json:"paymentsQuantity"`
	Amount           insurer.AmountDetails `json:"amount"`
	Coverages        []PremiumCoverage     `json:"coverages"`
	Payments         []Payment             `json:"payments"`
}

type PremiumCoverage struct {
	Branch        string                `json:"branch"`
	Code          string                `json:"code"`
	Description   *string               `json:"description,omitempty"`
	PremiumAmount insurer.AmountDetails `json:"premiumAmount"`
}

type DenialJustification string

const (
	DenialJustificationExcludedRisk            DenialJustification = "RISCO_EXCLUIDO"
	DenialJustificationAggravatedRisk          DenialJustification = "RISCO_AGRAVADO"
	DenialJustificationWithoutDocumentation    DenialJustification = "SEM_DOCUMENTACAO"
	DenialJustificationIncompleteDocumentation DenialJustification = "DOCUMENTACAO_INCOMPLETA"
	DenialJustificationPrescription            DenialJustification = "PRESCRICAO"
	DenialJustificationOutOfCoverage           DenialJustification = "FORA_COBERTURA"
	DenialJustificationOthers                  DenialJustification = "OUTROS"
)

type ClaimStatus string

const (
	ClaimStatusOpen                       ClaimStatus = "ABERTO"
	ClaimStatusClosedWithCompensation     ClaimStatus = "ENCERRADO_COM_INDENIZACAO"
	ClaimStatusClosedWithoutCompensation  ClaimStatus = "ENCERRADO_SEM_INDENIZACAO"
	ClaimStatusReopened                   ClaimStatus = "REABERTO"
	ClaimStatusCanceledByOperationalError ClaimStatus = "CANCELADO_POR_ERRO_OPERACIONAL" //nolint:misspell
	ClaimStatusInitialEvaluation          ClaimStatus = "AVALIACAO_INICIAL"
)

type ClaimCoverage struct {
	InsuredObjectID     *string              `json:"insuredObjectId,omitempty"`
	Branch              string               `json:"branch"`
	Code                string               `json:"code"`
	Description         *string              `json:"description,omitempty"`
	WarningDate         *timeutil.BrazilDate `json:"warningDate,omitempty"`
	ThirdPartyClaimDate *timeutil.BrazilDate `json:"thirdPartyClaimDate,omitempty"`
}

type Payment struct {
	MovementDate             timeutil.BrazilDate            `json:"movementDate"`
	MovementType             insurer.PaymentMovementType    `json:"movementType"`
	MovementOrigin           *insurer.PaymentMovementOrigin `json:"movementOrigin,omitempty"`
	MovementPaymentsNumber   string                         `json:"movementPaymentsNumber"`
	Amount                   insurer.AmountDetails          `json:"amount"`
	MaturityDate             timeutil.BrazilDate            `json:"maturityDate"`
	TellerID                 *string                        `json:"tellerId,omitempty"`
	TellerIDType             *insurer.IdentificationType    `json:"tellerIdType,omitempty"`
	TellerIDOthers           *string                        `json:"tellerIdOthers,omitempty"`
	TellerName               *string                        `json:"tellerName,omitempty"`
	FinancialInstitutionCode *string                        `json:"financialInstitutionCode,omitempty"`
	PaymentType              *insurer.PaymentType           `json:"paymentType,omitempty"`
	PaymentTypeOthers        *string                        `json:"paymentTypeOthers,omitempty"`
}
