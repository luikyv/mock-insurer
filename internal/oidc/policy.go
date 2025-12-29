package oidc

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/luikyv/mock-insurer/ui"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/internal/auto"
	"github.com/luikyv/mock-insurer/internal/capitalizationtitle"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/page"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"github.com/luikyv/mock-insurer/internal/user"
	"github.com/unrolled/secure"
)

const (
	sessionParamConsentID  = "consent_id"
	sessionParamCPF        = "cpf"
	sessionParamUserID     = "user_id"
	sessionParamBusinessID = "business_id"

	formParamUsername                   = "username"
	formParamPassword                   = "password"
	formParamLogin                      = "login"
	formParamConsent                    = "consent"
	formParamAutoPolicyIDs              = "auto-policies"
	formParamCapitalizationTitlePlanIDs = "capitalization-title-plans"

	correctPassword = "P@ssword01"
)

// TODO: Validate that the resources (accounts, ...) sent belong to the user.
// TODO: Pass the template as a parameter.
func Policies(
	baseURL string,
	userService user.Service,
	consentService consent.Service,
	autoService auto.Service,
	capitalizationTitleService capitalizationtitle.Service,
) []goidc.AuthnPolicy {
	tmpl := template.Must(template.ParseFS(ui.Templates, "*.html"))
	return []goidc.AuthnPolicy{
		goidc.NewPolicyWithSteps(
			"consent",
			func(r *http.Request, c *goidc.Client, as *goidc.AuthnSession) bool {
				consentID, ok := consent.IDFromScopes(as.Scopes)
				if !ok {
					return false
				}

				as.StoreParameter(sessionParamConsentID, consentID)
				as.StoreParameter(OrgIDKey, c.CustomAttribute(OrgIDKey))
				return true
			},
			goidc.NewAuthnStep("setup", validateConsentStep(consentService)),
			goidc.NewAuthnStep("login", loginStep(baseURL, tmpl, userService)),
			goidc.NewAuthnStep("consent", grantConsentStep(
				baseURL, tmpl,
				userService,
				consentService,
				autoService,
				capitalizationTitleService,
			)),
			goidc.NewAuthnStep("finish", grantAuthorizationStep()),
		),
	}
}

