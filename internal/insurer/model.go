package insurer

const (
	Brand string = "Mock Insurer"
	CNPJ  string = "00000000000000"
)

type IdentificationType string

const (
	IdentificationTypeCPF    IdentificationType = "CPF"
	IdentificationTypeCNPJ   IdentificationType = "CNPJ"
	IdentificationTypeOthers IdentificationType = "OUTROS"
)

type ValidityType string

const (
	ValidityTypeAnnual                 ValidityType = "ANUAL"              //nolint:misspell
	ValidityTypeAnnualIntermittent     ValidityType = "ANUAL_INTERMITENTE" //nolint:misspell
	ValidityTypePlurianual             ValidityType = "PLURIANUAL"
	ValidityTypePlurianualIntermittent ValidityType = "PLURIANUAL_INTERMITENTE"
	ValidityTypeSemestral              ValidityType = "SEMESTRAL"
	ValidityTypeSemestralIntermittent  ValidityType = "SEMESTRAL_INTERMITENTE"
	ValidityTypeMonthly                ValidityType = "MENSAL"
	ValidityTypeMonthlyIntermittent    ValidityType = "MENSAL_INTERMITENTE"
	ValidityTypeDaily                  ValidityType = "DIARIO"
	ValidityTypeDailyIntermittent      ValidityType = "DIARIO_INTERMITENTE"
	ValidityTypeOthers                 ValidityType = "OUTROS"
)

type AmountDetails struct {
	Amount         string   `json:"amount"`
	UnitType       UnitType `json:"unitType"`
	UnitTypeOthers *string  `json:"unitTypeOthers,omitempty"`
	Unit           *Unit    `json:"unit,omitempty"`
}

type UnitType string

const (
	UnitTypePercentage UnitType = "PORCENTAGEM"
	UnitTypeMonetary   UnitType = "MONETARIO"
	UnitTypeOthers     UnitType = "OUTROS"
)

type Unit struct {
	Code        UnitCode `json:"code"`
	Description Currency `json:"description"`
}

type UnitCode string

const (
	UnitCodeReal UnitCode = "R$"
)

type Currency string

const (
	CurrencyBRL Currency = "BRL"
)

type CountryCode string

const (
	CountryCodeBrazil CountryCode = "BRA"
)

type CountrySubDivision string

const (
	CountrySubDivisionAC CountrySubDivision = "AC"
	CountrySubDivisionAL CountrySubDivision = "AL"
	CountrySubDivisionAM CountrySubDivision = "AM"
	CountrySubDivisionAP CountrySubDivision = "AP"
	CountrySubDivisionBA CountrySubDivision = "BA"
	CountrySubDivisionCE CountrySubDivision = "CE"
	CountrySubDivisionDF CountrySubDivision = "DF"
	CountrySubDivisionES CountrySubDivision = "ES"
	CountrySubDivisionGO CountrySubDivision = "GO"
	CountrySubDivisionMA CountrySubDivision = "MA"
	CountrySubDivisionRJ CountrySubDivision = "RJ"
)

type PhoneAreaCode string

const (
	PhoneAreaCode11 PhoneAreaCode = "11"
	PhoneAreaCode12 PhoneAreaCode = "12"
	PhoneAreaCode13 PhoneAreaCode = "13"
	PhoneAreaCode14 PhoneAreaCode = "14"
	PhoneAreaCode15 PhoneAreaCode = "15"
	PhoneAreaCode16 PhoneAreaCode = "16"
	PhoneAreaCode17 PhoneAreaCode = "17"
	PhoneAreaCode18 PhoneAreaCode = "18"
	PhoneAreaCode19 PhoneAreaCode = "19"
)

type CivilStatus string

const (
	CivilStatusSingle   CivilStatus = "SOLTEIRO"
	CivilStatusMarried  CivilStatus = "CASADO"
	CivilStatusDivorced CivilStatus = "DIVORCIADO"
	CivilStatusWidowed  CivilStatus = "VIUVO"
)

type ValueType string

const (
	ValueTypeValue      ValueType = "VALOR"
	ValueTypePercentage ValueType = "PERCENTUAL"
	ValueTypeOthers     ValueType = "OUTROS"
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
	PremiumPeriodicityAnnual        PremiumPeriodicity = "ANUAL" //nolint:misspell
	PremiumPeriodicityOneTime       PremiumPeriodicity = "PAGAMENTO_UNICO"
	PremiumPeriodicityEsporadic     PremiumPeriodicity = "ESPORADICA"
	PremiumPeriodicityOthers        PremiumPeriodicity = "OUTROS"
)

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
	PaymentTypeOthers          PaymentType = "OUTROS"
)
