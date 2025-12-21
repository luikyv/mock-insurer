package customer

import "github.com/luikyv/mock-insurer/internal/timeutil"

type PersonalIdentificationData struct {
	UpdateDateTime          timeutil.DateTime        `json:"updateDateTime"`
	PersonalID              *string                  `json:"personalId,omitempty"`
	BrandName               string                   `json:"brandName"`
	CivilName               string                   `json:"civilName"`
	SocialName              *string                  `json:"socialName,omitempty"`
	CPF                     string                   `json:"cpf"`
	CompanyInfo             CompanyInfo              `json:"companyInfo"`
	Documents               *[]PersonalDocument      `json:"documents,omitempty"`
	HasBrazilianNationality *bool                    `json:"hasBrazilianNationality,omitempty"`
	OtherDocuments          *[]OtherPersonalDocument `json:"otherDocuments,omitempty"`
}

type PersonalQualificationData struct{}

type PersonalComplimentaryInformationData struct{}

type BusinessIdentificationData struct {
}

type BusinessQualificationData struct{}

type BusinessComplimentaryInformationData struct{}

type CompanyInfo struct {
	CNPJ string `json:"cnpj"`
	Name string `json:"name"`
}

type PersonalDocument struct {
	Type           *PersonalDocumentType `json:"type,omitempty"`
	Number         *string               `json:"number,omitempty"`
	ExpirationDate *timeutil.BrazilDate  `json:"expirationDate,omitempty"`
	IssueLocation  *string               `json:"issueLocation,omitempty"`
}

type PersonalDocumentType string

const (
	PersonalDocumentTypeCPF      PersonalDocumentType = "CPF"
	PersonalDocumentTypeCNPJ     PersonalDocumentType = "CNPJ"
	PersonalDocumentTypeRG       PersonalDocumentType = "RG"
	PersonalDocumentTypeCNH      PersonalDocumentType = "CNH"
	PersonalDocumentTypePassport PersonalDocumentType = "PASSPORT"
	PersonalDocumentTypeOthers   PersonalDocumentType = "OUTROS"
	PersonalDocumentTypeNone     PersonalDocumentType = "SEM_OUTROS_DOCUMENTOS"
)

type OtherPersonalDocument struct {
	Type           *string              `json:"type,omitempty"`
	Number         *string              `json:"number,omitempty"`
	Country        *string              `json:"country,omitempty"`
	ExpirationDate *timeutil.BrazilDate `json:"expirationDate,omitempty"`
}

type PersonalContact struct {
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}
