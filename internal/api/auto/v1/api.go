//go:generate oapi-codegen -config=./config.yml -package=v1 -o=./api_gen.go ./swagger.yml
package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/auto"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/errorutil"
	"github.com/luikyv/mock-insurer/internal/insurer"
	"github.com/luikyv/mock-insurer/internal/page"
)

type Server struct {
	baseURL        string
	service        auto.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(
	host string,
	service auto.Service,
	consentService consent.Service,
	op *provider.Provider,
) Server {
	return Server{
		baseURL:        host + "/open-insurance/insurance-auto/v1",
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) Handler() (http.Handler, string) {
	mux := http.NewServeMux()

	authCodeAuthMiddleware := middleware.Auth(s.op, goidc.GrantAuthorizationCode, goidc.ScopeOpenID, auto.Scope)
	swaggerMiddleware, swaggerVersion := middleware.Swagger(GetSwagger, func(err error) api.Error {
		return api.NewError("INVALID_REQUEST", http.StatusBadRequest, err.Error())
	})

	wrapper := ServerInterfaceWrapper{
		Handler: NewStrictHandlerWithOptions(s, nil, StrictHTTPServerOptions{
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				writeResponseError(w, r, err)
			},
		}),
		HandlerMiddlewares: []MiddlewareFunc{swaggerMiddleware},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusBadRequest, err.Error()))
		},
	}

	var handler http.Handler

	handler = http.HandlerFunc(wrapper.GetInsuranceAuto)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleAutoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-auto", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceAutopolicyIDPolicyInfo)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleAutoPolicyInfoRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-auto/{policyId}/policy-info", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceAutopolicyIDPremium)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleAutoPremiumRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-auto/{policyId}/premium", handler)

	handler = http.HandlerFunc(wrapper.GetInsuranceAutopolicyIDClaims)
	handler = middleware.PermissionWithOptions(s.consentService, nil, consent.PermissionDamagesAndPeopleAutoClaimRead)(handler)
	handler = authCodeAuthMiddleware(handler)
	mux.Handle("GET /insurance-auto/{policyId}/claim", handler)

	handler = middleware.FAPIID()(mux)
	return http.StripPrefix("/open-insurance/insurance-auto/v1", handler), swaggerVersion
}

