package auto

import (
	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/auto"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/quote"
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
	Data            QuoteData `gorm:"serializer:json"`
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

type QuoteData struct {
	ExpirationDateTime     timeutil.DateTime    `json:"expirationDateTime"`
	Customer               quote.Customer       `json:"customer"`
	IsCollectiveStipulated bool                 `json:"isCollectiveStipulated"`
	HasAnIndividualItem    bool                 `json:"hasAnIndividualItem"`
	TermStartDate          timeutil.BrazilDate  `json:"termStartDate"`
	TermEndDate            timeutil.BrazilDate  `json:"termEndDate"`
	TermType               insurer.ValidityType `json:"termType"`
	InsuranceType          InsuranceType        `json:"insuranceType"`
	PolicyID               *string              `json:"policyId,omitempty"`
	InsurerID              *string              `json:"insurerId,omitempty"`
	IdentifierCode         *string              `json:"identifierCode,omitempty"`
	BonusClass             *string              `json:"bonusClass,omitempty"`
	Currency               insurer.Currency     `json:"currency"`
	InsuredObject          *InsuredObject       `json:"insuredObject,omitempty"`
	Coverages              *[]Coverage          `json:"coverages,omitempty"`
	CustomData             *quote.CustomData    `json:"customData,omitempty"`
	HistoricalData         *HistoricalData      `json:"historicalData,omitempty"`
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
	IsEquipmentAttached                *bool                       `json:"isEquipmentAttached,omitempty"`
	EquipmentsAttached                 *EquipmentAttached          `json:"equipmentsAttached,omitempty"`
	Chasis                             *string                     `json:"chasis,omitempty"`
	IsAuctionChassisRescheduled        *bool                       `json:"isAuctionChassisRescheduled,omitempty"`
	IsBrandNew                         *bool                       `json:"isBrandNew,omitempty"`
	DepartureDateFromCarDealership     *timeutil.BrazilDate        `json:"departureDateFromCarDealership,omitempty"`
	VehicleInvoice                     *VehicleInvoice             `json:"vehicleInvoice,omitempty"`
	Fuel                               Fuel                        `json:"fuel"`
	IsGasKit                           *bool                       `json:"isGasKit,omitempty"`
	GasKit                             *GasKit                     `json:"gasKit,omitempty"`
	IsArmouredVehicle                  *bool                       `json:"isArmouredVehicle,omitempty"`
	IsActiveTrackingVehicle            *bool                       `json:"isActiveTrackingVehicle,omitempty"`
	FrequentTrafficArea                *TrafficArea                `json:"frequentTrafficArea,omitempty"`
	OvernightPostCode                  *string                     `json:"overnightPostCode,omitempty"`
	RiskLocation                       *RiskLocation               `json:"riskLocation,omitempty"`
	IsExtendCoverageAgedBetween18And25 *bool                       `json:"isExtendCoverageAgedBetween18And25,omitempty"`
	DriverBetween18And25YearsOldGender *DriverGender               `json:"driverBetween18and25YearsOldGender,omitempty"`
	WasThereAClaim                     *bool                       `json:"wasThereAClaim,omitempty"`
	ClaimNotifications                 *[]ClaimNotification        `json:"claimNotifications,omitempty"`
	Coverages                          *[]Coverage                 `json:"coverages,omitempty"`
}

type InsuredObjectModel struct {
	Brand           string  `json:"brand"`
	ModelName       string  `json:"modelName"`
	ModelYear       *string `json:"modelYear,omitempty"`
	ManufactureYear *string `json:"manufactureYear,omitempty"`
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
	CivilStatus        *CivilStatus                `json:"civilStatus,omitempty"`
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

type CivilStatus string

const (
	CivilStatusSingle   CivilStatus = "SOLTEIRO"
	CivilStatusMarried  CivilStatus = "CASADO"
	CivilStatusDivorced CivilStatus = "DIVORCIADO"
	CivilStatusWidowed  CivilStatus = "VIUVO"
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

type LeadQuery struct {
	ID        string
	ConsentID string
}

type Query struct {
	ID        string
	ConsentID string
}
