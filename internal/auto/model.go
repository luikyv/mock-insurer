package auto

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
	Scope = goidc.NewScope("insurance-auto")
)

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
	return "consent_insurance_auto_policies"
}

type Policy struct {
	ID        string     `gorm:"primaryKey"`
	Data      PolicyData `gorm:"serializer:json"`
	OwnerID   uuid.UUID
	CrossOrg  bool
	OrgID     string
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (Policy) TableName() string {
	return "insurance_auto_policies"
}

func (p *Policy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = strutil.Random(60)
	}
	return nil
}

type PolicyData struct {
	ProductName                   string                      `json:"productName"`
	DocumentType                  DocumentType                `json:"documentType"`
	SusepProcessNumber            *string                     `json:"susepProcessNumber,omitempty"`
	GroupCertificateID            *string                     `json:"groupCertificateID,omitempty"`
	IssuanceType                  IssuanceType                `json:"issuanceType"`
	IssuanceDate                  timeutil.BrazilDate         `json:"issuanceDate"`
	TermStartDate                 timeutil.BrazilDate         `json:"termStartDate"`
	TermEndDate                   timeutil.BrazilDate         `json:"termEndDate"`
	LeadInsurerCode               *string                     `json:"leadInsurerCode,omitempty"`
	LeadInsurerPolicyID           *string                     `json:"leadInsurerPolicyID,omitempty"`
	MaxLMG                        insurer.AmountDetails       `json:"maxLMG"`
	ProposalID                    string                      `json:"proposalID"`
	Insureds                      []Insured                   `json:"insureds"`
	Beneficiaries                 *[]Beneficiary              `json:"beneficiaries,omitempty"`
	Principals                    *[]Principal                `json:"principals,omitempty"`
	Intermediaries                *[]Intermediary             `json:"intermediaries,omitempty"`
	InsuredObjects                []InsuredObject             `json:"insuredObjects"`
	Coverages                     []Coverage                  `json:"coverages"`
	CoinsuranceRetainedPercentage *string                     `json:"coinsuranceRetainedPercentage,omitempty"`
	Coinurers                     *[]Coinsurer                `json:"coinurers,omitempty"`
	RepairNetwork                 RepairNetwork               `json:"repairNetwork,omitempty"`
	RepairNetworkOthers           *string                     `json:"repairNetworkOthers,omitempty"`
	RepairedPartsUsageType        RepairedPartsUsageType      `json:"repairedPartsUsageType"`
	RepairedPartsClassification   RepairedPartsClassification `json:"repairedPartsClassification"`
	RepairedPartsNationality      RepairedPartsNationality    `json:"repairedPartsNationality"`
	ValidityType                  insurer.ValidityType        `json:"validityType"`
	ValidateTypeOthers            *string                     `json:"validateTypeOthers,omitempty"`
	OtherCompensations            *string                     `json:"otherCompensations,omitempty"`
	OtherBenefits                 *OtherBenefits              `json:"otherBenefits,omitempty"`
	AssistancePackages            *AssistancePackages         `json:"assistancePackages,omitempty"`
	IsExpiredRiskPolicy           *bool                       `json:"isExpiredRiskPolicy,omitempty"`
	BonusDiscountRate             *string                     `json:"bonusDiscountRate,omitempty"`
	BonusClass                    *string                     `json:"bonusClass,omitempty"`
	Drivers                       *[]Driver                   `json:"drivers,omitempty"`
	Premium                       PremiumData                 `json:"premium"`
}

type PremiumData struct {
	PaymentsQuantity string                `json:"paymentsQuantity"`
	Amount           insurer.AmountDetails `json:"amount"`
	Coverages        []PremiumCoverage     `json:"coverages"`
	Payments         []Payment             `json:"payments"`
}

type PremiumCoverage struct {
	Branch        string                `json:"branch"`
	Code          CoverageCode          `json:"code"`
	Description   *string               `json:"description,omitempty"`
	PremiumAmount insurer.AmountDetails `json:"premiumAmount"`
}

