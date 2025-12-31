package auto

import (
	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/auto"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/quote"
	"github.com/luikyv/mock-insurer/internal/strutil"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

var (
	Scope     = goidc.NewScope("quote-auto")
	ScopeLead = goidc.NewScope("quote-auto-lead")
)

type Quote struct {
	ID              uuid.UUID `gorm:"primaryKey"`
	ConsentID       string
	Status          quote.Status
	StatusUpdatedAt timeutil.DateTime
	Data            Data `gorm:"serializer:json"`
	OrgID           string
	CreatedAt       timeutil.DateTime
	UpdatedAt       timeutil.DateTime
}

func (Quote) TableName() string {
	return "insurance_auto_quotes"
}

func (p *Quote) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (q *Quote) GetID() uuid.UUID {
	return q.ID
}

func (q *Quote) GetStatus() quote.Status {
	return q.Status
}

func (q *Quote) SetStatus(status quote.Status) {
	q.Status = status
}

func (q *Quote) SetStatusUpdatedAt(updatedAt timeutil.DateTime) {
	q.StatusUpdatedAt = updatedAt
}

func (q *Quote) SetUpdatedAt(updatedAt timeutil.DateTime) {
	q.UpdatedAt = updatedAt
}

func (q *Quote) SetCreatedAt(createdAt timeutil.DateTime) {
	q.CreatedAt = createdAt

}

func (q *Quote) GetTermStartDate() timeutil.BrazilDate {
	return q.Data.TermStartDate
}

func (q *Quote) GetTermEndDate() timeutil.BrazilDate {
	return q.Data.TermEndDate
}

func (q *Quote) SetRejectionReason(rejectionReason string) {
	q.Data.RejectionReason = &rejectionReason
}

func (q *Quote) SetInsurerQuoteID(insurerQuoteID string) {
	q.Data.InsurerQuoteID = &insurerQuoteID
}

func (q *Quote) SetProtocolDateTime(protocolDateTime timeutil.DateTime) {
	q.Data.ProtocolDateTime = &protocolDateTime
}

func (q *Quote) SetProtocolNumber(protocolNumber string) {
	q.Data.ProtocolNumber = &protocolNumber
}

func (q *Quote) SetRedirectLink(redirectLink string) {
	q.Data.RedirectLink = &redirectLink
}

func (q *Quote) GetPersonalIdentification() *string {
	if q.Data.Customer.Personal == nil {
		return nil
	}
	return &q.Data.Customer.Personal.Identification.CPF
}

func (q *Quote) GetBusinessIdentification() *string {
	if q.Data.Customer.Business == nil {
		return nil
	}
	return &q.Data.Customer.Business.Identification.CompanyInfo.CNPJ
}

func (q *Quote) GetOfferIDs() []string {
	if q.Data.Quotes == nil {
		return nil
	}

	offerIDs := make([]string, 0, len(*q.Data.Quotes))
	for _, o := range *q.Data.Quotes {
		offerIDs = append(offerIDs, o.InsurerQuoteID)
	}
	return offerIDs
}

func (q *Quote) CreateOffers() {
	q.Data.Quotes = &[]Offer{
		{
			InsurerQuoteID:      uuid.New().String(),
			SusepProcessNumbers: []string{strutil.Random(50)},
			Premium: Premium{
				PaymentsQuantity: "1",
				TotalAmount: insurer.AmountDetails{
					Amount:   "100.00",
					UnitType: insurer.UnitTypeMonetary,
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
					},
				},
				TotalNetAmount: insurer.AmountDetails{
					Amount:   "100.00",
					UnitType: insurer.UnitTypeMonetary,
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
					},
				},
				IOF: insurer.AmountDetails{
					Amount:   "100.00",
					UnitType: insurer.UnitTypeMonetary,
					Unit: &insurer.Unit{
						Code:        insurer.UnitCodeReal,
						Description: insurer.CurrencyBRL,
					},
				},
			},
			Coverages:   []OfferCoverage{},
			Assistances: []Assistance{},
		},
	}
}

func (q *Quote) GetOrgID() string {
	return q.OrgID
}

type Data struct {
	InsurerQuoteID             *string              `json:"insurerQuoteId,omitempty"`
	ProtocolDateTime           *timeutil.DateTime   `json:"protocolDateTime,omitempty"`
	ProtocolNumber             *string              `json:"protocolNumber,omitempty"`
	RedirectLink               *string              `json:"redirectLink,omitempty"`
	RejectionReason            *string              `json:"rejectionReason,omitempty"`
	ExpirationDateTime         timeutil.DateTime    `json:"expirationDateTime"`
	Customer                   quote.Customer       `json:"customer"`
	IsCollectiveStipulated     bool                 `json:"isCollectiveStipulated"`
	HasAnIndividualItem        bool                 `json:"hasAnIndividualItem"`
	TermStartDate              timeutil.BrazilDate  `json:"termStartDate"`
	TermEndDate                timeutil.BrazilDate  `json:"termEndDate"`
	TermType                   insurer.ValidityType `json:"termType"`
	InsuranceType              InsuranceType        `json:"insuranceType"`
	PolicyID                   *string              `json:"policyId,omitempty"`
	InsurerID                  *string              `json:"insurerId,omitempty"`
	IdentifierCode             *string              `json:"identifierCode,omitempty"`
	BonusClass                 *string              `json:"bonusClass,omitempty"`
	Currency                   insurer.Currency     `json:"currency"`
	IncludesAssistanceServices bool                 `json:"includesAssistanceServices"`
	InsuredObject              *InsuredObject       `json:"insuredObject,omitempty"`
	Coverages                  *[]Coverage          `json:"coverages,omitempty"`
	CustomData                 *quote.CustomData    `json:"customData,omitempty"`
	HistoricalData             *HistoricalData      `json:"historicalData,omitempty"`
	Quotes                     *[]Offer             `json:"quotes,omitempty"`
}

