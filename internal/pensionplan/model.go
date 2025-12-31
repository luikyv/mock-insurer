package pensionplan

import (
	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/timeutil"
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

type ContractData struct {
	ProductName       string `json:"productName"`
	ProductCode       string `json:"productCode"`
	ConjugatedPlan    bool   `json:"conjugatedPlan"`
	ProposalID        string `json:"proposalId"`
	CertificateActive bool   `json:"certificateActive"`
	// Insured            Insured             `json:"insured"`
	// Intermediary       Intermediary        `json:"intermediary"`
	ContractingType    ContractingType     `json:"contractingType"`
	ContractID         string              `json:"contractId"`
	PlanType           PlanType            `json:"planType"`
	EffectiveDateStart timeutil.BrazilDate `json:"effectiveDateStart"`
	EffectiveDateEnd   timeutil.BrazilDate `json:"effectiveDateEnd"`
	// Beneficiaries      []Beneficiary       `json:"beneficiaries"`
	Periodicity       Periodicity `json:"periodicity"`
	PeriodicityOthers *string     `json:"periodicityOthers,omitempty"`
	TaxRegime         TaxRegime   `json:"taxRegime"`
	// Suseps             []Suseps            `json:"suseps"`
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
	PeriodicityMonthly    Periodicity = "MENSAL"
	PeriodicityBimonthly  Periodicity = "BIMESTRAL"
	PeriodicityQuarterly  Periodicity = "TRIMESTRAL"
	PeriodicitySemiannual Periodicity = "SEMESTRAL"
	PeriodicityAnnual     Periodicity = "ANUAL" //nolint:misspell
	PeriodicityOneTime    Periodicity = "PAGAMENTO_UNICO"
	PeriodicityOthers     Periodicity = "OUTROS"
)

type TaxRegime string

const (
	TaxRegimeProgressive TaxRegime = "PROGRESSIVO" //nolint:misspell
	TaxRegimeRegressive  TaxRegime = "REGRESSIVO"  //nolint:misspell
)