type Payment struct {
	MovementDate             timeutil.BrazilDate         `json:"movementDate"`
	MovementType             PaymentMovementType         `json:"movementType"`
	MovementOrigin           *PaymentMovementOrigin      `json:"movementOrigin,omitempty"`
	MovementPaymentsNumber   int                         `json:"movementPaymentsNumber"`
	Amount                   insurer.AmountDetails       `json:"amount"`
	MaturityDate             timeutil.BrazilDate         `json:"maturityDate"`
	TellerID                 *string                     `json:"tellerId,omitempty"`
	TellerIDType             *insurer.IdentificationType `json:"tellerIdType,omitempty"`
	TellerIDOthers           *string                     `json:"tellerIdOthers,omitempty"`
	TellerName               *string                     `json:"tellerName,omitempty"`
	FinancialInstitutionCode *string                     `json:"financialInstitutionCode,omitempty"`
	PaymentType              *PaymentType                `json:"paymentType,omitempty"`
	PaymentTypeOthers        *string                     `json:"paymentTypeOthers,omitempty"`
}

type PaymentMovementType string

const (
	PaymentMovementTypePremiumLiquidation                                  PaymentMovementType = "LIQUIDACAO_DE_PREMIO"
	PaymentMovementTypePremiumRefundLiquidation                            PaymentMovementType = "LIQUIDACAO_DE_RESTITUICAO_DE_PREMIO"
	PaymentMovementTypeAcquisitionCostLiquidation                          PaymentMovementType = "LIQUIDACAO_DE_CUSTO_DE_AQUISICAO"
	PaymentMovementTypeAcquisitionCostRefundLiquidation                    PaymentMovementType = "LIQUIDACAO_DE_RESTITUICAO_DE_CUSTO_DE_AQUISICAO"
	PaymentMovementTypePremiumReversal                                     PaymentMovementType = "ESTORNO_DE_PREMIO"
	PaymentMovementTypePremiumRefundReversal                               PaymentMovementType = "ESTORNO_DE_RESTITUICAO_DE_PREMIO"
	PaymentMovementTypeAcquisitionCostReversal                             PaymentMovementType = "ESTORNO_DE_CUSTO_DE_AQUISICAO"
	PaymentMovementTypePremiumIssuanceWithoutEndorsement                   PaymentMovementType = "EMISSAO_DE_PREMIO_SEM_ENDOSSO"
	PaymentMovementTypeInstallmentCancellation                             PaymentMovementType = "CANCELAMENTO_DE_PARCELA"
	PaymentMovementTypePremiumRefundIssuanceWithoutEndorsement             PaymentMovementType = "EMISSAO_DE_RESTITUICAO_DE_PREMIO_SEM_ENDOSSO"
	PaymentMovementTypeInstallmentReopening                                PaymentMovementType = "REABERTURA_DE_PARCELA"
	PaymentMovementTypeWriteOffByLoss                                      PaymentMovementType = "BAIXA_POR_PERDA"
	PaymentMovementTypePremiumAndInstallmentCancellationWithoutEndorsement PaymentMovementType = "CANCELAMENTO_DE_PREMIO_E_PARCELA_SEM_ENDOSSO"
	PaymentMovementTypeFinancialCompensation                               PaymentMovementType = "COMPENSACAO_FINANCEIRA"
)

type PaymentMovementOrigin string

const (
	PaymentMovementOriginDirectIssuance              PaymentMovementOrigin = "EMISSAO_DIRETA"
	PaymentMovementOriginAcceptedCoinsuranceIssuance PaymentMovementOrigin = "EMISSAO_ACEITA_DE_COSSEGURO"
	PaymentMovementOriginCededCoinsuranceIssuance    PaymentMovementOrigin = "EMISSAO_CEDIDA_DE_COSSEGURO"
)

type PaymentType string

const (
	PaymentTypeBankSlip        PaymentType = "BOLETO"
	PaymentTypeTED             PaymentType = "TED"
	PaymentTypeTEF             PaymentType = "TEF"
	PaymentTypeCreditCard      PaymentType = "CARTAO"
	PaymentTypeDocument        PaymentType = "DOC"
	PaymentTypeCheque          PaymentType = "CHEQUE"
	PaymentTypeDiscountOnSheet PaymentType = "DESCONTO_EM_FOLHA"
	PaymentTypePix             PaymentType = "PIX"
	PaymentTypeCash            PaymentType = "DINHEIRO_EM_ESPECIE"
)

