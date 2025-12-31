package customer

import (
	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

var (
	Scope = goidc.NewScope("customers")
)

type PersonalIdentification struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      PersonalIdentificationData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (PersonalIdentification) TableName() string {
	return "customer_personal_identifications"
}

func (p *PersonalIdentification) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type PersonalIdentificationData struct {
	UpdateDateTime          timeutil.DateTime      `json:"updateDateTime"`
	PersonalID              *string                `json:"personalId,omitempty"`
	BrandName               string                 `json:"brandName"`
	CivilName               string                 `json:"civilName"`
	SocialName              *string                `json:"socialName,omitempty"`
	CPF                     string                 `json:"cpfNumber"`
	CompanyInfo             CompanyInfo            `json:"companyInfo"`
	Documents               *[]PersonalDocument    `json:"documents,omitempty"`
	HasBrazilianNationality *bool                  `json:"hasBrazilianNationality,omitempty"`
	OtherDocuments          *OtherPersonalDocument `json:"otherDocuments,omitempty"`
	Contact                 PersonalContact        `json:"contact"`
	CivilStatus             *insurer.CivilStatus   `json:"civilStatus,omitempty"`
	CivilStatusOthers       *string                `json:"civilStatusOthers,omitempty"`
	Sex                     *string                `json:"sex,omitempty"`
	BirthDate               *timeutil.BrazilDate   `json:"birthDate,omitempty"`
	Filiation               *Filiation             `json:"filiation,omitempty"`
	IdentificationDetails   *IdentificationDetails `json:"identificationDetails,omitempty"`
}

type PersonalQualification struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      PersonalQualificationData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (PersonalQualification) TableName() string {
	return "customer_personal_qualifications"
}

func (p *PersonalQualification) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type PersonalQualificationData struct {
	UpdateDateTime    timeutil.DateTime          `json:"updateDateTime"`
	PEPIdentification PEPIdentification          `json:"pepIdentification"`
	Occupations       *[]Occupation              `json:"occupations,omitempty"`
	LifePensionPlans  string                     `json:"lifePensionPlans"`
	InformedRevenue   *PersonalInformedRevenue   `json:"informedRevenue,omitempty"`
	InformedPatrimony *PersonalInformedPatrimony `json:"informedPatrimony,omitempty"`
}

type PersonalComplimentaryInformation struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      PersonalComplimentaryInformationData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (PersonalComplimentaryInformation) TableName() string {
	return "customer_personal_complimentary_informations"
}

func (p *PersonalComplimentaryInformation) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type PersonalComplimentaryInformationData struct {
	UpdateDateTime        timeutil.DateTime     `json:"updateDateTime"`
	StartDate             timeutil.BrazilDate   `json:"startDate"`
	RelationshipBeginning *timeutil.BrazilDate  `json:"relationshipBeginning,omitempty"`
	ProductsServices      []ProductsAndServices `json:"productsServices"`
}

type BusinessIdentification struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      BusinessIdentificationData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (BusinessIdentification) TableName() string {
	return "customer_business_identifications"
}

func (b *BusinessIdentification) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type BusinessIdentificationData struct {
	UpdateDateTime    timeutil.DateTime    `json:"updateDateTime"`
	BusinessID        *string              `json:"businessId,omitempty"`
	BrandName         string               `json:"brandName"`
	CompanyInfo       CompanyInfo          `json:"companyInfo"`
	BusinessName      string               `json:"businessName"`
	BusinessTradeName *string              `json:"businessTradeName,omitempty"`
	IncorporationDate *timeutil.BrazilDate `json:"incorporationDate,omitempty"`
	Document          BusinessDocument     `json:"document"`
	Type              *BusinessType        `json:"type,omitempty"`
	Contact           BusinessContact      `json:"contact"`
	Parties           *[]BusinessParty     `json:"parties,omitempty"`
}

type BusinessQualification struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      BusinessQualificationData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (BusinessQualification) TableName() string {
	return "customer_business_qualifications"
}

func (b *BusinessQualification) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type BusinessQualificationData struct {
	UpdateDateTime    timeutil.DateTime          `json:"updateDateTime"`
	MainBranch        *string                    `json:"mainBranch,omitempty"`
	SecondaryBranch   *string                    `json:"secondaryBranch,omitempty"`
	InformedPatrimony *BusinessInformedPatrimony `json:"informedPatrimony,omitempty"`
	InformedRevenue   *BusinessInformedRevenue   `json:"informedRevenue,omitempty"`
}

type BusinessComplimentaryInformation struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	OwnerID   uuid.UUID
	Data      BusinessComplimentaryInformationData `gorm:"serializer:json"`
	OrgID     string
	CrossOrg  bool
	CreatedAt timeutil.DateTime
	UpdatedAt timeutil.DateTime
}