type Lead struct {
	ID              uuid.UUID `gorm:"primaryKey"`
	ConsentID       string
	Status          quote.Status
	StatusUpdatedAt timeutil.DateTime
	Data            LeadData `gorm:"serializer:json"`
	OrgID           string
	CreatedAt       timeutil.DateTime
	UpdatedAt       timeutil.DateTime
}

func (Lead) TableName() string {
	return "insurance_auto_quote_leads"
}

func (l *Lead) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

func (l *Lead) GetID() uuid.UUID {
	return l.ID
}

func (l *Lead) GetStatus() quote.Status {
	return l.Status
}

func (l *Lead) SetStatus(status quote.Status) {
	l.Status = status
}

func (l *Lead) SetStatusUpdatedAt(updatedAt timeutil.DateTime) {
	l.StatusUpdatedAt = updatedAt
}

func (l *Lead) SetUpdatedAt(updatedAt timeutil.DateTime) {
	l.UpdatedAt = updatedAt
}

func (l *Lead) SetCreatedAt(createdAt timeutil.DateTime) {
	l.CreatedAt = createdAt
}

func (l *Lead) GetOrgID() string {
	return l.OrgID
}

type LeadData struct {
	ExpirationDateTime timeutil.DateTime `json:"expirationDateTime"`
	Customer           quote.Customer    `json:"customer"`
	HistoricalData     *HistoricalData   `json:"historicalData,omitempty"`
	Coverages          []LeadCoverage    `json:"coverages"`
}

type InsuranceType string

const (
	InsuranceTypeNew     InsuranceType = "NOVO"
	InsuranceTypeRenewal InsuranceType = "RENOVACAO"
)

type InsuredObject struct {
	Identification                     string                      `json:"identification"`
	Model                              *InsuredObjectModel         `json:"model,omitempty"`
	Modality                           *auto.InsuredObjectModality `json:"modality,omitempty"`
	TableUsed                          *auto.AmountReferenceTable  `json:"tableUsed,omitempty"`
	ModelCode                          *string                     `json:"modelCode,omitempty"`
	AdjustmentFactor                   *string                     `json:"adjustmentFactor,omitempty"`
	ValuedDetermined                   *string                     `json:"valuedDetermined,omitempty"`
	Tax                                *Tax                        `json:"tax,omitempty"`
	DoorsNumber                        string                      `json:"doorsNumber"`
	Color                              string                      `json:"color"`
	LicensePlate                       *string                     `json:"licensePlate,omitempty"`
	VehicleUsage                       []auto.VehicleUsage         `json:"vehicleUsage"`
	CommercialActivityType             *[]CommercialActivityType   `json:"commercialActivityType,omitempty"`
	RiskManagementSystem               *[]RiskManagementSystem     `json:"riskManagementSystem,omitempty"`
	IsTransportCargoInsurance          *bool                       `json:"isTransportCargoInsurance,omitempty"`
	LoadsCarriedInsured                *[]LoadType                 `json:"loadsCarriedInsured,omitempty"`
	IsEquipmentAttached                bool                        `json:"isEquipmentAttached"`
	EquipmentsAttached                 *EquipmentAttached          `json:"equipmentsAttached,omitempty"`
	Chasis                             *string                     `json:"chasis,omitempty"`
	IsAuctionChassisRescheduled        *bool                       `json:"isAuctionChassisRescheduled,omitempty"`
	IsBrandNew                         bool                        `json:"isBrandNew"`
	DepartureDateFromCarDealership     *timeutil.BrazilDate        `json:"departureDateFromCarDealership,omitempty"`
	VehicleInvoice                     *VehicleInvoice             `json:"vehicleInvoice,omitempty"`
	Fuel                               Fuel                        `json:"fuel"`
	IsGasKit                           bool                        `json:"isGasKit"`
	GasKit                             *GasKit                     `json:"gasKit,omitempty"`
	IsArmouredVehicle                  bool                        `json:"isArmouredVehicle"`
	ArmouredVehicle                    *ArmouredVehicle            `json:"armouredVehicle,omitempty"`
	IsActiveTrackingVehicle            *bool                       `json:"isActiveTrackingVehicle,omitempty"`
	FrequentTrafficArea                *TrafficArea                `json:"frequentTrafficArea,omitempty"`
	OvernightPostCode                  string                      `json:"overnightPostCode"`
	RiskLocation                       *RiskLocation               `json:"riskLocation,omitempty"`
	IsExtendCoverageAgedBetween18And25 *bool                       `json:"isExtendCoverageAgedBetween18And25,omitempty"`
	DriverBetween18And25YearsOldGender *DriverGender               `json:"driverBetween18and25YearsOldGender,omitempty"`
	WasThereAClaim                     *bool                       `json:"wasThereAClaim,omitempty"`
	ClaimNotifications                 *[]ClaimNotification        `json:"claimNotifications,omitempty"`
	LicensePlateType                   []LicensePlateType          `json:"licensePlateType"`
	Tariff                             *Tariff                     `json:"tariff,omitempty"`
}

type InsuredObjectModel struct {
	Brand           string  `json:"brand"`
	ModelName       string  `json:"modelName"`
	ModelYear       *string `json:"modelYear,omitempty"`
	ManufactureYear *string `json:"manufactureYear,omitempty"`
}

type ArmouredVehicle struct {
	Amount           *insurer.AmountDetails `json:"amount,omitempty"`
	IsDesireCoverage *bool                  `json:"isDesireCoverage,omitempty"`
}

