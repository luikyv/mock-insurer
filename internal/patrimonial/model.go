package patrimonial

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
	Scope = goidc.NewScope("insurance-patrimonial")
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
	return "insurance_patrimonial_policies"
}

func (p *Policy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = strutil.Random(60)
	}
	return nil
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
	return "insurance_patrimonial_claims"
}

func (c *Claim) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
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
	return "consent_insurance_patrimonial_policies"
}

type PolicyData struct {
	ProductName                   string              `json:"policyName"`
	DocumentType                  DocumentType        `json:"documentType"`
	SusepProcessNumber            string              `json:"susepProcessNumber"`
	GroupCertificateID            string              `json:"groupCertificateID"`
	IssuanceType                  IssuanceType        `json:"issuanceType"`
	IssuanceDate                  timeutil.BrazilDate `json:"issuanceDate"`
	TermStartDate                 timeutil.BrazilDate `json:"termStartDate"`
	TermEndDate                   timeutil.BrazilDate `json:"termEndDate"`
	LeadInsurerCode               *string             `json:"leadInsurerCode,omitempty"`
	LeadInsurerPolicyID           *string             `json:"leadInsurerPolicyID,omitempty"`
	MaxLMG                        AmountDetails       `json:"maxLMG"`
	ProposalID                    string              `json:"proposalID"`
	Insureds                      []Insured           `json:"insureds"`
	Beneficiaries                 *[]Beneficiary      `json:"beneficiaries,omitempty"`
	Principals                    *[]Principal        `json:"principals,omitempty"`
	Intermediaries                *[]Intermediary     `json:"intermediaries,omitempty"`
	InsuredObjects                []InsuredObject     `json:"insuredObjects"`
	Coverages                     *[]Coverage         `json:"coverages,omitempty"`
	CoinsuranceRetainedPercentage *string             `json:"coinsuranceRetainedPercentage,omitempty"`
	Coinsurers                    *[]Coinsurer        `json:"coinsurers,omitempty"`
	BranchInfo                    *BranchInfo         `json:"branchInfo,omitempty"`
	Premium                       Premium             `json:"premium"`
}

type DocumentType string

const (
	DocumentTypeIndividualPolicy           DocumentType = "APOLICE_INDIVIDUAL"
	DocumentTypeTicket                     DocumentType = "BILHETE"
	DocumentTypeCertificate                DocumentType = "CERTIFICADO"
	DocumentTypeIndividualAutomobilePolicy DocumentType = "APOLICE_INDIVIDUAL_AUTOMOVEL"
	DocumentTypeFleetAutomobilePolicy      DocumentType = "APOLICE_FROTA_AUTOMOVEL"
	DocumentTypeAutomobileCertificate      DocumentType = "CERTIFICADO_AUTOMOVEL"
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
	BirthDate                timeutil.BrazilDate        `json:"birthDate"`
	PostCode                 string                     `json:"postCode"`
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
	Country                  *insurer.CountryCode        `json:"country,omitempty"`
	Address                  *string                     `json:"address,omitempty"`
}

type IntermediaryType string

const (
	IntermediaryTypeBroker                       IntermediaryType = "CORRETOR"
	IntermediaryTypeRepresentative               IntermediaryType = "REPRESENTANTE"
	IntermediaryTypeStipulatorEndorserInstitutor IntermediaryType = "ESTIPULANTE_AVERBADOR_INSTITUIDOR"
	IntermediaryTypeCorrespondent                IntermediaryType = "CORRESPONDENTE"
	IntermediaryTypeMicroinsuranceAgent          IntermediaryType = "AGENTE_DE_MICROSSEGUROS"
	IntermediaryTypeOthers                       IntermediaryType = "OUTROS"
)

type InsuredObject struct {
	Identification     *string                 `json:"identification,omitempty"`
	Type               InsuredObjectType       `json:"type"`
	TypeAdditionalInfo *string                 `json:"typeAdditionalInfo,omitempty"`
	Description        string                  `json:"description"`
	Amount             *AmountDetails          `json:"amount,omitempty"`
	Coverages          []InsuredObjectCoverage `json:"coverages"`
}

