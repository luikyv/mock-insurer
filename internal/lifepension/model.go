package lifepension

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
	Scope = goidc.NewScope("insurance-life-pension")
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
	return "insurance_life_pension_contracts"
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
	return "consent_insurance_life_pension_contracts"
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
	return "insurance_life_pension_portabilities"
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
	return "insurance_life_pension_withdrawals"
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
	return "insurance_life_pension_claims"
}

func (c *Claim) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type ContractData struct {
	ProductCode        string              `json:"productCode"`
	ProductName        string              `json:"productName"`
	ProposalID         string              `json:"proposalId"`
	ContractID         *string             `json:"contractId,omitempty"`
	ContractingType    ContractingType     `json:"contractingType"`
	EffectiveDateStart timeutil.BrazilDate `json:"effectiveDateStart"`
	EffectiveDateEnd   timeutil.BrazilDate `json:"effectiveDateEnd"`
	CertificateActive  bool                `json:"certificateActive"`
	ConjugatedPlan     bool                `json:"conjugatedPlan"`
	PlanType           *PlanType           `json:"planType,omitempty"`
	Periodicity        Periodicity         `json:"periodicity"`
	PeriodicityOthers  *string             `json:"periodicityOthers,omitempty"`
	TaxRegime          TaxRegime           `json:"taxRegime"`
	Insured            Insured             `json:"insured"`
	Beneficiaries      *[]Beneficiary      `json:"beneficiaries,omitempty"`
	Intermediary       *Intermediary       `json:"intermediary,omitempty"`
	Suseps             []Suseps            `json:"suseps"`
	// Create a table for this.
	MovementContributions []MovementContribution `json:"movementContributions"`
	MovementBenefits      []MovementBenefit      `json:"movementBenefits"`
}

type ContractingType string

const (
	ContractingTypeColetivo   ContractingType = "COLETIVO"
	ContractingTypeIndividual ContractingType = "INDIVIDUAL"
)

type PlanType string

const (
	PlanTypeAverbado                   PlanType = "AVERBADO"
	PlanTypeInstituidoContributario    PlanType = "INSTITUIDO_CONTRIBUTARIO"
	PlanTypeInstituidoNaoContributario PlanType = "INSTITUIDO_NAO_CONTRIBUTARIO"
)

type Periodicity string

const (
	PeriodicityMensal         Periodicity = "MENSAL"
	PeriodicityBimestral      Periodicity = "BIMESTRAL"
	PeriodicityTrimestral     Periodicity = "TRIMESTRAL"
	PeriodicityQuadrimestral  Periodicity = "QUADRIMESTRAL"
	PeriodicitySemestral      Periodicity = "SEMESTRAL"
	PeriodicityAnual          Periodicity = "ANUAL" //nolint:misspell
	PeriodicityEsporadica     Periodicity = "ESPORADICA"
	PeriodicityPagamentoUnico Periodicity = "PAGAMENTO_UNICO"
	PeriodicityOutros         Periodicity = "OUTROS"
)

type TaxRegime string

const (
	TaxRegimeProgressivo TaxRegime = "PROGRESSIVO" //nolint:misspell
	TaxRegimeRegressivo  TaxRegime = "REGRESSIVO"  //nolint:misspell
)

type ProductType string

const (
	ProductTypePGBL   ProductType = "PGBL"
	ProductTypeVGBL   ProductType = "VGBL"
	ProductTypeOthers ProductType = "OUTROS"
)

type Insured struct {
	DocumentType          DocumentType        `json:"documentType"`
	DocumentTypeOthers    *string             `json:"documentTypeOthers,omitempty"`
	DocumentNumber        string              `json:"documentNumber"`
	Name                  string              `json:"name"`
	BirthDate             timeutil.BrazilDate `json:"birthDate"`
	Gender                Gender              `json:"gender"`
	PostCode              string              `json:"postCode"`
	TownName              string              `json:"townName"`
	CountrySubDivision    string              `json:"countrySubDivision"`
	CountryCode           string              `json:"countryCode"`
	Address               string              `json:"address"`
	AddressAdditionalInfo *string             `json:"addressAdditionalInfo,omitempty"`
	Email                 *string             `json:"email,omitempty"`
}

type Gender string

const (
	GenderMasculino    Gender = "MASCULINO"
	GenderFeminino     Gender = "FEMININO"
	GenderNaoInformado Gender = "NAO_INFORMADO"
)

type DocumentType string

const (
	DocumentTypeCPF      DocumentType = "CPF"
	DocumentTypeCNPJ     DocumentType = "CNPJ"
	DocumentTypePassport DocumentType = "PASSAPORTE"
	DocumentTypeOthers   DocumentType = "OUTROS"
)

