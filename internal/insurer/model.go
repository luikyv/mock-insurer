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
	ValidityTypeAnnual                 ValidityType = "ANUAL"
	ValidityTypeAnnualIntermittent     ValidityType = "ANUAL_INTERMITENTE"
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

type Currency string

const (
	CurrencyBRL Currency = "BRL"
)

type AmountDetails struct {
	Amount         string  `json:"amount"`
	UnitType       string  `json:"unitType"`
	UnitTypeOthers *string `json:"unitTypeOthers,omitempty"`
	Unit           *Unit   `json:"unit,omitempty"`
}

type Unit struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}