type Tax struct {
	Exempt              bool     `json:"exempt"`
	Type                *TaxType `json:"type,omitempty"`
	ExemptionPercentage *string  `json:"exemptionPercentage,omitempty"`
}

type TaxType string

const (
	TaxTypeICMS TaxType = "ICMS"
	TaxTypeIPI  TaxType = "IPI"
	TaxTypeBoth TaxType = "AMBOS"
)

type CommercialActivityType string

const (
	CommercialActivityTypeProfessionalCommercialRepresentationSalesPromotersAndServiceProviders CommercialActivityType = "COMERCIAL_ATIVIDADE_PROFISSIONAL_PARA_REPRESENTACAO_COMERCIAL_VENDEDORES_PROMOTORES_E_PRESTADORES_DE_SERVICOS"
	CommercialActivityTypePrivateTaxi                                                           CommercialActivityType = "TAXI_PARTICULAR"
	CommercialActivityTypeRideShareAppDriver                                                    CommercialActivityType = "MOTORISTA_DE_APLICATIVO_APLICATIVO_DE_TRANSPORTE"
	CommercialActivityTypeMotorcycleTaxi                                                        CommercialActivityType = "MOTO_TAXI"
	CommercialActivityTypeSharedTransport                                                       CommercialActivityType = "LOTACAO"
	CommercialActivityTypeSchoolTransport                                                       CommercialActivityType = "TRANSPORTE_ESCOLAR"
	CommercialActivityTypeRentalContract                                                        CommercialActivityType = "LOCADORA_CONTRATO"
	CommercialActivityTypeOccasionalRental                                                      CommercialActivityType = "LOCADORA_AVULSO"
	CommercialActivityTypeFreightTransport                                                      CommercialActivityType = "TRANSPORTE_DE_MERCADORIA"
	CommercialActivityTypeTransportCompanyService                                               CommercialActivityType = "PRESTA_SERVICO_PARA_TRANSPORTADORA"
	CommercialActivityTypeUrbanPassengerTransport                                               CommercialActivityType = "TRANSPORTE_DE_PESSOAS_URBANO"
	CommercialActivityTypeContinuousBusinessCharterPassengerTransport                           CommercialActivityType = "TRANSPORTE_DE_PESSOAS_FRETAMENTO_EMPRESARIAL_CONTINUO"
	CommercialActivityTypeMixedCharterPassengerTransport                                        CommercialActivityType = "TRANSPORTE_DE_PESSOAS_FRETAMENTO_MISTO_FRETE_PESSOAS"
	CommercialActivityTypeTouristCharterPassengerTransport                                      CommercialActivityType = "TRANSPORTE_DE_PESSOAS_FRETAMENTO_TURISTICO"
	CommercialActivityTypeOfficialPublicAgencyVehicles                                          CommercialActivityType = "VEICULOS_OFICIAIS_ORGAO_PUBLICO"
	CommercialActivityTypeAmbulance                                                             CommercialActivityType = "AMBULANCIA"
	CommercialActivityTypeFirefighters                                                          CommercialActivityType = "BOMBEIROS"
	CommercialActivityTypeGarbageCollectors                                                     CommercialActivityType = "COLETORES_DE_LIXO"
	CommercialActivityTypeSecurity                                                              CommercialActivityType = "VIGILANCIA"
	CommercialActivityTypePolicing                                                              CommercialActivityType = "POLICIAMENTO"
	CommercialActivityTypeCompetitionEvents                                                     CommercialActivityType = "COMPETICAO_EVENTOS"
	CommercialActivityTypeDrivingSchools                                                        CommercialActivityType = "AUTO_ESCOLAS"
	CommercialActivityTypeTestDrive                                                             CommercialActivityType = "TEST_DRIVE"
	CommercialActivityTypeSpecializedTrailerMotorhomeMobileHospitalsElectricRepairPlatformEtc   CommercialActivityType = "DIFERENCIADOS_EX_TRAILER_MOTORHOME_HOSPITAIS_VOLANTE_VEICULOS_COM_PLATAFORMA_PARA_REPAROS_DE_ENERGIA_ELETRICA_ETC"
	CommercialActivityTypeOthers                                                                CommercialActivityType = "OUTROS"
)

type RiskManagementSystem string

const (
	RiskManagementSystemNo             RiskManagementSystem = "NAO"
	RiskManagementSystemDriverRegistry RiskManagementSystem = "CADASTRO_DE_MOTORISTAS"
	RiskManagementSystemCargoEscort    RiskManagementSystem = "ESCOLTA_DE_CARGAS"
	RiskManagementSystemOthers         RiskManagementSystem = "OUTROS"
	RiskManagementSystemNotInformed    RiskManagementSystem = "NAO_INFORMADO"
)

type LoadType string

const (
	LoadTypeAutoParts LoadType = "AUTO_PECAS"
	LoadTypeVehicles  LoadType = "AUTOMOVEIS"
	LoadTypeBeverages LoadType = "BEBIDAS"
	LoadTypeToys      LoadType = "BRINQUEDOS"
	LoadTypeShoes     LoadType = "CALCADOS"
	LoadTypeMixedLoad LoadType = "CARGA_MISTA"
	LoadTypeLiveLoad  LoadType = "CARGA_VIVA"
	LoadTypeTobacco   LoadType = "CIGARROS"
	LoadTypeFlammable LoadType = "COMBUSTIVEIS_OU_INFLAMAVEIS"
	LoadTypeClothing  LoadType = "CONFECCOES"
)

type EquipmentAttached struct {
	Amount           *insurer.AmountDetails `json:"amount,omitempty"`
	IsDesireCoverage *bool                  `json:"isDesireCoverage,omitempty"`
}