type Beneficiary struct {
	DocumentNumber          string               `json:"documentNumber"`
	DocumentType            DocumentType         `json:"documentType"`
	DocumentTypeOthers      *string              `json:"documentTypeOthers,omitempty"`
	Name                    string               `json:"name"`
	BirthDate               *timeutil.BrazilDate `json:"birthDate,omitempty"`
	Kinship                 *string              `json:"kinship,omitempty"`
	KinshipOthers           *string              `json:"kinshipOthers,omitempty"`
	ParticipationPercentage string               `json:"participationPercentage"`
}

type MovementBenefit struct {
	Amount      insurer.AmountDetails `json:"amount"`
	PaymentDate timeutil.BrazilDate   `json:"paymentDate"`
}

type MovementContribution struct {
	Amount                 insurer.AmountDetails `json:"amount"`
	PaymentDate            timeutil.BrazilDate   `json:"paymentDate"`
	ExpirationDate         timeutil.BrazilDate   `json:"expirationDate"`
	ChargedInAdvanceAmount insurer.AmountDetails `json:"chargedInAdvanceAmount"`
	Periodicity            MovementPeriodicity   `json:"periodicity"`
	PeriodicityOthers      *string               `json:"periodicityOthers,omitempty"`
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
	PortabilityDate        timeutil.BrazilDate   `json:"portabilityDate"`
	SourceInstitution      string                `json:"sourceInstitution"`
	DestinationInstitution string                `json:"destinationInstitution"`
	PortabilityAmount      insurer.AmountDetails `json:"portabilityAmount"`
	Status                 PortabilityStatus     `json:"status"`
	StatusDate             timeutil.BrazilDate   `json:"statusDate"`
	Description            *string               `json:"description,omitempty"`
	FIE                    *[]PortabilityFIE     `json:"FIE,omitempty"`
	Direction              PortabilityDirection  `json:"direction"`
	PostedChargedAmount    insurer.AmountDetails `json:"postedChargedAmount"`
	SusepProcess           *string               `json:"susepProcess,omitempty"`
	TaxRegime              *PortabilityTaxRegime `json:"taxRegime,omitempty"`
	Type                   *PortabilityType      `json:"type,omitempty"`
}

type PortabilityFIE struct {
	FIECNPJ      string                   `json:"FIECNPJ"`
	FIEName      string                   `json:"FIEName"`
	FIETradeName string                   `json:"FIETradeName"`
	PortedType   PortabilityFIEPortedType `json:"portedType"`
}

type PortabilityFIEPortedType string

const (
	PortabilityFIEPortedTypeOrigin      PortabilityFIEPortedType = "ORIGEM"
	PortabilityFIEPortedTypeDestination PortabilityFIEPortedType = "DESTINO"
)

type PortabilityDirection string

const (
	PortabilityDirectionEntrada PortabilityDirection = "ENTRADA"
	PortabilityDirectionSaida   PortabilityDirection = "SAIDA"
)

type PortabilityTaxRegime string

const (
	PortabilityTaxRegimeProgressivo PortabilityTaxRegime = "PROGRESSIVO" //nolint:misspell
	PortabilityTaxRegimeRegressivo  PortabilityTaxRegime = "REGRESSIVO"  //nolint:misspell
)

type PortabilityType string

const (
	PortabilityTypeTotal   PortabilityType = "TOTAL"
	PortabilityTypeParcial PortabilityType = "PARCIAL"
	PortabilityTypeOutros  PortabilityType = "OUTROS"
)

type PortabilityStatus string

const (
	PortabilityStatusPending    PortabilityStatus = "PENDENTE"
	PortabilityStatusInProgress PortabilityStatus = "EM_ANDAMENTO"
	PortabilityStatusCompleted  PortabilityStatus = "CONCLUIDA"
	PortabilityStatusCancelled  PortabilityStatus = "CANCELADA"
)

type WithdrawalData struct {
	WithdrawalOccurence bool                   `json:"withdrawalOccurence"`
	Type                *WithdrawalType        `json:"type,omitempty"`
	RequestDate         *timeutil.DateTime     `json:"requestDate,omitempty"`
	Amount              *insurer.AmountDetails `json:"amount,omitempty"`
	LiquidationDate     *timeutil.DateTime     `json:"liquidationDate,omitempty"`
	PostedChargedAmount *insurer.AmountDetails `json:"postedChargedAmount,omitempty"`
	Nature              *WithdrawalNature      `json:"nature,omitempty"`
	FIE                 *[]WithdrawalFIE       `json:"FIE,omitempty"`
}