type Claim struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Data      ClaimData `gorm:"serializer:json"`
	OwnerID   uuid.UUID
	PolicyID  string
	CrossOrg  bool
	OrgID     string
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (Claim) TableName() string {
	return "insurance_auto_policy_claims"
}

func (c *Claim) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type ClaimData struct {
	Identification                 string                `json:"identification"`
	DocumentationDeliveryDate      *timeutil.BrazilDate  `json:"documentationDeliveryDate,omitempty"`
	Status                         ClaimStatus           `json:"status"`
	StatusAlterationDate           timeutil.BrazilDate   `json:"statusAlterationDate"`
	OccurrenceDate                 timeutil.BrazilDate   `json:"occurrenceDate"`
	WarningDate                    timeutil.BrazilDate   `json:"warningDate"`
	ThirdPartyClaimDate            *timeutil.BrazilDate  `json:"thirdPartyClaimDate,omitempty"`
	Amount                         insurer.AmountDetails `json:"amount"`
	DenialJustification            *DenialJustification  `json:"denialJustification,omitempty"`
	DenialJustificationDescription *string               `json:"denialJustificationDescription,omitempty"`
	Coverages                      []ClaimCoverage       `json:"coverages"`
	BranchInfo                     *CoverageBranchInfo   `json:"branchInfo,omitempty"`
}

type DocumentType string

const (
	DocumentTypeIndividual      DocumentType = "APOLICE_INDIVIDUAL"
	DocumentTypeTicket          DocumentType = "BILHETE"
	DocumentTypeCertificate     DocumentType = "CERTIFICADO"
	DocumentTypeIndividualAuto  DocumentType = "APOLICE_INDIVIDUAL_AUTOMOVEL"
	DocumentTypeFleetAuto       DocumentType = "APOLICE_FROTA_AUTOMOVEL"
	DocumentTypeCertificateAuto DocumentType = "CERTIFICADO_AUTOMOVEL"
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
	Country                  string                     `json:"country"`
	Address                  string                     `json:"address"`
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
	PostCode                 string                     `json:"postCode,omitempty"`
	Email                    *string                    `json:"email,omitempty"`
	City                     string                     `json:"city"`
	State                    string                     `json:"state"`
	Country                  string                     `json:"country"`
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
	IntermediaryTypeCorrespondent                IntermediaryType = "CORRESPONDENTE"
	IntermediaryTypeMicroinsuranceAgent          IntermediaryType = "AGENTE_DE_MICROSSEGUROS"
	IntermediaryTypeOthers                       IntermediaryType = "OUTROS"
)

type InsuredObject struct {
	Identification                   string                          `json:"identification"`
	IdentificationType               InsuredObjectIdentificationType `json:"identificationType"`
	IdentificationTypeAdditionalInfo *string                         `json:"identificationTypeAdditionalInfo,omitempty"`
	Description                      string                          `json:"description"`
	HasExactVehicleIdentification    *bool                           `json:"hasExactVehicleIdentification"`
	Modality                         *InsuredObjectModality          `json:"modality,omitempty"`
	ModalityOthers                   *string                         `json:"modalityOthers,omitempty"`
	AmountReferenceTable             *AmountReferenceTable           `json:"amountReferenceTable,omitempty"`
	AmountReferenceTableOthers       *string                         `json:"amountReferenceTableOthers,omitempty"`
	Model                            *string                         `json:"model,omitempty"`
	Year                             *string                         `json:"year,omitempty"`
	FareCategory                     *FareCategory                   `json:"fareCategory,omitempty"`
	RiskPostCode                     *string                         `json:"riskPostCode,omitempty"`
	VehicleUsage                     *VehicleUsage                   `json:"vehicleUsage,omitempty"`
	VehicleUsageOthers               *string                         `json:"vehicleUsageOthers,omitempty"`
	FrequentDestinationPostCode      *string                         `json:"frequentDestinationPostCode,omitempty"`
	OvernightPostCode                *string                         `json:"overnightPostCode,omitempty"`
	Coverages                        []InsuredObjectCoverage         `json:"coverages"`
}