type InsuredObjectType string

const (
	InsuredObjectTypeContract              InsuredObjectType = "CONTRATO"
	InsuredObjectTypeAdministrativeProcess InsuredObjectType = "PROCESSO_ADMINISTRATIVO"
	InsuredObjectTypeJudicialProcess       InsuredObjectType = "PROCESSO_JUDICIAL"
	InsuredObjectTypeAutomobile            InsuredObjectType = "AUTOMOVEL"
	InsuredObjectTypeDriver                InsuredObjectType = "CONDUTOR"
	InsuredObjectTypeFleet                 InsuredObjectType = "FROTA"
	InsuredObjectTypePerson                InsuredObjectType = "PESSOA"
	InsuredObjectTypeOthers                InsuredObjectType = "OUTROS"
)

type InsuredObjectCoverage struct {
	Branch                    string                        `json:"branch"`
	Code                      CoverageCode                  `json:"code"`
	Description               *string                       `json:"description,omitempty"`
	InternalCode              *string                       `json:"internalCode,omitempty"`
	SusepProcessNumber        string                        `json:"susepProcessNumber"`
	LMI                       AmountDetails                 `json:"LMI"`
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

type CoverageCode string

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

type Coinsurer struct {
	Identification  string `json:"identification"`
	CededPercentage string `json:"cededPercentage"`
}

type BranchInfo struct {
	BasicCoverageIndex *BasicCoverageIndex     `json:"basicCoverageIndex,omitempty"`
	InsuredObjects     []SpecificInsuredObject `json:"insuredObjects"`
}

type BasicCoverageIndex struct {
	Index string `json:"index"`
}

type SpecificInsuredObject struct {
	Identification   string           `json:"identification"`
	PropertyType     *PropertyType    `json:"propertyType,omitempty"`
	StructuringType  *StructuringType `json:"structuringType,omitempty"`
	PostCode         *string          `json:"postCode,omitempty"`
	BusinessActivity *string          `json:"businessActivity,omitempty"`
}

type PropertyType string

const (
	PropertyTypeHouse                  PropertyType = "CASA"
	PropertyTypeApartment              PropertyType = "APARTAMENTO"
	PropertyTypeCondominiumResidential PropertyType = "CONDOMINIO_RESIDENCIAL"
	PropertyTypeCondominiumCommercial  PropertyType = "CONDOMINIO_COMERCIAL"
	PropertyTypeCondominiumMixed       PropertyType = "CONDOMINIO_MISTO"
)

type StructuringType string

const (
	StructuringTypeCondominiumVertical   StructuringType = "CONDOMINIO_VERTICAL"
	StructuringTypeCondominiumHorizontal StructuringType = "CONDOMINIO_HORIZONTAL"
	StructuringTypeMixed                 StructuringType = "MISTO"
)

type Coverage struct {
	Branch      string              `json:"branch"`
	Code        CoverageCode        `json:"code"`
	Description *string             `json:"description,omitempty"`
	Deductible  *CoverageDeductible `json:"deductible,omitempty"`
	POS         *CoveragePOS        `json:"POS,omitempty"`
}

type CoverageDeductible struct {
	Type                 CoverageDeductibleType        `json:"type"`
	TypeAdditionalInfo   *string                       `json:"typeAdditionalInfo,omitempty"`
	Amount               AmountDetails                 `json:"amount"`
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
	ApplicationType insurer.ValueType `json:"applicationType"`
	Description     *string           `json:"description,omitempty"`
	MinValue        *AmountDetails    `json:"minValue,omitempty"`
	MaxValue        *AmountDetails    `json:"maxValue,omitempty"`
	Percentage      *string           `json:"percentage,omitempty"`
	ValueOthers     *AmountDetails    `json:"valueOthers,omitempty"`
}

type Premium struct {
	PaymentsQuantity int               `json:"paymentsQuantity"`
	Amount           AmountDetails     `json:"amount"`
	Coverages        []PremiumCoverage `json:"coverages"`
	Payments         []Payment         `json:"payments"`
}

type PremiumCoverage struct {
	Branch        string        `json:"branch"`
	Code          CoverageCode  `json:"code"`
	Description   *string       `json:"description,omitempty"`
	PremiumAmount AmountDetails `json:"premiumAmount"`
}

type Payment struct {
	MovementDate             timeutil.BrazilDate            `json:"movementDate"`
	MovementType             insurer.PaymentMovementType    `json:"movementType"`
	MovementOrigin           *insurer.PaymentMovementOrigin `json:"movementOrigin,omitempty"`
	MovementPaymentsNumber   string                         `json:"movementPaymentsNumber"`
	Amount                   AmountDetails                  `json:"amount"`
	MaturityDate             timeutil.BrazilDate            `json:"maturityDate"`
	TellerID                 *string                        `json:"tellerId,omitempty"`
	TellerIDType             *insurer.IdentificationType    `json:"tellerIdType,omitempty"`
	TellerIDOthers           *string                        `json:"tellerIdOthers,omitempty"`
	TellerName               *string                        `json:"tellerName,omitempty"`
	FinancialInstitutionCode *string                        `json:"financialInstitutionCode,omitempty"`
	PaymentType              *insurer.PaymentType           `json:"paymentType,omitempty"`
	PaymentTypeOthers        *string                        `json:"paymentTypeOthers,omitempty"`
}

type ClaimData struct {
	Identification                 string               `json:"identification"`
	DocumentationDeliveryDate      *timeutil.BrazilDate `json:"documentationDeliveryDate,omitempty"`
	Status                         ClaimStatus          `json:"status"`
	StatusAlterationDate           timeutil.BrazilDate  `json:"statusAlterationDate"`
	OccurrenceDate                 timeutil.BrazilDate  `json:"occurrenceDate"`
	WarningDate                    timeutil.BrazilDate  `json:"warningDate"`
	ThirdPartyClaimDate            *timeutil.BrazilDate `json:"thirdPartyClaimDate,omitempty"`
	Amount                         AmountDetails        `json:"amount"`
	DenialJustification            *DenialJustification `json:"denialJustification,omitempty"`
	DenialJustificationDescription *string              `json:"denialJustificationDescription,omitempty"`
	Coverages                      []ClaimCoverage      `json:"coverages"`
}

type ClaimStatus string

const (
	ClaimStatusOpen                           ClaimStatus = "ABERTO"
	ClaimStatusClosedWithCompensation         ClaimStatus = "ENCERRADO_COM_INDENIZACAO"
	ClaimStatusClosedWithoutCompensation      ClaimStatus = "ENCERRADO_SEM_INDENIZACAO"
	ClaimStatusReopened                       ClaimStatus = "REABERTO"
	ClaimStatusCancelledDueToOperationalError ClaimStatus = "CANCELADO_POR_ERRO_OPERACIONAL"
	ClaimStatusInitialAssessment              ClaimStatus = "AVALIACAO_INICIAL"
)

type DenialJustification string

const (
	DenialJustificationDocumentationIncomplete DenialJustification = "DOCUMENTACAO_INCOMPLETA"
	DenialJustificationOutOfCoverage           DenialJustification = "FORA_COBERTURA"
	DenialJustificationOthers                  DenialJustification = "OUTROS"
	DenialJustificationPrescription            DenialJustification = "PRESCRICAO"
	DenialJustificationAggravatedRisk          DenialJustification = "RISCO_AGRAVADO"
	DenialJustificationExcludedRisk            DenialJustification = "RISCO_EXCLUIDO"
	DenialJustificationWithoutDocumentation    DenialJustification = "SEM_DOCUMENTACAO"
)

type ClaimCoverage struct {
	InsuredObjectID     *string              `json:"insuredObjectId,omitempty"`
	Branch              string               `json:"branch"`
	Code                CoverageCode         `json:"code"`
	Description         *string              `json:"description,omitempty"`
	WarningDate         *timeutil.BrazilDate `json:"warningDate,omitempty"`
	ThirdPartyClaimDate *timeutil.BrazilDate `json:"thirdPartyClaimDate,omitempty"`
}

type AmountDetails struct {
	Amount   string           `json:"amount"`
	Currency insurer.Currency `json:"currency"`
}
