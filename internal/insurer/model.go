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
	Code        UnitCode        `json:"code"`
	Description UnitDescription `json:"description"`
}

type UnitCode string

const (
	UnitCodeReal UnitCode = "R$"
)

type UnitDescription string

const (
	UnitDescriptionBRL UnitDescription = "BRL"
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
