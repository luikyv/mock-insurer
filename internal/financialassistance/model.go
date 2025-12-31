package financialassistance

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
	Scope = goidc.NewScope("insurance-financial-assistance")
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
	return "insurance_financial_assistance_contracts"
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
	return "consent_insurance_financial_assistance_contracts"
}

type Movement struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	ContractID   string
	MovementData MovementData `gorm:"serializer:json"`
	OrgID        string
	CrossOrg     bool
	CreatedAt    timeutil.DateTime
	UpdatedAt    timeutil.DateTime
}

func (Movement) TableName() string {
	return "insurance_financial_assistance_movements"
}

func (m *Movement) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type MovementData struct {
	UpdatedDebitAmount                         insurer.AmountDetails  `json:"updatedDebitAmount"`
	RemainingCounterInstallmentsQuantity       int                    `json:"remainingCounterInstallmentsQuantity"`
	RemainingUnpaidCounterInstallmentsQuantity int                    `json:"remainingUnpaidCounterInstallmentsQuantity"`
	LifePensionPMBACAmount                     *insurer.AmountDetails `json:"lifePensionPmBacAmount,omitempty"`
	PensionPlanPMBACAmount                     *insurer.AmountDetails `json:"pensionPlanPmBacAmount,omitempty"`
}

type ContractData struct {
	CertificateID           string                 `json:"certificateId"`
	GroupContractID         *string                `json:"groupContractId,omitempty"`
	SusepProcessNumber      string                 `json:"susepProcessNumber"`
	Insureds                []Insured              `json:"insureds"`
	ConceivedCreditValue    insurer.AmountDetails  `json:"conceivedCreditValue"`
	CreditedLiquidValue     insurer.AmountDetails  `json:"creditedLiquidValue"`
	CounterInstallments     CounterInstallment     `json:"counterInstallments"`
	InterestRate            insurer.AmountDetails  `json:"interestRate"`
	EffectiveCostRate       insurer.AmountDetails  `json:"effectiveCostRate"`
	AmortizationPeriod      int                    `json:"amortizationPeriod"`
	AcquittanceValue        *insurer.AmountDetails `json:"acquittanceValue,omitempty"`
	AcquittanceDate         *timeutil.BrazilDate   `json:"acquittanceDate,omitempty"`
	TaxesValue              insurer.AmountDetails  `json:"taxesValue"`
	ExpensesValue           *insurer.AmountDetails `json:"expensesValue,omitempty"`
	FinesValue              *insurer.AmountDetails `json:"finesValue,omitempty"`
	MonetaryUpdatesValue    *insurer.AmountDetails `json:"monetaryUpdatesValue,omitempty"`
	AdministrativeFeesValue insurer.AmountDetails  `json:"administrativeFeesValue"`
	InterestValue           insurer.AmountDetails  `json:"interestValue"`
}

type Insured struct {
	DocumentType       DocumentType `json:"documentType"`
	DocumentTypeOthers *string      `json:"documentTypeOthers,omitempty"`
	DocumentNumber     string       `json:"documentNumber"`
	Name               string       `json:"name"`
}

type DocumentType string

const (
	DocumentTypeCPF      DocumentType = "CPF"
	DocumentTypeCNPJ     DocumentType = "CNPJ"
	DocumentTypePassport DocumentType = "PASSAPORTE"
	DocumentTypeOthers   DocumentType = "OUTROS"
)

type CounterInstallment struct {
	FirstDate   timeutil.BrazilDate           `json:"firstDate"`
	LastDate    timeutil.BrazilDate           `json:"lastDate"`
	Periodicity CounterInstallmentPeriodicity `json:"periodicity"`
	Quantity    int                           `json:"quantity"`
	Value       insurer.AmountDetails         `json:"value"`
}

type CounterInstallmentPeriodicity string

const (
	CounterInstallmentPeriodicityMonthly    CounterInstallmentPeriodicity = "MENSAL"
	CounterInstallmentPeriodicityBimonthly  CounterInstallmentPeriodicity = "BIMESTRAL"
	CounterInstallmentPeriodicityQuarterly  CounterInstallmentPeriodicity = "TRIMESTRAL"
	CounterInstallmentPeriodicitySemiannual CounterInstallmentPeriodicity = "SEMESTRAL"
	CounterInstallmentPeriodicityAnnual     CounterInstallmentPeriodicity = "ANUAL"
)