type VehicleInvoice struct {
	Amount *insurer.AmountDetails `json:"amount,omitempty"`
	Number *string                `json:"number,omitempty"`
}

type Fuel string

const (
	FuelGasoline   Fuel = "GASOLINA"
	FuelFlex       Fuel = "FLEX"
	FuelHybrid     Fuel = "HIBRIDO"
	FuelElectric   Fuel = "ELETRICO"
	FuelDiesel     Fuel = "DIESEL"
	FuelNaturalGas Fuel = "GAS_GNV"
	FuelEthanol    Fuel = "ALCOOL_ETANOL"
	FuelFlexAndGNV Fuel = "FLEX_E_GNV"
)

type GasKit struct {
	IsDesireCoverage *bool                  `json:"isDesireCoverage,omitempty"`
	Amount           *insurer.AmountDetails `json:"amount,omitempty"`
}

type TrafficArea string

const (
	TrafficAreaMunicipalityAndSurroundingsUpToKM TrafficArea = "MUNICIPIO_E_ARREDORES_ATE_KM"
	TrafficAreaWithinOwnState                    TrafficArea = "DENTRO_DO_PROPRIO_ESTADO_DA_SEDE"
	TrafficAreaNorthRegion                       TrafficArea = "REGIAO_NORTE"
	TrafficAreaNortheastRegion                   TrafficArea = "REGIAO_NORDESTE"
	TrafficAreaCentralWestRegion                 TrafficArea = "REGIAO_CENTRO_OESTE"
	TrafficAreaSouthRegion                       TrafficArea = "REGIAO_SUL"
	TrafficAreaSoutheastRegion                   TrafficArea = "REGIAO_SUDESTE"
	TrafficAreaMercosur                          TrafficArea = "MERCOSUL"
	TrafficAreaSouthAmerica                      TrafficArea = "AMERICA_DO_SUL"
	TrafficAreaNotInformed                       TrafficArea = "NAO_INFORMADO"
)

type RiskLocation struct {
	IsUsedCollege     *bool                `json:"isUsedCollege,omitempty"`
	UsedCollege       *RiskLocationUsage   `json:"usedCollege,omitempty"`
	IsUsedCommuteWork *bool                `json:"isUsedCommuteWork,omitempty"`
	UsedCommuteWork   *RiskLocationUsage   `json:"usedCommuteWork,omitempty"`
	KmAveragePerWeek  *string              `json:"kmAveragePerWeek,omitempty"`
	Housing           *RiskLocationHousing `json:"housing,omitempty"`
}

type RiskLocationHousing struct {
	Type           *HousingType `json:"type,omitempty"`
	IsKeptInGarage *bool        `json:"isKeptInGarage,omitempty"`
	GateType       *GateType    `json:"gateType,omitempty"`
}

type RiskLocationUsage struct {
	IsKeptInGarage        *bool   `json:"isKeptInGarage,omitempty"`
	DistanceFromResidence *string `json:"distanceFromResidence,omitempty"`
}

type HousingType string

const (
	HousingTypeApartment   HousingType = "APARTAMENTO_OU_FLAT"
	HousingTypeHouse       HousingType = "CASA_OU_SOBRADO"
	HousingTypeCondominium HousingType = "CASA_OU_SOBRADO_EM_CONDOMINIO_FECHADO"
	HousingTypeFarm        HousingType = "CHACARA_FAZENDA_OU_SITIO"
)

type GateType string

const (
	GateTypeAutomatic GateType = "PORTAO_AUTOMATICO"
	GateTypeManual    GateType = "PORTAO_MANUAL"
)

type DriverGender string

const (
	DriverGenderMale               DriverGender = "MASCULINO"
	DriverGenderFemale             DriverGender = "FEMININO"
	DriverGenderMaleAndFemale      DriverGender = "MASCULINO_E_FEMININO"
	DriverGenderPreferNotToDeclare DriverGender = "PREFIRO_NAO_DECLARAR"
)

type Beneficiary struct {
	Identification         *string                     `json:"identification,omitempty"`
	IdentificationType     *insurer.IdentificationType `json:"identificationType,omitempty"`
	HousingVehiclesNumber  *string                     `json:"housingVehiclesNumber,omitempty"`
	IsUndeterminedDriver   *bool                       `json:"isUndeterminedDriver,omitempty"`
	IsInsuredTheOwner      *bool                       `json:"isInsuredTheOwner,omitempty"`
	RelationshipMainDriver *Relationship               `json:"relationshipMainDriver,omitempty"`
	IsInsuredTheMainDriver *bool                       `json:"isInsuredTheMainDriver,omitempty"`
	MainDriver             *MainDriver                 `json:"mainDriver,omitempty"`
}

type Relationship string

const (
	RelationshipSonOrStepchild  Relationship = "FILHO_A_ENTEADO_A"
	RelationshipSpouseOrPartner Relationship = "CONJUGE_COMPANHEIRO_A"
	RelationshipParent          Relationship = "PAI_MAE"
	RelationshipSibling         Relationship = "IRMAO_IRMA"
	RelationshipOthers          Relationship = "OUTROS"
	RelationshipNotApplicable   Relationship = "NAO_SE_APLICA"
)

type MainDriver struct {
	Identification     *string                     `json:"identification,omitempty"`
	IdentificationType *insurer.IdentificationType `json:"identificationType,omitempty"`
	Name               *string                     `json:"name,omitempty"`
	BirthDate          *timeutil.BrazilDate        `json:"birthDate,omitempty"`
	Sex                *Gender                     `json:"sex,omitempty"`
	CivilStatus        *insurer.CivilStatus        `json:"civilStatus,omitempty"`
	Gender             *Gender                     `json:"gender,omitempty"`
	PostCode           *string                     `json:"postCode,omitempty"`
	LicensedExperience *string                     `json:"licensedExperience,omitempty"`
	Profession         *string                     `json:"profession,omitempty"`
}

