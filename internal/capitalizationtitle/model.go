package capitalizationtitle

import (
	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/resource"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

var (
	Scope = goidc.NewScope("capitalization-title")
)

type Plan struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      PlanData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (Plan) TableName() string {
	return "insurance_capitalization_title_plans"
}

func (p *Plan) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type ConsentPlan struct {
	ConsentID uuid.UUID
	PlanID    uuid.UUID
	OwnerID   uuid.UUID
	Status    resource.Status
	Plan      *Plan
	OrgID     string
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (ConsentPlan) TableName() string {
	return "consent_insurance_capitalization_title_plans"
}

type PlanData struct {
	Series []PlanSeries `json:"series"`
}

type PlanSeries struct {
	ID                           string    `json:"id"`
	Modality                     Modality  `json:"modality"`
	SusepProcessNumber           string    `json:"susepProcessNumber"`
	CommercialName               *string   `json:"commercialName,omitempty"`
	SerieSize                    int       `json:"serieSize"`
	Quotas                       []Quota   `json:"quotas"`
	GracePeriodRedemption        *int      `json:"gracePeriodRedemption,omitempty"`
	GracePeriodForFullRedemption int       `json:"gracePeriodForFullRedemption"`
	UpdateIndex                  Index     `json:"updateIndex"`
	UpdateIndexOthers            *string   `json:"updateIndexOthers,omitempty"`
	ReadjustmentIndex            Index     `json:"readjustmentIndex"`
	ReadjustmentIndexOthers      *string   `json:"readjustmentIndexOthers,omitempty"`
	BonusClause                  bool      `json:"bonusClause"`
	Frequency                    Frequency `json:"frequency"`
	FrequencyDescription         *string   `json:"frequencyDescription,omitempty"`
	InterestRate                 string    `json:"interestRate"`
	Brokers                      *[]Broker `json:"brokers,omitempty"`
	Titles                       []Title   `json:"titles"`
}

type Event struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	PlanID    uuid.UUID
	Data      EventData `gorm:"serializer:json"`
	CrossOrg  bool
	OrgID     string
	CreatedAt timeutil.DateTime `gorm:"autoCreateTime"`
	UpdatedAt timeutil.DateTime `gorm:"autoUpdateTime"`
}

func (Event) TableName() string {
	return "insurance_capitalization_title_events"
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type EventData struct {
	TitleID    *string     `json:"titleId,omitempty"`
	Type       *EventType  `json:"type,omitempty"`
	Raffle     *Raffle     `json:"raffle,omitempty"`
	Redemption *Redemption `json:"redemption,omitempty"`
}

type Settlement struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	PlanID    uuid.UUID
	Data      SettlementData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime `gorm:"autoCreateTime"`
	UpdatedAt timeutil.DateTime `gorm:"autoUpdateTime"`
}

func (Settlement) TableName() string {
	return "insurance_capitalization_title_settlements"
}

func (s *Settlement) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type SettlementData struct {
	FinancialAmount insurer.AmountDetails `json:"financialAmount"`
	PaymentDate     timeutil.BrazilDate   `json:"paymentDate"`
	DueDate         timeutil.BrazilDate   `json:"dueDate"`
}

type Raffle struct {
	Amount         insurer.AmountDetails `json:"amount"`
	Date           timeutil.BrazilDate   `json:"date"`
	SettlementDate timeutil.BrazilDate   `json:"settlementDate"`
}

type Redemption struct {
	Amount           insurer.AmountDetails  `json:"amount"`
	BonusAmount      insurer.AmountDetails  `json:"bonusAmount"`
	Date             timeutil.BrazilDate    `json:"date"`
	SettlementDate   timeutil.BrazilDate    `json:"settlementDate"`
	Type             RedemptionType         `json:"type"`
	UnreturnedAmount *insurer.AmountDetails `json:"unreturnedAmount,omitempty"`
}

type RedemptionType string

const (
	RedemptionTypePartialAnticipation RedemptionType = "ANTECIPADO_PARCIAL"
	RedemptionTypeTotalAnticipation   RedemptionType = "ANTECIPADO_TOTAL"
	RedemptionTypeFinalValidity       RedemptionType = "FINAL_VIGENCIA"
)

type Modality string

const (
	ModalityTraditional         Modality = "TRADICIONAL"
	ModalityGuaranteeInstrument Modality = "INSTRUMENTO_GARANTIA"
	ModalityScheduledPurchase   Modality = "COMPRA_PROGRAMADA"
	ModalityPrizePhilanthropy   Modality = "FILANTROPIA_PREMIAVEL"
	ModalityPopular             Modality = "POPULAR"
)

type Quota struct {
	Number              int    `json:"number"`
	CapitalizationQuota string `json:"capitalizationQuota"`
	ChargingQuota       string `json:"chargingQuota"`
	RaffleQuota         string `json:"raffleQuota"`
}

type Index string

const (
	IndexBasicRemunerationSavingsDeposits Index = "INDICE_REMUNERACAO_BASICA_DEPOSITOS_POUPANCA"
	IndexIPCA                             Index = "IPCA"
	IndexINCC                             Index = "INCC"
	IndexINPC                             Index = "INPC"
	IndexIGPM                             Index = "IGPM"
	IndexOthers                           Index = "OUTROS"
)

type Frequency string

const (
	FrequencyOnce     Frequency = "PAGAMENTO_UNICO"
	FrequencyMonthly  Frequency = "PAGAMENTO_MENSAL"
	FrequencyPeriodic Frequency = "PAGAMENTO_PERIODICO"
)

type Broker struct {
	SusepBrokerCode   string `json:"susepBrokerCode"`
	BrokerDescription string `json:"brokerDescription"`
}

type Title struct {
	ID                  string                `json:"id"`
	RegistrationForm    string                `json:"registrationForm"`
	IssueTitleDate      timeutil.BrazilDate   `json:"issueTitleDate"`
	TermStartDate       timeutil.BrazilDate   `json:"termStartDate"`
	TermEndDate         timeutil.BrazilDate   `json:"termEndDate"`
	RafflePremiumAmount insurer.AmountDetails `json:"rafflePremiumAmount"`
	ContributionAmount  insurer.AmountDetails `json:"contributionAmount"`
	Subscribers         []Subscriber          `json:"subscribers"`
	TechnicalProvisions []TechnicalProvision  `json:"technicalProvisions"`
}

type Subscriber struct {
	Name                  string                     `json:"name"`
	DocumentType          DocumentType               `json:"documentType"`
	DocumentTypeOthers    *string                    `json:"documentTypeOthers,omitempty"`
	DocumentNumber        string                     `json:"documentNumber"`
	Phones                *[]Phone                   `json:"phones,omitempty"`
	Address               string                     `json:"address"`
	AddressAdditionalInfo *string                    `json:"addressAdditionalInfo,omitempty"`
	TownName              string                     `json:"townName"`
	CountrySubDivision    insurer.CountrySubDivision `json:"countrySubDivision"`
	CountryCode           string                     `json:"countryCode"`
	PostCode              string                     `json:"postCode"`
	Holders               *[]Holder                  `json:"holders,omitempty"`
}

type DocumentType string

const (
	DocumentTypeCPF      DocumentType = "CPF"
	DocumentTypeCNPJ     DocumentType = "CNPJ"
	DocumentTypePassport DocumentType = "PASSPORTE"
	DocumentTypeOthers   DocumentType = "OUTROS"
)

type Phone struct {
	CountryCallingCode *string                `json:"countryCallingCode,omitempty"`
	AreaCode           *insurer.PhoneAreaCode `json:"areaCode,omitempty"`
	Number             *string                `json:"number,omitempty"`
}

type Holder struct {
	Name                  string                     `json:"name"`
	DocumentType          DocumentType               `json:"documentType"`
	DocumentTypeOthers    *string                    `json:"documentTypeOthers,omitempty"`
	DocumentNumber        string                     `json:"documentNumber"`
	Phones                *[]Phone                   `json:"phones,omitempty"`
	Address               string                     `json:"address"`
	AddressAdditionalInfo *string                    `json:"addressAdditionalInfo,omitempty"`
	TownName              string                     `json:"townName"`
	CountrySubDivision    insurer.CountrySubDivision `json:"countrySubDivision"`
	CountryCode           string                     `json:"countryCode"`
	PostCode              string                     `json:"postCode"`
	Redemption            bool                       `json:"redemption"`
	Raffle                bool                       `json:"raffle"`
}

type TechnicalProvision struct {
	PMCAmount insurer.AmountDetails `json:"pmcAmount"`
	PRAmount  insurer.AmountDetails `json:"prAmount"`
	PSPAmount insurer.AmountDetails `json:"pspAmount"`
	PDBAmount insurer.AmountDetails `json:"pdbAmount"`
}

type EventType string

const (
	EventTypeRaffle     EventType = "SORTEIO"
	EventTypeRedemption EventType = "RESGATE"
)
