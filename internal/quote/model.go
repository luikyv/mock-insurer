package quote

import (
	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/customer"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

type Lead interface {
	GetID() uuid.UUID
	GetStatus() Status
	SetStatus(Status)
	SetStatusUpdatedAt(timeutil.DateTime)
	SetUpdatedAt(timeutil.DateTime)
	SetCreatedAt(timeutil.DateTime)
	GetOrgID() string
}

type Quote interface {
	GetID() uuid.UUID
	GetStatus() Status
	SetStatus(Status)
	SetStatusUpdatedAt(timeutil.DateTime)
	SetUpdatedAt(timeutil.DateTime)
	SetCreatedAt(timeutil.DateTime)
	GetTermStartDate() timeutil.BrazilDate
	GetTermEndDate() timeutil.BrazilDate
	SetRejectionReason(string)
	SetInsurerQuoteID(string)
	SetProtocolDateTime(timeutil.DateTime)
	SetProtocolNumber(string)
	SetRedirectLink(string)
	GetPersonalIdentification() *string
	GetBusinessIdentification() *string
	GetOfferIDs() []string
	CreateOffers()
	GetOrgID() string
}

type Status string

const (
	StatusReceived     Status = "RCVD"
	StatusEvaluated    Status = "EVAL"
	StatusAccepted     Status = "ACPT"
	StatusRejected     Status = "RJCT"
	StatusAcknowledged Status = "ACKN"
	StatusCancelled    Status = "CANC"
)

type Customer struct {
	Personal *PersonalData `json:"personal,omitempty"`
	Business *BusinessData `json:"business,omitempty"`
}

type PersonalData struct {
	Identification    *customer.PersonalIdentificationData           `json:"identification,omitempty"`
	Qualification     *customer.PersonalQualificationData            `json:"qualification,omitempty"`
	ComplimentaryInfo *customer.PersonalComplimentaryInformationData `json:"complimentaryInfo,omitempty"`
}

type BusinessData struct {
	Identification    *customer.BusinessIdentificationData           `json:"identification,omitempty"`
	Qualification     *customer.BusinessQualificationData            `json:"qualification,omitempty"`
	ComplimentaryInfo *customer.BusinessComplimentaryInformationData `json:"complimentaryInfo,omitempty"`
}

type CustomData struct {
	CustomerIdentification    *[]CustomDataField `json:"customerIdentification,omitempty"`
	CustomerQualification     *[]CustomDataField `json:"customerQualification,omitempty"`
	CustomerComplimentaryInfo *[]CustomDataField `json:"customerComplimentaryInfo,omitempty"`
	GeneralQuoteInfo          *[]CustomDataField `json:"generalQuoteInfo,omitempty"`
	RiskLocationInfo          *[]CustomDataField `json:"riskLocationInfo,omitempty"`
	InsuredObjects            *[]CustomDataField `json:"insuredObjects,omitempty"`
	Beneficiaries             *[]CustomDataField `json:"beneficiaries,omitempty"`
	Coverages                 *[]CustomDataField `json:"coverages,omitempty"`
	GeneralClaimInfo          *[]CustomDataField `json:"generalClaimInfo,omitempty"`
}

type CustomDataField struct {
	FieldID string `json:"fieldId"`
	Value   any    `json:"value"`
}

type PatchData struct {
	Status                     Status
	InsurerQuoteID             *string
	AuthorIdentificationType   insurer.IdentificationType
	AuthorIdentificationNumber string
}

type LeadQuery struct {
	ID        string
	ConsentID string
}

type Query struct {
	ID        string
	ConsentID string
}