type Gender string

const (
	GenderMale   Gender = "MASCULINO"
	GenderFemale Gender = "FEMININO"
)

type ClaimNotification struct {
	ClaimAmount      *insurer.AmountDetails `json:"claimAmount,omitempty"`
	ClaimDescription *string                `json:"claimDescription,omitempty"`
}

type Coverage struct {
	Branch                       string                `json:"branch"`
	Code                         auto.CoverageCode     `json:"code"`
	Description                  *string               `json:"description,omitempty"`
	IsSeparateContractingAllowed bool                  `json:"isSeparateContractingAllowed"`
	MaxLMI                       insurer.AmountDetails `json:"maxLMI"`
	InternalCode                 *string               `json:"internalCode,omitempty"`
}

type HistoricalData struct {
	Customer *quote.Customer `json:"customer,omitempty"`
	Policies *[]Policy       `json:"policies,omitempty"`
}

type Policy struct {
	Info  auto.PolicyData `json:"info"`
	Claim auto.ClaimData  `json:"claim"`
}

type LeadCoverage struct {
	Branch string            `json:"branch"`
	Code   auto.CoverageCode `json:"code"`
}

type Offer struct {
	InsurerQuoteID      string          `json:"insurerQuoteId"`
	SusepProcessNumbers []string        `json:"susepProcessNumbers"`
	Coverages           []OfferCoverage `json:"coverages"`
	Assistances         []Assistance    `json:"assistances"`
	Premium             Premium         `json:"premium"`
}

type OfferCoverage struct {
	Branch                       string                        `json:"branch"`
	Code                         auto.CoverageCode             `json:"code"`
	Description                  *string                       `json:"description,omitempty"`
	InternalCode                 *string                       `json:"internalCode,omitempty"`
	IsSeparateContractingAllowed bool                          `json:"isSeparateContractingAllowed"`
	GracePeriod                  *int                          `json:"gracePeriod,omitempty"`
	GracePeriodicity             *insurer.Periodicity          `json:"gracePeriodicity,omitempty"`
	GracePeriodCountingMethod    *insurer.PeriodCountingMethod `json:"gracePeriodCountingMethod,omitempty"`
	GracePeriodStartDate         *timeutil.BrazilDate          `json:"gracePeriodStartDate,omitempty"`
	GracePeriodEndDate           *timeutil.BrazilDate          `json:"gracePeriodEndDate,omitempty"`
	Deductible                   *auto.CoverageDeductible      `json:"deductible,omitempty"`
	POS                          auto.CoveragePOS              `json:"POS"`
	FullIndemnity                insurer.ValueType             `json:"fullIndemnity"`
}

type Assistance struct {
	Type          AssistanceType         `json:"type"`
	Service       AssistanceService      `json:"service"`
	Description   string                 `json:"description"`
	PremiumAmount *insurer.AmountDetails `json:"premiumAmount,omitempty"`
}

type AssistanceType string

const (
	AssistanceTypeAuto                 AssistanceType = "ASSISTENCIA_AUTO"
	AssistanceTypeRE                   AssistanceType = "ASSISTENCIA_RE"
	AssistanceTypeLife                 AssistanceType = "ASSISTENCIA_VIDA"
	AssistanceTypeBenefits             AssistanceType = "BENEFICIOS"
	AssistanceTypeDispatcher           AssistanceType = "DESPACHANTE"
	AssistanceTypeVehicleRental        AssistanceType = "LOCACAO_DE_VEICULOS"
	AssistanceTypeAutomotiveRepairs    AssistanceType = "REPAROS_AUTOMOTIVOS"
	AssistanceTypeEmergencyRepairs     AssistanceType = "REPAROS_EMERGENCIAIS"
	AssistanceTypeMaintenanceService   AssistanceType = "SERVICO_DE_MANUTENCAO"
	AssistanceTypeServiceInCaseOfClaim AssistanceType = "SERVICO_EM_CASO_DE_SINISTRO"
	AssistanceTypeEmergencyTransport   AssistanceType = "TRANSPORTE_DO_EMERGENCIAL"
	AssistanceTypeOthers               AssistanceType = "OUTROS"
)

type AssistanceService string