func (BusinessComplimentaryInformation) TableName() string {
	return "customer_business_complimentary_informations"
}

func (b *BusinessComplimentaryInformation) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type BusinessComplimentaryInformationData struct {
	UpdateDateTime        timeutil.DateTime     `json:"updateDateTime"`
	StartDate             timeutil.BrazilDate   `json:"startDate"`
	RelationshipBeginning *timeutil.BrazilDate  `json:"relationshipBeginning,omitempty"`
	ProductsServices      []ProductsAndServices `json:"productsServices"`
}

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
	PersonalDocumentTypePassport PersonalDocumentType = "PASSAPORTE"
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
	PostalAddresses []PersonalPostalAddress `json:"postalAddresses"`
	Phones          *[]Phone                `json:"phones,omitempty"`
	Emails          *[]Email                `json:"emails,omitempty"`
}

type PersonalPostalAddress struct {
	Address            string                     `json:"address"`
	AdditionalInfo     *string                    `json:"additionalInfo,omitempty"`
	DistrictName       *string                    `json:"districtName,omitempty"`
	TownName           string                     `json:"townName"`
	CountrySubDivision insurer.CountrySubDivision `json:"countrySubDivision"`
	PostCode           string                     `json:"postCode"`
	Country            insurer.CountryCode        `json:"country"`
}

type Phone struct {
	CountryCallingCode *string                `json:"countryCallingCode,omitempty"`
	AreaCode           *insurer.PhoneAreaCode `json:"areaCode,omitempty"`
	Number             *string                `json:"number,omitempty"`
	PhoneExtension     *string                `json:"phoneExtension,omitempty"`
}

type Email struct {
	Email *string `json:"email,omitempty"`
}

type Filiation struct {
	Type      *FiliationType `json:"type,omitempty"`
	CivilName *string        `json:"civilName,omitempty"`
}

type FiliationType string

const (
	FiliationTypeMother      FiliationType = "MAE"
	FiliationTypeFather      FiliationType = "PAI"
	FiliationTypeNoFiliation FiliationType = "SEM_FILIACAO"
)

type IdentificationDetails struct {
	CivilName *string `json:"civilName,omitempty"`
	CpfNumber *string `json:"cpfNumber,omitempty"`
}

type PEPIdentification string

const (
	PEPIdentificationNotExposed                      PEPIdentification = "NAO_EXPOSTO"
	PEPIdentificationPoliticallyExposedPerson        PEPIdentification = "PESSOA_POLITICAMENTE_EXPOSTA_PPE"
	PEPIdentificationPersonCloseToPoliticallyExposed PEPIdentification = "PESSOA_PROXIMA_A_PESSOA_POLITICAMENTE_EXPOSTA_PPEE"
	PEPIdentificationNoInformation                   PEPIdentification = "SEM_INFORMACAO"
)

type Occupation struct {
	Details                  *string             `json:"details,omitempty"`
	OccupationCode           *string             `json:"occupationCode,omitempty"`
	OccupationCodeType       *OccupationCodeType `json:"occupationCodeType,omitempty"`
	OccupationCodeTypeOthers *string             `json:"occupationCodeTypeOthers,omitempty"`
}

type OccupationCodeType string

const (
	OccupationCodeTypeRFB    OccupationCodeType = "RFB"
	OccupationCodeTypeCBO    OccupationCodeType = "CBO"
	OccupationCodeTypeOthers OccupationCodeType = "OUTROS"
)

type PersonalInformedRevenue struct {
	IncomeFrequency *IncomeFrequency     `json:"incomeFrequency,omitempty"`
	Currency        *insurer.Currency    `json:"currency,omitempty"`
	Amount          *string              `json:"amount,omitempty"`
	Date            *timeutil.BrazilDate `json:"date,omitempty"`
}

type BusinessInformedRevenue struct {
	IncomeFrequency *IncomeFrequency  `json:"incomeFrequency,omitempty"`
	Currency        *insurer.Currency `json:"currency,omitempty"`
	Amount          *string           `json:"amount,omitempty"`
	Year            *string           `json:"year,omitempty"`
}

type IncomeFrequency string

const (
	IncomeFrequencyDaily       IncomeFrequency = "DIARIA"
	IncomeFrequencyWeekly      IncomeFrequency = "SEMANAL"
	IncomeFrequencyFortnightly IncomeFrequency = "QUINZENAL"
	IncomeFrequencyMonthly     IncomeFrequency = "MENSAL"
	IncomeFrequencyBimonthly   IncomeFrequency = "BIMESTRAL"
	IncomeFrequencyQuarterly   IncomeFrequency = "TRIMESTRAL"
	IncomeFrequencySemiannual  IncomeFrequency = "SEMESTRAL"
	IncomeFrequencyAnnual      IncomeFrequency = "ANUAL"
)