func (s Server) GetInsuranceAuto(ctx context.Context, req GetInsuranceAutoRequestObject) (GetInsuranceAutoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(req.Params.Page, req.Params.PageSize)
	policies, err := s.service.ConsentedPolicies(ctx, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	respPolicies := []struct {
		PolicyID    string `json:"policyId"`
		ProductName string `json:"productName"`
	}{}
	for _, policy := range policies.Records {
		respPolicies = append(respPolicies, struct {
			PolicyID    string `json:"policyId"`
			ProductName string `json:"productName"`
		}{
			PolicyID:    policy.ID,
			ProductName: policy.Data.ProductName,
		})
	}

	resp := ResponseInsuranceAuto{
		Meta:  *api.NewPaginatedMeta(policies),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-auto", policies),
	}
	resp.Data = append(resp.Data, struct {
		Brand     string `json:"brand"`
		Companies []struct {
			CnpjNumber  string `json:"cnpjNumber"`
			CompanyName string `json:"companyName"`
			Policies    []struct {
				PolicyID    string `json:"policyId"`
				ProductName string `json:"productName"`
			} `json:"policies"`
		} `json:"companies"`
	}{
		Brand: insurer.Brand,
		Companies: []struct {
			CnpjNumber  string `json:"cnpjNumber"`
			CompanyName string `json:"companyName"`
			Policies    []struct {
				PolicyID    string `json:"policyId"`
				ProductName string `json:"productName"`
			} `json:"policies"`
		}{{
			CnpjNumber:  insurer.CNPJ,
			CompanyName: insurer.Brand,
			Policies:    respPolicies,
		}},
	})

	return GetInsuranceAuto200JSONResponse{OKResponseInsuranceAutoJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceAutopolicyIDPolicyInfo(ctx context.Context, request GetInsuranceAutopolicyIDPolicyInfoRequestObject) (GetInsuranceAutopolicyIDPolicyInfoResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, request.PolicyID, consentID, orgID)
	if err != nil {
		return nil, err
	}

	policyInfo := InsuranceAutoPolicyInfo{
		PolicyID:      policy.ID,
		DocumentType:  InsuranceAutoPolicyInfoDocumentType(policy.Data.DocumentType),
		IssuanceType:  InsuranceAutoPolicyInfoIssuanceType(policy.Data.IssuanceType),
		IssuanceDate:  policy.Data.IssuanceDate,
		TermStartDate: policy.Data.TermStartDate,
		TermEndDate:   policy.Data.TermEndDate,
		ProposalID:    policy.Data.ProposalID,
		Beneficiaries: func() *[]BeneficiaryInfo {
			if policy.Data.Beneficiaries == nil {
				return nil
			}
			beneficiaries := make([]BeneficiaryInfo, len(*policy.Data.Beneficiaries))
			for i, b := range *policy.Data.Beneficiaries {
				beneficiaries[i] = BeneficiaryInfo{
					Identification:           b.Identification,
					IdentificationType:       BeneficiaryInfoIdentificationType(b.IdentificationType),
					IdentificationTypeOthers: b.IdentificationTypeOthers,
					Name:                     b.Name,
				}
			}
			return &beneficiaries
		}(),
		MaxLMG:                        policy.Data.MaxLMG,
		SusepProcessNumber:            policy.Data.SusepProcessNumber,
		GroupCertificateID:            policy.Data.GroupCertificateID,
		LeadInsurerCode:               policy.Data.LeadInsurerCode,
		LeadInsurerPolicyID:           policy.Data.LeadInsurerPolicyID,
		CoinsuranceRetainedPercentage: policy.Data.CoinsuranceRetainedPercentage,
		RepairNetwork:                 InsuranceAutoPolicyInfoRepairNetwork(policy.Data.RepairNetwork),
		RepairNetworkOthers:           policy.Data.RepairNetworkOthers,
		RepairedPartsUsageType:        InsuranceAutoPolicyInfoRepairedPartsUsageType(policy.Data.RepairedPartsUsageType),
		RepairedPartsClassification:   InsuranceAutoPolicyInfoRepairedPartsClassification(policy.Data.RepairedPartsClassification),
		RepairedPartsNationality:      InsuranceAutoPolicyInfoRepairedPartsNationality(policy.Data.RepairedPartsNationality),
		ValidityType:                  InsuranceAutoPolicyInfoValidityType(policy.Data.ValidityType),
		ValidityTypeOthers:            policy.Data.ValidateTypeOthers,
		OtherCompensations:            policy.Data.OtherCompensations,
		IsExpiredRiskPolicy:           policy.Data.IsExpiredRiskPolicy,
		BonusDiscountRate:             policy.Data.BonusDiscountRate,
		BonusClass:                    policy.Data.BonusClass,
	}

	if policy.Data.Principals != nil {
		principals := make([]Principals, len(*policy.Data.Principals))
		for i, p := range *policy.Data.Principals {
			principals[i] = Principals{
				Identification:           p.Identification,
				IdentificationType:       PrincipalsIdentificationType(p.IdentificationType),
				IdentificationTypeOthers: p.IdentificationTypeOthers,
				Name:                     p.Name,
				PostCode:                 p.PostCode,
				Email:                    p.Email,
				City:                     p.City,
				State:                    PrincipalsState(p.State),
				Country:                  PrincipalsCountry(p.Country),
				Address:                  p.Address,
				AddressAditionalInfo:     p.AddressAdditionalInfo,
			}
		}
		policyInfo.Principals = &principals
	}

	if policy.Data.Intermediaries != nil {
		intermediaries := make([]Intermediary, len(*policy.Data.Intermediaries))
		for i, in := range *policy.Data.Intermediaries {
			intermediary := Intermediary{
				Type:                     IntermediaryType(in.Type),
				TypeOthers:               in.TypeOthers,
				Identification:           in.Identification,
				BrokerID:                 in.BrokerID,
				IdentificationTypeOthers: in.IdentificationTypeOthers,
				Name:                     in.Name,
				PostCode:                 in.PostCode,
				City:                     in.City,
				Address:                  in.Address,
			}

			if in.IdentificationType != nil {
				idType := IntermediaryIdentificationType(*in.IdentificationType)
				intermediary.IdentificationType = &idType
			}

			if in.State != nil {
				state := IntermediaryState(*in.State)
				intermediary.State = &state
			}

			if in.Country != nil {
				country := IntermediaryCountry(*in.Country)
				intermediary.Country = &country
			}

			intermediaries[i] = intermediary
		}
		policyInfo.Intermediaries = &intermediaries
	}

	insureds := make([]PersonalInfo, len(policy.Data.Insureds))
	for i, ins := range policy.Data.Insureds {
		insureds[i] = PersonalInfo{
			Identification:           ins.Identification,
			IdentificationType:       PersonalInfoIdentificationType(ins.IdentificationType),
			IdentificationTypeOthers: ins.IdentificationTypeOthers,
			Name:                     ins.Name,
			BirthDate:                ins.BirthDate,
			PostCode:                 ins.PostCode,
			Email:                    ins.Email,
			City:                     ins.City,
			State:                    PersonalInfoState(ins.State),
			Country:                  PersonalInfoCountry(ins.Country),
			Address:                  ins.Address,
		}
	}
	policyInfo.Insureds = insureds

	insuredObjects := make([]InsuranceAutoInsuredObject, len(policy.Data.InsuredObjects))
	for i, obj := range policy.Data.InsuredObjects {
		insuredObj := InsuranceAutoInsuredObject{
			Identification:                obj.Identification,
			Type:                          InsuranceAutoInsuredObjectType(obj.IdentificationType),
			TypeAdditionalInfo:            obj.IdentificationTypeAdditionalInfo,
			Description:                   obj.Description,
			HasExactVehicleIdentification: obj.HasExactVehicleIdentification,
			ModalityOthers:                obj.ModalityOthers,
			AmountReferenceTableOthers:    obj.AmountReferenceTableOthers,
			Model:                         obj.Model,
			Year:                          obj.Year,
			RiskPostCode:                  obj.RiskPostCode,
			VehicleUsageOthers:            obj.VehicleUsageOthers,
			FrequentDestinationPostCode:   obj.FrequentDestinationPostCode,
			OvernightPostCode:             obj.OvernightPostCode,
		}

		if obj.Modality != nil {
			modality := InsuranceAutoInsuredObjectModality(*obj.Modality)
			insuredObj.Modality = &modality
		}
		if obj.AmountReferenceTable != nil {
			table := InsuranceAutoInsuredObjectAmountReferenceTable(*obj.AmountReferenceTable)
			insuredObj.AmountReferenceTable = &table
		}
		if obj.FareCategory != nil {
			category := InsuranceAutoInsuredObjectFareCategory(*obj.FareCategory)
			insuredObj.FareCategory = &category
		}
		if obj.VehicleUsage != nil {
			usage := InsuranceAutoInsuredObjectVehicleUsage(*obj.VehicleUsage)
			insuredObj.VehicleUsage = &usage
		}

		coverages := make([]InsuranceAutoInsuredObjectCoverage, len(obj.Coverages))
		for j, cov := range obj.Coverages {
			coverages[j] = InsuranceAutoInsuredObjectCoverage{
				Branch:                        cov.Branch,
				Code:                          InsuranceAutoInsuredObjectCoverageCode(cov.Code),
				Description:                   cov.Description,
				InternalCode:                  cov.InternalCode,
				SusepProcessNumber:            cov.SusepProcessNumber,
				LMI:                           cov.LMI,
				TermStartDate:                 cov.TermStartDate,
				TermEndDate:                   cov.TermEndDate,
				IsMainCoverage:                cov.IsMainCoverage,
				Feature:                       InsuranceAutoInsuredObjectCoverageFeature(cov.Feature),
				Type:                          InsuranceAutoInsuredObjectCoverageType(cov.Type),
				GracePeriod:                   cov.GracePeriod,
				AdjustmentRate:                cov.AdjustmentRate,
				PremiumAmount:                 cov.PremiumAmount,
				CompensationTypeOthers:        cov.CompensationTypeOthers,
				PartialCompensationPercentage: cov.PartialCompensationPercentage,
				PercentageOverLMI:             cov.PercentageOverLMI,
				DaysForTotalCompensation:      cov.DaysForTotalCompensation,
				BoundCoverageOthers:           cov.BoundCoverageOthers,
			}

			if cov.GracePeriodicity != nil {
				periodicity := InsuranceAutoInsuredObjectCoverageGracePeriodicity(*cov.GracePeriodicity)
				coverages[j].GracePeriodicity = &periodicity
			}
			if cov.GracePeriodCountingMethod != nil {
				method := InsuranceAutoInsuredObjectCoverageGracePeriodCountingMethod(*cov.GracePeriodCountingMethod)
				coverages[j].GracePeriodCountingMethod = &method
			}
			if cov.GracePeriodStartDate != nil {
				coverages[j].GracePeriodStartDate = cov.GracePeriodStartDate
			}
			if cov.GracePeriodEndDate != nil {
				coverages[j].GracePeriodEndDate = cov.GracePeriodEndDate
			}
			if cov.CompensationType != nil {
				compType := InsuranceAutoInsuredObjectCoverageCompensationType(*cov.CompensationType)
				coverages[j].CompensationType = &compType
			}
			if cov.BoundCoverage != nil {
				bound := InsuranceAutoInsuredObjectCoverageBoundCoverage(*cov.BoundCoverage)
				coverages[j].BoundCoverage = &bound
			}
			premiumPeriodicity := InsuranceAutoInsuredObjectCoveragePremiumPeriodicity(cov.PremiumPeriodicity)
			coverages[j].PremiumPeriodicity = premiumPeriodicity
			coverages[j].PremiumPeriodicityOthers = cov.PremiumPeriodicityOthers
		}
		insuredObj.Coverages = coverages
		insuredObjects[i] = insuredObj
	}
	policyInfo.InsuredObjects = insuredObjects

	if len(policy.Data.Coverages) > 0 {
		coverages := make([]InsuranceAutoCoverage, len(policy.Data.Coverages))
		for i, cov := range policy.Data.Coverages {
			coverage := InsuranceAutoCoverage{
				Branch:      "",
				Code:        InsuranceAutoCoverageCode(cov.Code),
				Description: cov.Description,
			}
			if cov.Branch != nil {
				coverage.Branch = *cov.Branch
			}
			if cov.Deductible != nil {
				deductible := InsuranceAutoDeductible{
					Type:                               InsuranceAutoDeductibleType(cov.Deductible.Type),
					TypeAdditionalInfo:                 cov.Deductible.TypeOthers,
					Amount:                             cov.Deductible.Amount,
					Period:                             cov.Deductible.Period,
					Description:                        cov.Deductible.Description,
					HasDeductibleOverTotalCompensation: cov.Deductible.HasDeductibleOverTotalCompensation,
					PeriodStartDate:                    cov.Deductible.PeriodStartDate,
					PeriodEndDate:                      cov.Deductible.PeriodEndDate,
				}
				if cov.Deductible.Periodicity != nil {
					periodicity := InsuranceAutoDeductiblePeriodicity(*cov.Deductible.Periodicity)
					deductible.Periodicity = &periodicity
				}
				if cov.Deductible.PeriodCountingMethod != nil {
					method := InsuranceAutoDeductiblePeriodCountingMethod(*cov.Deductible.PeriodCountingMethod)
					deductible.PeriodCountingMethod = &method
				}
				coverage.Deductible = &deductible
			}
			if cov.POS != nil {
				pos := InsuranceAutoPOS{
					ApplicationType: InsuranceAutoPOSApplicationType(cov.POS.ApplicationType),
					Description:     cov.POS.Description,
					MinValue:        cov.POS.MinValue,
					MaxValue:        cov.POS.MaxValue,
					Percentage:      cov.POS.Percentage,
					ValueOthers:     cov.POS.ValueOthers,
				}
				coverage.POS = &pos
			}
			coverages[i] = coverage
		}
		policyInfo.Coverages = &coverages
	}

	if policy.Data.Coinurers != nil {
		coinsurers := make([]Coinsurer, len(*policy.Data.Coinurers))
		for i, c := range *policy.Data.Coinurers {
			coinsurers[i] = Coinsurer{
				Identification:  c.Identification,
				CededPercentage: c.CededPercentage,
			}
		}
		policyInfo.Coinsurers = &coinsurers
	}

	if policy.Data.Drivers != nil {
		drivers := make([]Driver, len(*policy.Data.Drivers))
		for i, d := range *policy.Data.Drivers {
			driver := Driver{
				Identification:     d.Identification,
				BirthDate:          d.BirthDate,
				LicensedExperience: d.LicensedExperience,
				SexOthers:          d.SexOthers,
			}
			if d.Sex != nil {
				sex := DriverSex(*d.Sex)
				driver.Sex = &sex
			}
			drivers[i] = driver
		}
		policyInfo.Drivers = &drivers
	}

	if policy.Data.OtherBenefits != nil {
		benefits := InsuranceAutoPolicyInfoOtherBenefits(*policy.Data.OtherBenefits)
		policyInfo.OtherBenefits = &benefits
	}

	if policy.Data.AssistancePackages != nil {
		packages := InsuranceAutoPolicyInfoAssistancePackages(*policy.Data.AssistancePackages)
		policyInfo.AssistancePackages = &packages
	}

	resp := ResponseInsuranceAutoPolicyInfo{
		Data:  policyInfo,
		Links: *api.NewLinks(s.baseURL + "/insurance-auto/" + request.PolicyID + "/policy-info"),
		Meta:  *api.NewMeta(),
	}

	return GetInsuranceAutopolicyIDPolicyInfo200JSONResponse{OKResponseInsuranceAutoPolicyInfoJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceAutopolicyIDPremium(ctx context.Context, request GetInsuranceAutopolicyIDPremiumRequestObject) (GetInsuranceAutopolicyIDPremiumResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	policy, err := s.service.ConsentedPolicy(ctx, request.PolicyID, consentID, orgID)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceAutoPremium{
		Meta:  *api.NewMeta(),
		Links: *api.NewLinks(s.baseURL + "/insurance-auto/" + request.PolicyID + "/premium"),
	}

	// Convert premium data - take the first premium if available
	// Convert coverages
	coverages := make([]InsuranceAutoPremiumCoverage, 0, len(policy.Data.Premium.Coverages))
	for _, cov := range policy.Data.Premium.Coverages {
		coverageResp := InsuranceAutoPremiumCoverage{
			Branch:        cov.Branch,
			Code:          InsuranceAutoPremiumCoverageCode(cov.Code),
			Description:   cov.Description,
			PremiumAmount: cov.PremiumAmount,
		}
		coverages = append(coverages, coverageResp)
	}

	// Convert payments
	payments := make([]Payment, 0, len(policy.Data.Premium.Payments))
	for _, pay := range policy.Data.Premium.Payments {
		paymentResp := Payment{
			MovementDate:             pay.MovementDate,
			MovementType:             PaymentMovementType(pay.MovementType),
			MovementPaymentsNumber:   float32(pay.MovementPaymentsNumber),
			Amount:                   pay.Amount,
			MaturityDate:             pay.MaturityDate,
			TellerID:                 pay.TellerID,
			TellerIDOthers:           pay.TellerIDOthers,
			TellerName:               pay.TellerName,
			FinancialInstitutionCode: pay.FinancialInstitutionCode,
			PaymentTypeOthers:        pay.PaymentTypeOthers,
		}

		if pay.MovementOrigin != nil {
			origin := PaymentMovementOrigin(*pay.MovementOrigin)
			paymentResp.MovementOrigin = &origin
		}
		if pay.TellerIDType != nil {
			tellerIDType := PaymentTellerIDType(*pay.TellerIDType)
			paymentResp.TellerIDType = &tellerIDType
		}
		if pay.PaymentType != nil {
			paymentType := PaymentPaymentType(*pay.PaymentType)
			paymentResp.PaymentType = &paymentType
		}

		payments = append(payments, paymentResp)
	}

	resp.Data = InsuranceAutoPremium{
		PaymentsQuantity: policy.Data.Premium.PaymentsQuantity,
		Amount:           policy.Data.Premium.Amount,
		Coverages:        coverages,
		Payments:         payments,
	}

	return GetInsuranceAutopolicyIDPremium200JSONResponse{OKResponseInsuranceAutoPremiumJSONResponse(resp)}, nil
}

func (s Server) GetInsuranceAutopolicyIDClaims(ctx context.Context, request GetInsuranceAutopolicyIDClaimsRequestObject) (GetInsuranceAutopolicyIDClaimsResponseObject, error) {
	orgID := ctx.Value(api.CtxKeyOrgID).(string)
	consentID := ctx.Value(api.CtxKeyConsentID).(string)
	pag := page.NewPagination(request.Params.Page, request.Params.PageSize)
	claims, err := s.service.ConsentedClaims(ctx, request.PolicyID, consentID, orgID, pag)
	if err != nil {
		return nil, err
	}

	resp := ResponseInsuranceAutoClaims{
		Meta:  *api.NewPaginatedMeta(claims),
		Links: *api.NewPaginatedLinks(s.baseURL+"/insurance-auto/"+request.PolicyID+"/claim", claims),
	}

	respClaims := make([]InsuranceAutoClaim, 0, len(claims.Records))
	for _, claim := range claims.Records {
		claimResp := InsuranceAutoClaim{
			Identification:                 claim.Data.Identification,
			DocumentationDeliveryDate:      claim.Data.DocumentationDeliveryDate,
			Status:                         InsuranceAutoClaimStatus(claim.Data.Status),
			StatusAlterationDate:           claim.Data.StatusAlterationDate,
			OccurrenceDate:                 claim.Data.OccurrenceDate,
			WarningDate:                    claim.Data.WarningDate,
			ThirdPartyClaimDate:            claim.Data.ThirdPartyClaimDate,
			Amount:                         claim.Data.Amount,
			DenialJustificationDescription: claim.Data.DenialJustificationDescription,
			Coverages:                      make([]InsuranceAutoClaimCoverage, 0, len(claim.Data.Coverages)),
		}

		if claim.Data.DenialJustification != nil {
			denialJust := InsuranceAutoClaimDenialJustification(*claim.Data.DenialJustification)
			claimResp.DenialJustification = &denialJust
		}

		for _, cov := range claim.Data.Coverages {
			claimResp.Coverages = append(claimResp.Coverages, InsuranceAutoClaimCoverage{
				InsuredObjectID:     cov.InsuredObjectId,
				Branch:              cov.Branch,
				Code:                InsuranceAutoClaimCoverageCode(cov.Code),
				Description:         cov.Description,
				WarningDate:         cov.WarningDate,
				ThirdPartyClaimDate: cov.ThirdPartyClaimDate,
			})
		}

		if claim.Data.BranchInfo != nil {
			branchInfo := InsuranceAutoSpecificClaim{
				CovenantNumber:              claim.Data.BranchInfo.CovenantNumber,
				OccurrenceCauseOthers:       claim.Data.BranchInfo.OccurenceCauseOthers,
				DriverAtOccurrenceSexOthers: claim.Data.BranchInfo.DriverAtOccurrenceSexOthers,
				DriverAtOccurrenceBirthDate: claim.Data.BranchInfo.DriverAtOccurrenceBirthDate,
				OccurrencePostCode:          claim.Data.BranchInfo.OccurrencePostCode,
			}
			if claim.Data.BranchInfo.OccurenceCause != nil {
				occCause := InsuranceAutoSpecificClaimOccurrenceCause(*claim.Data.BranchInfo.OccurenceCause)
				branchInfo.OccurrenceCause = &occCause
			}
			if claim.Data.BranchInfo.DriverAtOccurrenceSex != nil {
				sex := InsuranceAutoSpecificClaimDriverAtOccurrenceSex(*claim.Data.BranchInfo.DriverAtOccurrenceSex)
				branchInfo.DriverAtOccurrenceSex = &sex
			}
			if claim.Data.BranchInfo.OccurrenceCountry != nil {
				country := InsuranceAutoSpecificClaimOccurrenceCountry(*claim.Data.BranchInfo.OccurrenceCountry)
				branchInfo.OccurrenceCountry = &country
			}
			claimResp.BranchInfo = &branchInfo
		}

		respClaims = append(respClaims, claimResp)
	}
	resp.Data = respClaims

	return GetInsuranceAutopolicyIDClaims200JSONResponse{OKResponseInsuranceAutoClaimsJSONResponse(resp)}, nil
}

func writeResponseError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &errorutil.Error{}) {
		api.WriteError(w, r, api.NewError("INVALID_REQUEST", http.StatusUnprocessableEntity, err.Error()))
		return
	}

	api.WriteError(w, r, err)
}