const (
	AssistanceServiceActivationAndOrSchedulingOfPickupAndDelivery             AssistanceService = "ACIONAMENTO_E_OU_AGENDAMENTO_DE_LEVA_E_TRAZ"
	AssistanceServiceChildCare                                                AssistanceService = "AMPARO_DE_CRIANCAS"
	AssistanceServiceHomeVaccination                                          AssistanceService = "APLICACAO_DE_VACINAS_EM_DOMICILIO"
	AssistanceServiceHeaters                                                  AssistanceService = "AQUECEDORES"
	AssistanceServiceApplianceAssistance                                      AssistanceService = "ASSISTENCIA_A_ELETRODOMESTICOS"
	AssistanceServiceAutoAndOrMotorcycleAssistance                            AssistanceService = "ASSISTENCIA_AUTO_E_OU_MOTO"
	AssistanceServiceBikeAssistance                                           AssistanceService = "ASSISTENCIA_BIKE"
	AssistanceServiceTravelAssistance                                         AssistanceService = "ASSISTENCIA_EM_VIAGEM"
	AssistanceServiceSchoolAssistance                                         AssistanceService = "ASSISTENCIA_ESCOLAR"
	AssistanceServiceFuneralAssistance                                        AssistanceService = "ASSISTENCIA_FUNERAL"
	AssistanceServicePetFuneralAssistance                                     AssistanceService = "ASSISTENCIA_FUNERAL_PET"
	AssistanceServiceITAssistance                                             AssistanceService = "ASSISTENCIA_INFORMATICA"
	AssistanceServiceNutritionalAssistance                                    AssistanceService = "ASSISTENCIA_NUTRICIONAL"
	AssistanceServicePetAssistance                                            AssistanceService = "ASSISTENCIA_PET"
	AssistanceServiceResidentialAssistance                                    AssistanceService = "ASSISTENCIA_RESIDENCIAL"
	AssistanceServiceSustainableAssistance                                    AssistanceService = "ASSISTENCIA_SUSTENTAVEL"
	AssistanceServiceEmergencyVeterinaryAssistance                            AssistanceService = "ASSISTENCIA_VETERINARIA_EMERGENCIAL"
	AssistanceServiceHealthAndWellnessAssistance                              AssistanceService = "ASSISTENCIAS_SAUDE_E_BEM_ESTAR"
	AssistanceServiceBabySitter                                               AssistanceService = "BABY_SITTER"
	AssistanceServiceDumpster                                                 AssistanceService = "CACAMBA"
	AssistanceServiceReserveCar                                               AssistanceService = "CARRO_RESERVA"
	AssistanceServiceBasicBasket                                              AssistanceService = "CESTA_BASICA"
	AssistanceServiceFoodBasket                                               AssistanceService = "CESTA_DE_ALIMENTOS"
	AssistanceServiceBirthBasket                                              AssistanceService = "CESTA_NATALIDADE"
	AssistanceServiceLocksmith                                                AssistanceService = "CHAVEIRO"
	AssistanceServiceCheckUp                                                  AssistanceService = "CHECK_UP"
	AssistanceServiceProvisionalRoofCoverage                                  AssistanceService = "COBERTURA_PROVISORIA_DE_TELHADO"
	AssistanceServiceConcierge                                                AssistanceService = "CONCIERGE"
	AssistanceServiceAirConditioningRepair                                    AssistanceService = "CONSERTO_DE_AR_CONDICIONADO"
	AssistanceServiceWhiteGoodsRepair                                         AssistanceService = "CONSERTO_DE_ELETRODOMESTICOS_LINHA_BRANCA"
	AssistanceServiceBrownGoodsRepair                                         AssistanceService = "CONSERTO_DE_ELETROELETRONICO_LINHA_MARROM"
	AssistanceServiceCorrugatedDoorRepair                                     AssistanceService = "CONSERTO_DE_PORTA_ONDULADA"
	AssistanceServiceVeterinaryConsultations                                  AssistanceService = "CONSULTAS_VETERINARIAS"
	AssistanceServiceBudgetConsulting                                         AssistanceService = "CONSULTORIA_ORCAMENTARIA"
	AssistanceServiceTravelConvenience                                        AssistanceService = "CONVENIENCIA_EM_VIAGEM"
	AssistanceServicePestControl                                              AssistanceService = "DEDETIZACAO"
	AssistanceServiceUnstuck                                                  AssistanceService = "DESATOLAMENTO"
	AssistanceServiceResponsibleDisposal                                      AssistanceService = "DESCARTE_RESPONSAVEL"
	AssistanceServiceDiscountsOnConsultationsAndExams                         AssistanceService = "DESCONTOS_EM_CONSULTAS_E_EXAMES"
	AssistanceServiceDiscountsOnMedications                                   AssistanceService = "DESCONTOS_EM_MEDICAMENTOS"
	AssistanceServiceUnclogging                                               AssistanceService = "DESENTUPIMENTO"
	AssistanceServiceDisinsectizationAndDeratization                          AssistanceService = "DESINSETIZACAO_E_DESRATIZACAO"
	AssistanceServiceDispatcher                                               AssistanceService = "DESPACHANTE"
	AssistanceServicePharmaceuticalExpenses                                   AssistanceService = "DESPESAS_FARMACEUTICAS"
	AssistanceServiceMedicalSurgicalAndHospitalizationExpenses                AssistanceService = "DESPESAS_MEDICAS_CIRURGICAS_E_DE_HOSPITALIZACAO"
	AssistanceServiceDentalExpenses                                           AssistanceService = "DESPESAS_ODONTOLOGICAS"
	AssistanceServiceElectrician                                              AssistanceService = "ELETRICISTA"
	AssistanceServiceEmergencies                                              AssistanceService = "EMERGENCIAS"
	AssistanceServicePlumber                                                  AssistanceService = "ENCANADOR"
	AssistanceServiceSendingCompanionInCaseOfAccident                         AssistanceService = "ENVIO_DE_ACOMPANHANTE_EM_CASO_DE_ACIDENTE"
	AssistanceServiceSendingFamilyMemberForAccompanimentOfMinorsUnderFourteen AssistanceService = "ENVIO_DE_FAMILIAR_PARA_ACOMPANHAMENTO_DE_MENORES_DE_CATORZE_ANOS"
	AssistanceServicePetFoodDelivery                                          AssistanceService = "ENVIO_DE_RACAO"
	AssistanceServiceVirtualOffice                                            AssistanceService = "ESCRITORIO_VIRTUAL"
	AssistanceServicePetBoarding                                              AssistanceService = "GUARDA_DE_ANIMAIS"
	AssistanceServiceVehicleStorage                                           AssistanceService = "GUARDA_DO_VEICULO"
	AssistanceServiceTowTruck                                                 AssistanceService = "GUINCHO"
	AssistanceServiceHelpDesk                                                 AssistanceService = "HELP_DESK"
	AssistanceServiceHydraulics                                               AssistanceService = "HIDRAULICA"
	AssistanceServiceAccommodation                                            AssistanceService = "HOSPEDAGEM"
	AssistanceServicePetAccommodation                                         AssistanceService = "HOSPEDAGEM_DE_ANIMAIS"
	AssistanceServiceBathAndGroomingReferral                                  AssistanceService = "INDICACAO_DE_BANHO_E_TOSA"
	AssistanceServiceProfessionalReferral                                     AssistanceService = "INDICACAO_DE_PROFISSIONAIS"
	AssistanceServiceDogBreedInformation                                      AssistanceService = "INFORMACAO_SOBRE_RACAS_DE_CAES"
	AssistanceServicePuppySaleInformation                                     AssistanceService = "INFORMACAO_SOBRE_VENDA_DE_FILHOTES"
	AssistanceServiceVaccineInformation                                       AssistanceService = "INFORMACOES_SOBRE_VACINAS"
	AssistanceServiceUsefulVeterinaryInformation                              AssistanceService = "INFORMACOES_VETERINARIAS_UTEIS"
	AssistanceServiceResidentialInstallation                                  AssistanceService = "INSTALACAO_RESIDENCIA"
	AssistanceServiceElectricShowerInstallationAndOrResistanceReplacement     AssistanceService = "INSTALACAO_DE_CHUVEIRO_ELETRICO_E_OU_TROCA_DE_RESISTENCIA"
	AssistanceServiceTVBracketInstallationUpToSeventy                         AssistanceService = "INSTALACAO_DE_SUPORTE_TV_ATE_SETENTA"
	AssistanceServiceCleaning                                                 AssistanceService = "LIMPEZA"
	AssistanceServiceAirConditioningCleaning                                  AssistanceService = "LIMPEZA_DE_AR_CONDICIONADO"
	AssistanceServiceWaterTankCleaning                                        AssistanceService = "LIMPEZA_DE_CAIXA_D_AGUA"
	AssistanceServiceGutterCleaning                                           AssistanceService = "LIMPEZA_DE_CALHAS"
	AssistanceServiceDrainAndTrapCleaning                                     AssistanceService = "LIMPEZA_DE_RALOS_E_SIFOES"
	AssistanceServiceApplianceRental                                          AssistanceService = "LOCACAO_DE_ELETRODOMESTICOS"
	AssistanceServiceVehicleRental                                            AssistanceService = "LOCACAO_DE_VEICULOS"
	AssistanceServiceBaggageLocation                                          AssistanceService = "LOCALIZACAO_DE_BAGAGEM"
	AssistanceServiceMaintenance                                              AssistanceService = "MANUTENCAO"
	AssistanceServiceDentRepairAndQuickRepair                                 AssistanceService = "MARTELINHO_E_REPARO_RAPIDO"
	AssistanceServiceMechanic                                                 AssistanceService = "MECANICO"
	AssistanceServiceTransport                                                AssistanceService = "MEIO_DE_TRANSPORTE"
	AssistanceServiceMedicalMonitoring                                        AssistanceService = "MONITORACAO_MEDICA"
	AssistanceServiceMotorcycle                                               AssistanceService = "MOTO"
	AssistanceServiceFriendDriver                                             AssistanceService = "MOTORISTA_AMIGO"
	AssistanceServiceSubstituteDriver                                         AssistanceService = "MOTORISTA_SUBSTITUTO"
	AssistanceServiceMTAAlternativeTransport                                  AssistanceService = "MTA_MEIO_DE_TRANSPORTE_ALTERNATIVO"
	AssistanceServiceMovingAndFurnitureStorage                                AssistanceService = "MUDANCA_E_GUARDA_DE_MOVEIS"
	AssistanceServiceOrganization                                             AssistanceService = "ORGANIZACAO"
	AssistanceServiceGuidanceInCaseOfDocumentLoss                             AssistanceService = "ORIENTACAO_EM_CASO_DE_PERDA_DE_DOCUMENTOS"
	AssistanceServiceMedicalGuidance                                          AssistanceService = "ORIENTACAO_MEDICA"
	AssistanceServicePsychologicalGuidance                                    AssistanceService = "ORIENTACAO_PSICOLOGICA"
	AssistanceServicePersonalFitness                                          AssistanceService = "PERSONAL_FITNESS"
	AssistanceServiceTowing                                                   AssistanceService = "REBOQUE"
	AssistanceServiceBikeTowing                                               AssistanceService = "REBOQUE_BIKE"
	AssistanceServiceVehicleRecovery                                          AssistanceService = "RECUPERACAO_DO_VEICULO"
	AssistanceServiceEarlyReturnInCaseOfFamilyDeath                           AssistanceService = "REGRESSO_ANTECIPADO_EM_CASO_DE_FALECIMENTO_DE_PARENTES"
	AssistanceServiceUserReturnAfterHospitalDischarge                         AssistanceService = "REGRESSO_DO_USUARIO_APOS_ALTA_HOSPITALAR"
	AssistanceServiceCeilingFanReinstallationAndRepair                        AssistanceService = "REINSTALACAO_E_REPARO_DO_VENTILADOR_DE_TETO"
	AssistanceServiceFurnitureRearrangement                                   AssistanceService = "REMANEJAMENTO_DE_MOVEIS"
	AssistanceServiceHospitalRemoval                                          AssistanceService = "REMOCAO_HOSPITALAR"
	AssistanceServiceMedicalRemoval                                           AssistanceService = "REMOCAO_MEDICA"
	AssistanceServiceInterHospitalMedicalRemoval                              AssistanceService = "REMOCAO_MEDICA_INTER_HOSPITALAR"
	AssistanceServiceAutomotiveRepair                                         AssistanceService = "REPARACAO_AUTOMOTIVA"
	AssistanceServiceTelephonyRepair                                          AssistanceService = "REPARO_DE_TELEFONIA"
	AssistanceServiceAutomaticGateRepair                                      AssistanceService = "REPARO_EM_PORTOES_AUTOMATICOS"
	AssistanceServiceAntennaMountingRepair                                    AssistanceService = "REPARO_FIXACAO_DE_ANTENAS"
	AssistanceServiceElectricalRepairs                                        AssistanceService = "REPAROS_ELETRICOS"
	AssistanceServiceEarlyReturnToResidence                                   AssistanceService = "RETORNO_ANTECIPADO_AO_DOMICILIO"
	AssistanceServiceStoveReversal                                            AssistanceService = "REVERSAO_DE_FOGAO"
	AssistanceServiceElectricalInstallationReview                             AssistanceService = "REVISAO_DE_INSTALACAO_ELETRICA"
	AssistanceServiceInternationalSecondMedicalOpinion                        AssistanceService = "SEGUNDA_OPINIAO_MEDICA_INTERNACIONAL"
	AssistanceServiceSecurity                                                 AssistanceService = "SEGURANCA"
	AssistanceServiceMetalworker                                              AssistanceService = "SERRALHEIRO"
	AssistanceServiceMedicalReferralService                                   AssistanceService = "SERVICO_DE_INDICACAO_MEDICA"
	AssistanceServiceCleaningService                                          AssistanceService = "SERVICO_DE_LIMPEZA"
	AssistanceServiceAutoServices                                             AssistanceService = "SERVICOS_AUTO"
	AssistanceServiceSpecialObjectMountingServices                            AssistanceService = "SERVICOS_ESPECIAIS_FIXACAO_DE_OBJETOS"
	AssistanceServiceGeneralServices                                          AssistanceService = "SERVICOS_GERAIS"
	AssistanceServiceTireReplacement                                          AssistanceService = "SUBSTITUICAO_DE_PNEUS"
	AssistanceServiceRoofTileReplacement                                      AssistanceService = "SUBSTITUICAO_DE_TELHAS"
	AssistanceServiceTaxi                                                     AssistanceService = "TAXI"
	AssistanceServiceTelemedicine                                             AssistanceService = "TELEMEDICINA"
	AssistanceServiceUrgentMessageTransmission                                AssistanceService = "TRANSMISSAO_DE_MENSAGENS_URGENTES"
	AssistanceServiceFamilyTransportAndSending                                AssistanceService = "TRANSPORTE_E_ENVIO_DE_FAMILIAR"
	AssistanceServiceFurnitureTransportAndStorage                             AssistanceService = "TRANSPORTE_E_GUARDA_MOVEIS"
	AssistanceServiceSchoolPeopleTransport                                    AssistanceService = "TRANSPORTE_ESCOLAR_PESSOAS"
	AssistanceServiceEmergencyVeterinaryTransport                             AssistanceService = "TRANSPORTE_VETERINARIO_EMERGENCIAL"
	AssistanceServiceBodyTransfer                                             AssistanceService = "TRASLADO_DE_CORPO"
	AssistanceServiceBatteryReplacement                                       AssistanceService = "TROCA_DE_BATERIA"
	AssistanceServiceTireChange                                               AssistanceService = "TROCA_DE_PNEUS"
	AssistanceServiceLeakDetection                                            AssistanceService = "VERIFICACAO_DE_POSSIVEIS_VAZAMENTOS"
	AssistanceServiceGlassAndAccessories                                      AssistanceService = "VIDROS_E_ACESSORIOS"
	AssistanceServiceSurveillanceAndSecurity                                  AssistanceService = "VIGILANCIA_E_SEGURANCA"
	AssistanceServiceOthers                                                   AssistanceService = "OUTROS"
)