type WithdrawalFIE struct {
	FIECNPJ      *string `json:"FIECNPJ,omitempty"`
	FIEName      *string `json:"FIEName,omitempty"`
	FIETradeName *string `json:"FIETradeName,omitempty"`
}

type WithdrawalNature string

const (
	WithdrawalNatureRegularWithdrawal                 WithdrawalNature = "RESGATE_REGULAR"
	WithdrawalNatureDeath                             WithdrawalNature = "MORTE"
	WithdrawalNatureDisability                        WithdrawalNature = "INVALIDEZ"
	WithdrawalNatureScheduledFinancialPayment         WithdrawalNature = "PAGAMENTO_FINANCEIRO_PROGRAMADO"
	WithdrawalNatureRiskCoverageCostInConjugatedPlans WithdrawalNature = "CUSTEIO_DE_COBERTURA_DE_RISCO_EM_PLANOS_CONJUGADOS"
	WithdrawalNatureFinancialAssistance               WithdrawalNature = "ASSISTENCIA_FINANCEIRA"
)

type WithdrawalType string

const (
	WithdrawalTypeTotal   WithdrawalType = "TOTAL"
	WithdrawalTypePartial WithdrawalType = "PARCIAL"
)

type WithdrawalStatus string

const (
	WithdrawalStatusPending   WithdrawalStatus = "PENDENTE"
	WithdrawalStatusApproved  WithdrawalStatus = "APROVADO"
	WithdrawalStatusCompleted WithdrawalStatus = "CONCLUIDO"
	WithdrawalStatusCancelled WithdrawalStatus = "CANCELADO"
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
	EventStatusClosedWithCompensation         EventStatus = "ENCERRADO_COM_INDENIZACAO"
	EventStatusClosedWithoutCompensation      EventStatus = "ENCERRADO_SEM_INDENIZACAO"
	EventStatusReopened                       EventStatus = "REABERTO"
	EventStatusCancelledDueToOperationalError EventStatus = "CANCELADO_POR_ERRO_OPERACIONAL" //nolint:misspell
	EventStatusInitialAssessment              EventStatus = "AVALIACAO_INICIAL"
	EventStatusClosedWithSinglePaymentBenefit EventStatus = "ENCERRADO_COM_PAGAMENTO_UNICO_BENEFICIO"
	EventStatusClosedWithIncomeBenefitGrant   EventStatus = "ENCERRADO_COM_CONCESSAO_DE_RENDA_BENEFICIO"
	EventStatusClosedDeniedBenefit            EventStatus = "ENCERRADO_INDEFERIDO_BENEFICIO"
)

type IncomeInfo struct {
	BeneficiaryBirthDate     timeutil.BrazilDate   `json:"beneficiaryBirthDate"`
	BeneficiaryCategory      BeneficiaryCategory   `json:"beneficiaryCategory"`
	BeneficiaryDocTypeOthers *string               `json:"beneficiaryDocTypeOthers,omitempty"`
	BeneficiaryDocument      string                `json:"beneficiaryDocument"`
	BeneficiaryDocumentType  DocumentType          `json:"beneficiaryDocumentType"`
	BeneficiaryName          string                `json:"beneficiaryName"`
	BenefitAmount            int                   `json:"benefitAmount"`
	DefermentDueDate         *timeutil.BrazilDate  `json:"defermentDueDate,omitempty"`
	GrantedDate              timeutil.BrazilDate   `json:"grantedDate"`
	IncomeAmount             insurer.AmountDetails `json:"incomeAmount"`
	IncomeType               IncomeType            `json:"incomeType"`
	IncomeTypeDetails        *string               `json:"incomeTypeDetails,omitempty"`
	LastUpdateDate           timeutil.BrazilDate   `json:"lastUpdateDate"`
	MonetaryUpdIndexOthers   *string               `json:"monetaryUpdIndexOthers,omitempty"`
	MonetaryUpdateIndex      MonetaryUpdateIndex   `json:"monetaryUpdateIndex"`
	PaymentTerms             *string               `json:"paymentTerms,omitempty"`
	ReversedIncome           *bool                 `json:"reversedIncome,omitempty"`
}

type BeneficiaryCategory string

const (
	BeneficiaryCategoryInsured              BeneficiaryCategory = "SEGURADO"
	BeneficiaryCategorySpouse               BeneficiaryCategory = "CÔNJUGE"
	BeneficiaryCategoryMinorChild           BeneficiaryCategory = "FILHO_MENOR_DE_IDADE"
	BeneficiaryCategoryIndicatedBeneficiary BeneficiaryCategory = "BENEFICIÁRIO_INDICADO"
)