func loginStep(baseURL string, tmpl *template.Template, userService user.Service) goidc.AuthnFunc {
	type Page struct {
		BaseURL    string
		CallbackID string
		Nonce      string
		Error      string
	}

	renderLoginPage := func(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {
		return renderPage(w, tmpl, "login", Page{
			BaseURL:    baseURL,
			CallbackID: as.CallbackID,
			Nonce:      secure.CSPNonce(r.Context()),
		})
	}

	renderLoginErrorPage := func(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession, err string) (goidc.Status, error) {
		return renderPage(w, tmpl, "login", Page{
			BaseURL:    baseURL,
			CallbackID: as.CallbackID,
			Nonce:      secure.CSPNonce(r.Context()),
			Error:      err,
		})
	}

	return func(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {
		slog.InfoContext(r.Context(), "starting login step")

		isLogin := r.PostFormValue(formParamLogin)
		if isLogin == "" {
			slog.InfoContext(r.Context(), "rendering login page")
			return renderLoginPage(w, r, as)
		}

		if isLogin != "true" {
			slog.InfoContext(r.Context(), "user cancelled login")
			return goidc.StatusFailure, errors.New("user cancelled login")
		}

		orgID := as.StoredParameter(OrgIDKey).(string)
		username := r.PostFormValue(formParamUsername)
		u, err := userService.User(r.Context(), user.Query{Username: username}, orgID)
		if err != nil {
			slog.InfoContext(r.Context(), "could not fetch user", "error", err)
			return renderLoginErrorPage(w, r, as, "invalid username")
		}

		password := r.PostFormValue(formParamPassword)
		if password != correctPassword {
			slog.InfoContext(r.Context(), "invalid password")
			return renderLoginErrorPage(w, r, as, "invalid credentials")
		}

		slog.InfoContext(r.Context(), "login step finished successfully", "user_id", u.ID, "user_cpf", u.CPF)
		as.StoreParameter(sessionParamUserID, u.ID.String())
		as.StoreParameter(sessionParamCPF, u.CPF)
		return goidc.StatusSuccess, nil
	}
}

func validateConsentStep(consentService consent.Service) goidc.AuthnFunc {
	return func(_ http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {
		orgID := as.StoredParameter(OrgIDKey).(string)
		consentID := as.StoredParameter(sessionParamConsentID).(string)
		c, err := consentService.Consent(r.Context(), consentID, orgID)
		if err != nil {
			slog.InfoContext(r.Context(), "could not fetch the consent", "error", err)
			return goidc.StatusFailure, err
		}

		if c.Status != consent.StatusAwaitingAuthorization {
			slog.InfoContext(r.Context(), "consent is not awaiting authorization", "status", c.Status)
			return goidc.StatusFailure, errors.New("consent is not awaiting authorization")
		}

		return goidc.StatusSuccess, nil
	}
}

func grantConsentStep(
	baseURL string,
	tmpl *template.Template,
	userService user.Service,
	consentService consent.Service,
	autoService auto.Service,
	capitalizationTitleService capitalizationtitle.Service,
) goidc.AuthnFunc {
	type Page struct {
		BaseURL                  string
		CallbackID               string
		UserCPF                  string
		BusinessCNPJ             string
		Nonce                    string
		CustomerPersonalInfo     bool
		CustomerBusinessInfo     bool
		AutoPolicies             []*auto.Policy
		CapitalizationTitlePlans []*capitalizationtitle.Plan
	}

	renderConsentPage := func(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession, c *consent.Consent) (goidc.Status, error) {
		consentPage := Page{
			BaseURL:    baseURL,
			UserCPF:    c.UserIdentification,
			CallbackID: as.CallbackID,
			Nonce:      secure.CSPNonce(r.Context()),
		}

		userID := as.StoredParameter(sessionParamUserID).(string)
		orgID := as.StoredParameter(OrgIDKey).(string)

		if c.Permissions.HasCustomerPersonalPermissions() {
			slog.InfoContext(r.Context(), "rendering consent page with customer personal information")
			consentPage.CustomerPersonalInfo = true
		}

		if c.Permissions.HasCustomerBusinessPermissions() {
			slog.InfoContext(r.Context(), "rendering consent page with customer business information")
			consentPage.CustomerBusinessInfo = true
		}

		if c.Permissions.HasAutoPermissions() {
			slog.InfoContext(r.Context(), "rendering consent page with auto policies")
			policies, err := autoService.Policies(r.Context(), userID, orgID, page.NewPagination(nil, nil))
			if err != nil {
				slog.ErrorContext(r.Context(), "could not load the user's auto policies", "error", err)
				return goidc.StatusFailure, fmt.Errorf("could not load the user's auto policies")
			}
			consentPage.AutoPolicies = policies.Records
		}

		if c.Permissions.HasCapitalizationTitlePermissions() {
			slog.InfoContext(r.Context(), "rendering consent page with capitalization title plans")
			plans, err := capitalizationTitleService.Plans(r.Context(), userID, orgID, page.NewPagination(nil, nil))
			if err != nil {
				slog.ErrorContext(r.Context(), "could not load the user's capitalization title plans", "error", err)
				return goidc.StatusFailure, fmt.Errorf("could not load the user's capitalization title plans")
			}
			consentPage.CapitalizationTitlePlans = plans.Records
		}

		return renderPage(w, tmpl, "consent", consentPage)
	}

	return func(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {
		orgID := as.StoredParameter(OrgIDKey).(string)
		consentID := as.StoredParameter(sessionParamConsentID).(string)
		c, err := consentService.Consent(r.Context(), consentID, orgID)
		if err != nil {
			return goidc.StatusFailure, err
		}

		isConsented := r.PostFormValue(formParamConsent)
		if isConsented == "" {
			if as.StoredParameter(sessionParamCPF) != c.UserIdentification {
				slog.InfoContext(r.Context(), "consent was not created for the correct user")
				reasonAdditionalInfo := "consent was not created for the correct user"
				_ = consentService.Reject(r.Context(), consentID, orgID, consent.Rejection{
					By:                   consent.RejectedByASPSP,
					ReasonCode:           consent.RejectionReasonCodeInternalSecurityReason,
					ReasonAdditionalInfo: &reasonAdditionalInfo,
				})
				return goidc.StatusFailure, errors.New("consent not created for the correct user")
			}

			if c.BusinessIdentification != nil {
				userID := as.StoredParameter(sessionParamUserID).(string)
				business, err := userService.Business(r.Context(), userID, *c.BusinessIdentification, orgID)
				if err != nil {
					slog.InfoContext(r.Context(), "could not fetch the business", "error", err)
					reasonAdditionalInfo := "user has no access to the business"
					_ = consentService.Reject(r.Context(), consentID, orgID, consent.Rejection{
						By:                   consent.RejectedByASPSP,
						ReasonCode:           consent.RejectionReasonCodeInternalSecurityReason,
						ReasonAdditionalInfo: &reasonAdditionalInfo,
					})
					return goidc.StatusFailure, errors.New("user has no access to the business")
				}
				as.StoreParameter(sessionParamBusinessID, business.ID.String())
			}

			slog.InfoContext(r.Context(), "rendering consent page")
			return renderConsentPage(w, r, as, c)
		}

		if isConsented != "true" {
			reasonAdditionalInfo := "user manually rejected consent"
			_ = consentService.Reject(r.Context(), consentID, orgID, consent.Rejection{
				By:                   consent.RejectedByUser,
				ReasonCode:           consent.RejectionReasonCodeCustomerManuallyRejected,
				ReasonAdditionalInfo: &reasonAdditionalInfo,
			})
			return goidc.StatusFailure, errors.New("consent not granted")
		}

		slog.InfoContext(r.Context(), "authorizing consent")
		if err := consentService.Authorize(r.Context(), c); err != nil {
			return goidc.StatusFailure, err
		}

		userID := as.StoredParameter(sessionParamUserID).(string)
		if c.Permissions.HasAutoPermissions() {
			autoPolicyIDs := r.Form[formParamAutoPolicyIDs]
			slog.InfoContext(r.Context(), "authorizing auto policies", "auto policies", autoPolicyIDs)
			if err := autoService.Authorize(r.Context(), autoPolicyIDs, userID, c.ID.String(), orgID); err != nil {
				slog.InfoContext(r.Context(), "could not authorize auto policies", "error", err)
				return goidc.StatusFailure, err
			}
		}

		if c.Permissions.HasCapitalizationTitlePermissions() {
			capitalizationTitlePlanIDs := r.Form[formParamCapitalizationTitlePlanIDs]
			slog.InfoContext(r.Context(), "authorizing capitalization title plans", "capitalization title plans", capitalizationTitlePlanIDs)
			if err := capitalizationTitleService.Authorize(r.Context(), capitalizationTitlePlanIDs, userID, c.ID.String(), orgID); err != nil {
				slog.InfoContext(r.Context(), "could not authorize capitalization title plans", "error", err)
				return goidc.StatusFailure, err
			}
		}

		return goidc.StatusSuccess, nil
	}
}

func grantAuthorizationStep() goidc.AuthnFunc {
	return func(_ http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {
		slog.InfoContext(r.Context(), "auth flow finished, filling oauth session")

		sub := as.StoredParameter(sessionParamUserID).(string)
		if businessID := as.StoredParameter(sessionParamBusinessID); businessID != nil {
			sub = businessID.(string)
		}
		as.SetUserID(sub)
		as.GrantScopes(as.Scopes)
		as.SetIDTokenClaimACR(ACROpenInsuranceLOA2)
		as.SetIDTokenClaimAuthTime(timeutil.Timestamp())

		if as.Claims != nil {
			if slices.Contains(as.Claims.IDTokenEssentials(), goidc.ClaimACR) {
				as.SetIDTokenClaimACR(ACROpenInsuranceLOA2)
			}

			if slices.Contains(as.Claims.UserInfoEssentials(), goidc.ClaimACR) {
				as.SetUserInfoClaimACR(ACROpenInsuranceLOA2)
			}
		}

		return goidc.StatusSuccess, nil
	}
}

func renderPage(w http.ResponseWriter, tmpl *template.Template, name string, data any) (goidc.Status, error) {
	if !strings.HasSuffix(name, ".html") {
		name = name + ".html"
	}

	w.WriteHeader(http.StatusOK)
	// TODO: What happens when an error occurs?
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		return goidc.StatusFailure, fmt.Errorf("could not render template: %w", err)
	}
	return goidc.StatusInProgress, nil
}