type Premium struct {
	PaymentsQuantity         string                 `json:"paymentsQuantity"`
	TotalAmount              insurer.AmountDetails  `json:"totalAmount"`
	TotalNetAmount           insurer.AmountDetails  `json:"totalNetAmount"`
	IOF                      insurer.AmountDetails  `json:"IOF"`
	InterestRateOverPayments *float32               `json:"interestRateOverPayments,omitempty"`
	Coverages                []auto.PremiumCoverage `json:"coverages"`
	Payments                 []auto.Payment         `json:"payments"`
}

type LicensePlateType string

const (
	LicensePlateTypeIncendio           LicensePlateType = "INCENDIO"
	LicensePlateTypeColisao            LicensePlateType = "COLISAO"
	LicensePlateTypeRoubo              LicensePlateType = "ROUBO"
	LicensePlateTypeFurto              LicensePlateType = "FURTO"
	LicensePlateTypeAlagamentoEnchente LicensePlateType = "ALAGAMENTO_ENCHENTE"
	LicensePlateTypeInundacoes         LicensePlateType = "INUNDACOES"
)

type Tariff string

const (
	TariffPassengerNational               Tariff = "PASSEIO_NACIONAL"
	TariffPassengerImported               Tariff = "PASSEIO_IMPORTADO"
	TariffPickupNationalAndImported       Tariff = "PICK_UP_NACIONAL_E_IMPORTADO"
	TariffCargoVehicleNationalAndImported Tariff = "VEICULO_DE_CARGA_NACIONAL_E_IMPORTADO"
	TariffMotorcycleNationalAndImported   Tariff = "MOTOCICLETA_NACIONAL_E_IMPORTADO"
	TariffBusNationalAndImported          Tariff = "ONIBUS_NACIONAL_E_IMPORTADO"
	TariffUtilityNationalAndImported      Tariff = "UTILITARIO_NACIONAL_E_IMPORTADO"
)