type InsuredObjectIdentificationType string

const (
	AUTOMOVEL InsuredObjectIdentificationType = "AUTOMOVEL"
	CONDUTOR  InsuredObjectIdentificationType = "CONDUTOR"
	FROTA     InsuredObjectIdentificationType = "FROTA"
	OUTROS    InsuredObjectIdentificationType = "OUTROS"
)

type InsuredObjectModality string

const (
	InsuredObjectModalityValueDetermined     InsuredObjectModality = "VALOR_DETERMINADO"
	InsuredObjectModalityValueMarketReferred InsuredObjectModality = "VALOR_DE_MERCADO_REFERENCIADO"
	InsuredObjectModalityDifferentCriteria   InsuredObjectModality = "CRITERIO_DIVERSO"
	InsuredObjectModalityOthers              InsuredObjectModality = "OUTROS"
)

type AmountReferenceTable string

const (
	AmountReferenceTableMolicar      AmountReferenceTable = "MOLICAR"
	AmountReferenceTableFipe         AmountReferenceTable = "FIPE"
	AmountReferenceTableJournalOfCar AmountReferenceTable = "JORNAL_DO_CARRO"
	AmountReferenceTableVD           AmountReferenceTable = "VD"
	AmountReferenceTableOthers       AmountReferenceTable = "OUTROS"
)

type FareCategory string

const (
	FareCategory10  FareCategory = "10"
	FareCategory11  FareCategory = "11"
	FareCategory14A FareCategory = "14A"
	FareCategory14B FareCategory = "14B"
	FareCategory14C FareCategory = "14C"
	FareCategory15  FareCategory = "15"
	FareCategory16  FareCategory = "16"
	FareCategory17  FareCategory = "17"
	FareCategory18  FareCategory = "18"
	FareCategory19  FareCategory = "19"
	FareCategory20  FareCategory = "20"
	FareCategory21  FareCategory = "21"
	FareCategory22  FareCategory = "22"
	FareCategory23  FareCategory = "23"
	FareCategory30  FareCategory = "30"
	FareCategory31  FareCategory = "31"
	FareCategory40  FareCategory = "40"
	FareCategory41  FareCategory = "41"
	FareCategory42  FareCategory = "42"
	FareCategory43  FareCategory = "43"
	FareCategory50  FareCategory = "50"
	FareCategory51  FareCategory = "51"
	FareCategory52  FareCategory = "52"
	FareCategory53  FareCategory = "53"
	FareCategory58  FareCategory = "58"
	FareCategory59  FareCategory = "59"
	FareCategory60  FareCategory = "60"
	FareCategory61  FareCategory = "61"
	FareCategory62  FareCategory = "62"
	FareCategory63  FareCategory = "63"
	FareCategory68  FareCategory = "68"
	FareCategory69  FareCategory = "69"
	FareCategory70  FareCategory = "70"
	FareCategory71  FareCategory = "71"
	FareCategory72  FareCategory = "72"
	FareCategory73  FareCategory = "73"
	FareCategory80  FareCategory = "80"
	FareCategory81  FareCategory = "81"
	FareCategory82  FareCategory = "82"
	FareCategory83  FareCategory = "83"
	FareCategory84  FareCategory = "84"
	FareCategory85  FareCategory = "85"
	FareCategory86  FareCategory = "86"
	FareCategory87  FareCategory = "87"
	FareCategory88  FareCategory = "88"
	FareCategory89  FareCategory = "89"
	FareCategory90  FareCategory = "90"
	FareCategory91  FareCategory = "91"
	FareCategory92  FareCategory = "92"
	FareCategory93  FareCategory = "93"
	FareCategory94  FareCategory = "94"
	FareCategory95  FareCategory = "95"
	FareCategory96  FareCategory = "96"
	FareCategory97  FareCategory = "97"
)

type VehicleUsage string

