package quote

import (
	"github.com/luikyv/mock-insurer/internal/customer"
	"github.com/luikyv/mock-insurer/internal/insurer"
)

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