type IncomeType string

const (
	IncomeTypeSinglePayment                                          IncomeType = "PAGAMENTO_UNICO"
	IncomeTypeFixedTermIncome                                        IncomeType = "RENDA_POR_PRAZO_CERTO"
	IncomeTypeTemporaryIncome                                        IncomeType = "RENDA_TEMPORARIA"
	IncomeTypeReversibleTemporaryIncome                              IncomeType = "RENDA_TEMPORARIA_REVERSIVEL"
	IncomeTypeTemporaryIncomeWithMinimumGuaranteedTerm               IncomeType = "RENDA_TEMPORARIA_COM_PRAZO_MINIMO_GARANTIDO"
	IncomeTypeReversibleTemporaryIncomeWithMinimumGuaranteedTerm     IncomeType = "RENDA_TEMPORARIA_REVERSIVEL_COM_PRAZO_MINIMO_GARANTIDO"
	IncomeTypeLifetimeIncome                                         IncomeType = "RENDA_VITALICIA"
	IncomeTypeLifetimeIncomeReversibleToIndicatedBeneficiary         IncomeType = "RENDA_VITALICIA_REVERSIVEL_AO_BENEFICIARIO_INDICADO"
	IncomeTypeLifetimeIncomeReversibleToSpouseWithContinuityToMinors IncomeType = "RENDA_VITALICIA_REVERSIVEL_AO_CONJUGE_COM_CONTINUIDADE_AOS_MENORES"
	IncomeTypeLifetimeIncomeWithMinimumGuaranteedTerm                IncomeType = "RENDA_VITALICIA_COM_PRAZO_MINIMO_GARANTIDO"
	IncomeTypeLifetimeIncomeReversibleWithMinimumGuaranteedTerm      IncomeType = "RENDA_VITALICIA_REVERSIVEL_COM_PRAZO_MINIMO_GARANTIDO"
	IncomeTypeLifetimeIncomeReversibleToSpouse                       IncomeType = "RENDA_VITALICIA_REVERSIVEL_AO_CONJUGE"
	IncomeTypeOthers                                                 IncomeType = "OUTROS"
)

type MonetaryUpdateIndex string

const (
	MonetaryUpdateIndexIPCA   MonetaryUpdateIndex = "IPCA"
	MonetaryUpdateIndexIGPM   MonetaryUpdateIndex = "IGPM"
	MonetaryUpdateIndexINPC   MonetaryUpdateIndex = "INPC"
	MonetaryUpdateIndexTR     MonetaryUpdateIndex = "TR"
	MonetaryUpdateIndexOutros MonetaryUpdateIndex = "OUTROS"
)

type Intermediary struct {
	Type               IntermediaryType `json:"type"`
	TypeOthers         *string          `json:"typeOthers,omitempty"`
	DocumentNumber     *string          `json:"documentNumber,omitempty"`
	IntermediaryID     *string          `json:"intermediaryId,omitempty"`
	DocumentType       *DocumentType    `json:"documentType,omitempty"`
	DocumentTypeOthers *string          `json:"documentTypeOthers,omitempty"`
	Name               *string          `json:"name,omitempty"`
	PostCode           *string          `json:"postCode,omitempty"`
	TownName           *string          `json:"townName,omitempty"`
	CountrySubDivision *string          `json:"countrySubDivision,omitempty"`
	CountryCode        *string          `json:"countryCode,omitempty"`
	Address            *string          `json:"address,omitempty"`
	AdditionalInfo     *string          `json:"additionalInfo,omitempty"`
}

type IntermediaryType string

const (
	IntermediaryTypeCorretor                        IntermediaryType = "CORRETOR"
	IntermediaryTypeRepresentante                   IntermediaryType = "REPRESENTANTE"
	IntermediaryTypeEstipulanteAverbadorInstituidor IntermediaryType = "ESTIPULANTE_AVERBADOR_INSTITUIDOR"
	IntermediaryTypeCorrespondente                  IntermediaryType = "CORRESPONDENTE" //nolint:misspell
	IntermediaryTypeAgenteDeMicrosseguros           IntermediaryType = "AGENTE_DE_MICROSSEGUROS"
	IntermediaryTypeOutros                          IntermediaryType = "OUTROS"
)