const (
	VehicleUsageLeisure      VehicleUsage = "LAZER"
	VehicleUsageDailyRental  VehicleUsage = "LOCOMOCAO_DIARIA"
	VehicleUsageWorkExercise VehicleUsage = "EXERCICIO_DO_TRABALHO"
	VehicleUsageOthers       VehicleUsage = "OUTROS"
)

type InsuredObjectCoverage struct {
	Branch                        string                    `json:"branch"`
	Code                          string                    `json:"code"`
	Description                   *string                   `json:"description,omitempty"`
	InternalCode                  *string                   `json:"internalCode,omitempty"`
	SusepProcessNumber            string                    `json:"susepProcessNumber"`
	LMI                           insurer.AmountDetails     `json:"LMI,omitempty"`
	TermStartDate                 timeutil.BrazilDate       `json:"termStartDate"`
	TermEndDate                   timeutil.BrazilDate       `json:"termEndDate"`
	IsMainCoverage                bool                      `json:"isMainCoverage"`
	Feature                       CoverageFeature           `json:"feature"`
	Type                          CoverageType              `json:"type"`
	GracePeriod                   *int                      `json:"gracePeriod,omitempty"`
	GracePeriodicity              *Periodicity              `json:"gracePeriodicity,omitempty"`
	GracePeriodCountingMethod     *PeriodCountingMethod     `json:"gracePeriodCountingMethod,omitempty"`
	GracePeriodStartDate          *timeutil.BrazilDate      `json:"gracePeriodStartDate,omitempty"`
	GracePeriodEndDate            *timeutil.BrazilDate      `json:"gracePeriodEndDate,omitempty"`
	AdjustmentRate                *string                   `json:"adjustmentRate,omitempty"`
	PremiumAmount                 insurer.AmountDetails     `json:"premiumAmount"`
	PremiumPeriodicity            PremiumPeriodicity        `json:"premiumPeriodicity"`
	PremiumPeriodicityOthers      *string                   `json:"premiumPeriodicityOthers,omitempty"`
	CompensationType              *CoverageCompensationType `json:"compensationType,omitempty"`
	CompensationTypeOthers        *string                   `json:"compensationTypeOthers,omitempty"`
	PartialCompensationPercentage *string                   `json:"partialCompensationPercentage,omitempty"`
	PercentageOverLMI             *string                   `json:"percentageOverLMI,omitempty"`
	DaysForTotalCompensation      *int                      `json:"daysForTotalCompensation,omitempty"`
	BoundCoverage                 *BoundCoverage            `json:"boundCoverage,omitempty"`
	BoundCoverageOthers           *string                   `json:"boundCoverageOthers,omitempty"`
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

type Periodicity string

const (
	PeriodicityDay   Periodicity = "DIA"
	PeriodicityMonth Periodicity = "MES"
	PeriodicityYear  Periodicity = "ANO"
)

type PeriodCountingMethod string

const (
	PeriodCountingMethodBusinessDays PeriodCountingMethod = "DIAS_UTEIS"
	PeriodCountingMethodCalendarDays PeriodCountingMethod = "DIAS_CORRIDOS"
)

type PremiumPeriodicity string

const (
	PremiumPeriodicityMonthly       PremiumPeriodicity = "MENSAL"
	PremiumPeriodicityBimonthly     PremiumPeriodicity = "BIMESTRAL"
	PremiumPeriodicityQuarterly     PremiumPeriodicity = "TRIMESTRAL"
	PremiumPeriodicityQuadrimestral PremiumPeriodicity = "QUADRIMESTRAL"
	PremiumPeriodicitySemiannual    PremiumPeriodicity = "SEMESTRAL"
	PremiumPeriodicityAnnual        PremiumPeriodicity = "ANUAL"
	PremiumPeriodicityOneTime       PremiumPeriodicity = "PAGAMENTO_UNICO"
	PremiumPeriodicityEsporadic     PremiumPeriodicity = "ESPORADICA"
	PremiumPeriodicityOthers        PremiumPeriodicity = "OUTROS"
)

type CoverageCompensationType string

const (
	CoverageCompensationTypeIntegral CoverageCompensationType = "INTEGRAL"
	CoverageCompensationTypePartial  CoverageCompensationType = "PARCIAL"
	CoverageCompensationTypeOthers   CoverageCompensationType = "OUTROS"
)

type BoundCoverage string

const (
	BoundCoverageVehicle BoundCoverage = "VEICULO"
	BoundCoverageDriver  BoundCoverage = "CONDUTOR"
	BoundCoverageOthers  BoundCoverage = "OUTROS"
)

type Coverage struct {
	Branch      *string             `json:"branch,omitempty"`
	Code        CoverageCode        `json:"code"`
	Description *string             `json:"description,omitempty"`
	Deductible  *CoverageDeductible `json:"deductible,omitempty"`
	POS         *CoveragePOS        `json:"POS,omitempty"`
}

type CoverageCode string

const (
	CoverageCodeComprehensive                     CoverageCode = "CASCO_COMPREENSIVA"
	CoverageCodeFireTheftBurglary                 CoverageCode = "CASCO_INCENDIO_ROUBO_E_FURTO"
	CoverageCodeTheftBurglary                     CoverageCode = "CASCO_ROUBO_E_FURTO"
	CoverageCodeFire                              CoverageCode = "CASCO_INCENDIO"
	CoverageCodeFlood                             CoverageCode = "CASCO_ALAGAMENTO"
	CoverageCodeCollisionPartial                  CoverageCode = "CASCO_COLISAO_INDENIZACAO_PARCIAL"
	CoverageCodeCollisionFull                     CoverageCode = "CASCO_COLISAO_INDENIZACAO_INTEGRAL"
	CoverageCodeOptionalVehicleLiability          CoverageCode = "RESPONSABILIDADE_CIVIL_FACULTATIVA_DE_VEICULOS_RCFV"
	CoverageCodeOptionalDriverLiability           CoverageCode = "RESPONSABILIDADE_CIVIL_FACULTATIVA_DO_CONDUTOR_RCFC"
	CoverageCodePassengerAccidentVehicle          CoverageCode = "ACIDENTE_PESSOAIS_DE_PASSAGEIROS_APP_VEICULO"
	CoverageCodePassengerAccidentDriver           CoverageCode = "ACIDENTE_PESSOAIS_DE_PASSAGEIROS_APP_CONDUTOR"
	CoverageCodeGlass                             CoverageCode = "VIDROS"
	CoverageCodeDailyUnavailability               CoverageCode = "DIARIA_POR_INDISPONIBILIDADE"
	CoverageCodeLightsHeadlightsRearviewMirrors   CoverageCode = "LFR_LANTERNAS_FAROIS_E_RETROVISORES"
	CoverageCodeAccessoriesEquipment              CoverageCode = "ACESSORIOS_E_EQUIPAMENTOS"
	CoverageCodeReserveCar                        CoverageCode = "CARRO_RESERVA"
	CoverageCodeMinorRepairs                      CoverageCode = "PEQUENOS_REPAROS"
	CoverageCodeGreenCard                         CoverageCode = "RESPONSABILIDADE_CIVIL_CARTA_VERDE"
	CoverageCodeLiabilityPassengerCarsNonMercosur CoverageCode = "RESPONSABILIDADE_CIVIL_VEICULOS_DE_PASSEIO_ACORDOS_FORA_DO_MERCOSUL"
	CoverageCodeOthers                            CoverageCode = "OUTRAS"
)

type CoverageDeductible struct {
	Type                               CoverageDeductibleType `json:"type"`
	TypeOthers                         *string                `json:"typeOthers,omitempty"`
	Amount                             *insurer.AmountDetails `json:"amount,omitempty"`
	Period                             *int                   `json:"period,omitempty"`
	Periodicity                        *Periodicity           `json:"periodicity,omitempty"`
	PeriodCountingMethod               *PeriodCountingMethod  `json:"periodCountingMethod,omitempty"`
	PeriodStartDate                    *timeutil.BrazilDate   `json:"periodStartDate,omitempty"`
	PeriodEndDate                      *timeutil.BrazilDate   `json:"periodEndDate,omitempty"`
	Description                        *string                `json:"description,omitempty"`
	HasDeductibleOverTotalCompensation *bool                  `json:"hasDeductibleOverTotalCompensation,omitempty"`
}

type CoverageDeductibleType string

const (
	CoverageDeductibleTypeReduced   CoverageDeductibleType = "REDUZIDA"
	CoverageDeductibleTypeNormal    CoverageDeductibleType = "NORMAL"
	CoverageDeductibleTypeIncreased CoverageDeductibleType = "MAJORADA"
	CoverageDeductibleTypeExempt    CoverageDeductibleType = "ISENTA"
	CoverageDeductibleTypeFlexible  CoverageDeductibleType = "FLEXIVEL"
	CoverageDeductibleTypeOthers    CoverageDeductibleType = "OUTROS"
)

type CoveragePOS struct {
	ApplicationType CoveragePOSApplicationType `json:"applicationType"`
	Description     *string                    `json:"description,omitempty"`
	MinValue        *insurer.AmountDetails     `json:"minValue,omitempty"`
	MaxValue        *insurer.AmountDetails     `json:"maxValue,omitempty"`
	Percentage      *insurer.AmountDetails     `json:"percentage,omitempty"`
	ValueOthers     *insurer.AmountDetails     `json:"valueOthers,omitempty"`
}

type CoveragePOSApplicationType string

const (
	CoveragePOSApplicationTypeValue      CoveragePOSApplicationType = "VALOR"
	CoveragePOSApplicationTypePercentage CoveragePOSApplicationType = "PERCENTUAL"
	CoveragePOSApplicationTypeOthers     CoveragePOSApplicationType = "OUTROS"
)

type Coinsurer struct {
	Identification  string `json:"identification"`
	CededPercentage string `json:"cededPercentage"`
}

type RepairNetwork string

const (
	RepairNetworkFreeChoice RepairNetwork = "LIVRE_ESCOLHA"
	RepairNetworkReferred   RepairNetwork = "REDE_REFERENCIADA"
	RepairNetworkBoth       RepairNetwork = "AMBAS"
	RepairNetworkOthers     RepairNetwork = "OUTROS"
)

type RepairedPartsUsageType string

const (
	RepairedPartsUsageTypeNew        RepairedPartsUsageType = "NOVA"
	RepairedPartsUsageTypeUsed       RepairedPartsUsageType = "USADA"
	RepairedPartsUsageTypeNewAndUsed RepairedPartsUsageType = "NOVA_E_USADA"
)

type RepairedPartsClassification string

const (
	RepairedPartsClassificationOriginal              RepairedPartsClassification = "ORIGINAL"
	RepairedPartsClassificationCompatible            RepairedPartsClassification = "COMPATIVEL"
	RepairedPartsClassificationOriginalAndCompatible RepairedPartsClassification = "ORIGINAL_E_COMPATIVEL"
)

type RepairedPartsNationality string

const (
	RepairedPartsNationalityNational            RepairedPartsNationality = "NACIONAL"
	RepairedPartsNationalityImported            RepairedPartsNationality = "IMPORTADA"
	RepairedPartsNationalityNationalAndImported RepairedPartsNationality = "NACIONAL_E_IMPORTADA"
)

type OtherBenefits string

const (
	OtherBenefitsFreeRaffle   OtherBenefits = "SORTEIO_GRATUITO"
	OtherBenefitsBenefitsClub OtherBenefits = "CLUBE_DE_BENEFICIOS"
	OtherBenefitsCashBack     OtherBenefits = "CASH_BACK"
	OtherBenefitsDiscounts    OtherBenefits = "DESCONTOS"
	OtherBenefitsCustomizable OtherBenefits = "CUSTOMIZAVEL"
)

type AssistancePackages string

const (
	AssistancePackagesUpTo10Services  AssistancePackages = "ATE_10_SERVICOS"
	AssistancePackagesUpTo20Services  AssistancePackages = "ATE_20_SERVICOS"
	AssistancePackagesAbove20Services AssistancePackages = "ACIMA_DE_20_SERVICOS"
	AssistancePackagesCustomizable    AssistancePackages = "CUSTOMIZAVEL"
)

type Driver struct {
	Identification     *string              `json:"identification,omitempty"`
	Sex                *Sex                 `json:"sex,omitempty"`
	SexOthers          *string              `json:"sexOthers,omitempty"`
	BirthDate          *timeutil.BrazilDate `json:"birthDate,omitempty"`
	LicensedExperience *int                 `json:"licensedExperience,omitempty"`
}

type Sex string

const (
	SexMale      Sex = "MASCULINO"
	SexFemale    Sex = "FEMININO"
	SexUndefined Sex = "NAO_DECLARADO"
	SexOther     Sex = "OUTROS"
)

type ClaimStatus string

const (
	ClaimStatusOpen                       ClaimStatus = "ABERTO"
	ClaimStatusClosedWithCompensation     ClaimStatus = "ENCERRADO_COM_INDENIZACAO"
	ClaimStatusClosedWithoutCompensation  ClaimStatus = "ENCERRADO_SEM_INDENIZACAO"
	ClaimStatusReopened                   ClaimStatus = "REABERTO"
	ClaimStatusCanceledByOperationalError ClaimStatus = "CANCELADO_POR_ERRO_OPERACIONAL"
	ClaimStatusInitialEvaluation          ClaimStatus = "AVALIACAO_INICIAL"
)

type DenialJustification string

const (
	DenialJustificationRiskExcluded            DenialJustification = "RISCO_EXCLUIDO"
	DenialJustificationRiskAgravated           DenialJustification = "RISCO_AGRAVADO"
	DenialJustificationNoDocumentation         DenialJustification = "SEM_DOCUMENTACAO"
	DenialJustificationIncompleteDocumentation DenialJustification = "DOCUMENTACAO_INCOMPLETA"
	DenialJustificationPrescription            DenialJustification = "PRESCRICAO"
	DenialJustificationOutOfCoverage           DenialJustification = "FORA_COBERTURA"
	DenialJustificationOthers                  DenialJustification = "OUTROS"
)

type ClaimCoverage struct {
	InsuredObjectId     *string              `json:"insuredObjectId,omitempty"`
	Branch              string               `json:"branch"`
	Code                CoverageCode         `json:"code"`
	Description         *string              `json:"description,omitempty"`
	WarningDate         *timeutil.BrazilDate `json:"warningDate,omitempty"`
	ThirdPartyClaimDate *timeutil.BrazilDate `json:"thirdPartyClaimDate,omitempty"`
}

type CoverageBranchInfo struct {
	CovenantNumber              *string              `json:"covenantNumber,omitempty"`
	OccurenceCause              *OccurenceCause      `json:"occurenceCause,omitempty"`
	OccurenceCauseOthers        *string              `json:"occurenceCauseOthers,omitempty"`
	DriverAtOccurrenceSex       *Sex                 `json:"driverAtOccurrenceSex,omitempty"`
	DriverAtOccurrenceSexOthers *string              `json:"driverAtOccurrenceSexOthers,omitempty"`
	DriverAtOccurrenceBirthDate *timeutil.BrazilDate `json:"driverAtOccurrenceBirthDate,omitempty"`
	OccurrenceCountry           *CountryCode         `json:"occurrenceCountry,omitempty"`
	OccurrencePostCode          *string              `json:"occurrencePostCode,omitempty"`
}

type OccurenceCause string

const (
	OccurenceCauseRobberyOrTheft   OccurenceCause = "ROUBO_OU_FURTO"
	OccurenceCauseRobbery          OccurenceCause = "ROUBO"
	OccurenceCauseTheft            OccurenceCause = "FURTO"
	OccurenceCauseCollisionPartial OccurenceCause = "COLISAO_PARCIAL"
	OccurenceCauseCollisionFull    OccurenceCause = "COLISAO_INDENIZACAO_INTEGRAL"
	OccurenceCauseFire             OccurenceCause = "INCENDIO"
	OccurenceCauseFlood            OccurenceCause = "ALAGAMENTO"
	OccurenceCauseOthers           OccurenceCause = "OUTROS"
)

type CountryCode string

const (
	CountryCodeBrazil CountryCode = "BRA"
)

type Filter struct {
	OwnerID string
}

func policyID() string {
	return ""
}