type PersonalInformedPatrimony struct {
	Currency *insurer.Currency `json:"currency,omitempty"`
	Amount   *string           `json:"amount,omitempty"`
	Year     *string           `json:"year,omitempty"`
}

type BusinessInformedPatrimony struct {
	Currency *insurer.Currency    `json:"currency,omitempty"`
	Amount   *string              `json:"amount,omitempty"`
	Date     *timeutil.BrazilDate `json:"date,omitempty"`
}

type ProductsAndServices struct {
	Contract          string             `json:"contract"`
	Type              ProductServiceType `json:"type"`
	InsuranceLineCode *string            `json:"insuranceLineCode,omitempty"`
	Procurators       *[]Procurator      `json:"procurators,omitempty"`
}

type ProductServiceType string

const (
	ProductServiceTypeMicroinsurance            ProductServiceType = "MICROSSEGUROS"
	ProductServiceTypeCapitalizationTitles      ProductServiceType = "TITULOS_DE_CAPITALIZACAO"
	ProductServiceTypeLifeInsurance             ProductServiceType = "SEGUROS_DE_PESSOAS"
	ProductServiceTypeComplementaryPensionPlans ProductServiceType = "PLANOS_DE_PREVIDENCIA_COMPLEMENTAR"
	ProductServiceTypePropertyCasualtyInsurance ProductServiceType = "SEGUROS_DE_DANOS"
)

type Procurator struct {
	Nature     ProcuratorNature `json:"nature"`
	CpfNumber  *string          `json:"cpfNumber,omitempty"`
	CivilName  *string          `json:"civilName,omitempty"`
	SocialName *string          `json:"socialName,omitempty"`
}

type ProcuratorNature string

const (
	ProcuratorNatureLegalRepresentative ProcuratorNature = "REPRESENTANTE_LEGAL"
	ProcuratorNatureProcurator          ProcuratorNature = "PROCURADOR"
	ProcuratorNatureNotApplicable       ProcuratorNature = "NAO_SE_APLICA"
)

type BusinessDocument struct {
	CNPJNumber                      string               `json:"cnpjNumber"`
	RegistrationNumberOriginCountry *string              `json:"registrationNumberOriginCountry,omitempty"`
	Country                         *insurer.CountryCode `json:"country,omitempty"`
	ExpirationDate                  *timeutil.BrazilDate `json:"expirationDate,omitempty"`
}

type BusinessType string

const (
	BusinessTypePrivate BusinessType = "PRIVADO"
	BusinessTypePublic  BusinessType = "PUBLICO"
)

type BusinessContact struct {
	PostalAddresses []BusinessPostalAddress `json:"postalAddresses"`
	Phones          *[]Phone                `json:"phones,omitempty"`
	Emails          *[]Email                `json:"emails,omitempty"`
}

type BusinessPostalAddress struct {
	Address               string                     `json:"address"`
	AdditionalInfo        *string                    `json:"additionalInfo,omitempty"`
	DistrictName          *string                    `json:"districtName,omitempty"`
	TownName              string                     `json:"townName"`
	CountrySubDivision    insurer.CountrySubDivision `json:"countrySubDivision"`
	PostCode              string                     `json:"postCode"`
	IBGETownCode          *string                    `json:"ibgeTownCode,omitempty"`
	Country               string                     `json:"country"`
	CountryCode           *insurer.CountryCode       `json:"countryCode,omitempty"`
	GeographicCoordinates *GeographicCoordinates     `json:"geographicCoordinates,omitempty"`
}

type GeographicCoordinates struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type BusinessParty struct {
	Type                   *BusinessPartyType   `json:"type,omitempty"`
	CivilName              *string              `json:"civilName,omitempty"`
	SocialName             *string              `json:"socialName,omitempty"`
	StartDate              *timeutil.BrazilDate `json:"startDate,omitempty"`
	Shareholding           *string              `json:"shareholding,omitempty"`
	DocumentType           *string              `json:"documentType,omitempty"`
	DocumentNumber         *string              `json:"documentNumber,omitempty"`
	DocumentCountry        *insurer.CountryCode `json:"documentCountry,omitempty"`
	DocumentExpirationDate *timeutil.BrazilDate `json:"documentExpirationDate,omitempty"`
}

type BusinessPartyType string

const (
	BusinessPartyTypeShareholder   BusinessPartyType = "SOCIO"
	BusinessPartyTypeAdministrator BusinessPartyType = "ADMINISTRADOR"
)

type Filter struct {
	OwnerID string
}