type Suseps struct {
	CoverageCode                       string                 `json:"coverageCode"`
	SusepProcessNumber                 string                 `json:"susepProcessNumber"`
	StructureModality                  StructureModality      `json:"structureModality"`
	Type                               SusepsType             `json:"type"`
	TypeDetails                        *string                `json:"typeDetails,omitempty"`
	LockedPlan                         bool                   `json:"lockedPlan"`
	QualifiedProposer                  bool                   `json:"qualifiedProposer"`
	BenefitPaymentMethod               BenefitPaymentMethod   `json:"benefitPaymentMethod"`
	FinancialResultReversal            bool                   `json:"financialResultReversal"`
	FinancialResultReversalPercentage  *string                `json:"financialResultReversalPercentage,omitempty"`
	CalculationBasis                   CalculationBasis       `json:"calculationBasis"`
	FIE                                []FIE                  `json:"FIE"`
	BenefitAmount                      *insurer.AmountDetails `json:"benefitAmount,omitempty"`
	RentsInterestRate                  *string                `json:"rentsInterestRate,omitempty"`
	Grace                              *[]Grace               `json:"grace,omitempty"`
	BiometricTable                     *string                `json:"biometricTable,omitempty"`
	PmbacInterestRate                  *string                `json:"pmbacInterestRate,omitempty"`
	PmbacGuaranteePriceIndex           *string                `json:"pmbacGuaranteePriceIndex,omitempty"`
	PmbacGuaranteePriceOthers          *string                `json:"pmbacGuaranteePriceOthers,omitempty"`
	PmbacIndexLagging                  *int                   `json:"pmbacIndexLagging,omitempty"`
	PdrOrVdrminimalGuaranteeIndex      *string                `json:"pdrOrVdrminimalGuaranteeIndex,omitempty"`
	PdrOrVdrminimalGuaranteeOthers     *string                `json:"pdrOrVdrminimalGuaranteeOthers,omitempty"`
	PdrOrVdrminimalGuaranteePercentage *string                `json:"pdrOrVdrminimalGuaranteePercentage,omitempty"`
}

type StructureModality string

const (
	StructureModalityBeneficioDefinido    StructureModality = "BENEFICIO_DEFINIDO"
	StructureModalityContribuicaoVariavel StructureModality = "CONTRIBUICAO_VARIAVEL"
)

type SusepsType string

const (
	SusepsTypePGBL        SusepsType = "PGBL"
	SusepsTypePRGP        SusepsType = "PRGP"
	SusepsTypePAGP        SusepsType = "PAGP"
	SusepsTypePRSA        SusepsType = "PRSA"
	SusepsTypePRI         SusepsType = "PRI"
	SusepsTypePDR         SusepsType = "PDR"
	SusepsTypeVGBL        SusepsType = "VGBL"
	SusepsTypeVRGP        SusepsType = "VRGP"
	SusepsTypeVAGP        SusepsType = "VAGP"
	SusepsTypeVRSA        SusepsType = "VRSA"
	SusepsTypeVRI         SusepsType = "VRI"
	SusepsTypeVDR         SusepsType = "VDR"
	SusepsTypeTradicional SusepsType = "TRADICIONAL" //nolint:misspell
	SusepsTypeOutros      SusepsType = "OUTROS"
)

type BenefitPaymentMethod string

const (
	BenefitPaymentMethodUnico BenefitPaymentMethod = "UNICO"
	BenefitPaymentMethodRenda BenefitPaymentMethod = "RENDA"
)

type CalculationBasis string

const (
	CalculationBasisMensal CalculationBasis = "MENSAL"
	CalculationBasisAnual  CalculationBasis = "ANUAL" //nolint:misspell
)

type FIE struct {
	FIECNPJ                string                `json:"FIECNPJ"`
	FIEName                string                `json:"FIEName"`
	FIETradeName           string                `json:"FIETradeName"`
	PmbacAmount            insurer.AmountDetails `json:"pmbacAmount"`
	ProvisionSurplusAmount insurer.AmountDetails `json:"provisionSurplusAmount"`
}

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

type GraceType string

const (
	GraceTypePortabilidade GraceType = "PORTABILIDADE"
	GraceTypeResgate       GraceType = "RESGATE"
)

type GracePeriodicity string

const (
	GracePeriodicityDia GracePeriodicity = "DIA"
	GracePeriodicityMes GracePeriodicity = "MES"
	GracePeriodicityAno GracePeriodicity = "ANO"
)

type DayIndicator string

const (
	DayIndicatorUteis    DayIndicator = "UTEIS"
	DayIndicatorCorridos DayIndicator = "CORRIDOS"
)

type GracePeriodBetweenType string

const (
	GracePeriodBetweenTypeDia GracePeriodBetweenType = "DIA"
	GracePeriodBetweenTypeMes GracePeriodBetweenType = "MES"
	GracePeriodBetweenTypeAno GracePeriodBetweenType = "ANO"
)
