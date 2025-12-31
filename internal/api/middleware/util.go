package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/oidc"
	netmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

const (
	HeaderCustomerIPAddress  = "X-FAPI-Customer-IP-Address"
	HeaderCustomerUserAgent  = "X-Customer-User-Agent"
	HeaderXFAPIInteractionID = "X-Fapi-Interaction-Id"
	HeaderVersion            = "X-V"
	HeaderMinVersion         = "X-Min-V"
)

type Options struct{}

func CertCN(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cert, err := oidc.ClientCert(r)
		if err != nil {
			slog.ErrorContext(r.Context(), "could not get client certificate", "error", err)
			api.WriteError(w, r, api.NewError("UNAUTHORISED", http.StatusUnauthorized, "invalid certificate: could not get client certificate"))
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, api.CtxKeyCertCN, cert.Subject.CommonName)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func FAPIID() func(http.Handler) http.Handler {
	return FAPIIDWithOptions(nil)
}

func FAPIIDWithOptions(_ *Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			interactionID := r.Header.Get(HeaderXFAPIInteractionID)
			if _, err := uuid.Parse(interactionID); err != nil {
				w.Header().Add(HeaderXFAPIInteractionID, uuid.NewString())
				api.WriteError(w, r, api.NewError("PARAMETRO_INVALIDO", http.StatusBadRequest, "The fapi interaction id is missing or invalid"))
				return
			}

			// Return the same interaction ID in the response.
			w.Header().Set(HeaderXFAPIInteractionID, interactionID)
			next.ServeHTTP(w, r)
		})
	}
}

func Swagger(getSwagger func() (*openapi3.T, error), errCodeFunc func(error) api.Error) (middleware func(http.Handler) http.Handler, version string) {
	spec, err := getSwagger()
	if err != nil {
		panic(err)
	}

	return netmiddleware.OapiRequestValidatorWithOptions(spec, &netmiddleware.Options{
		DoNotValidateServers: true,
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
		ErrorHandlerWithOpts: func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request, opts netmiddleware.ErrorHandlerOpts) {
			api.WriteError(w, r, errCodeFunc(err))
		},
	}), spec.Info.Version[:5]
}

func VersionHeader(v string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(HeaderVersion, v)
			next.ServeHTTP(w, r)
		})
	}
}

func VersionRouting(versionHandlers map[string]http.Handler) http.Handler {
	versions := make([]string, 0, len(versionHandlers))
	for v := range versionHandlers {
		versions = append(versions, v)
	}
	// Sort versions in descending order (highest first) using semantic version comparison.
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})
	lastVersion := versions[0]

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := r.Header.Get(HeaderVersion)
		if version == "" {
			version = lastVersion
		}

		minVersion := r.Header.Get(HeaderMinVersion)
		if minVersion == "" || minVersion > version {
			minVersion = lastVersion
		}

		for _, v := range versions {
			if v >= minVersion && v <= version {
				versionHandlers[v].ServeHTTP(w, r)
				return
			}
		}

		api.WriteError(w, r, api.NewError("NOT_ACCEPTABLE", http.StatusNotAcceptable, "version not supported"))
	})
}

// compareVersions compares two semantic versions (format: X.Y.Z).
// It returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := range 3 {
		num1, err1 := strconv.Atoi(parts1[i])
		num2, err2 := strconv.Atoi(parts2[i])
		if err1 != nil || err2 != nil {
			return 0
		}

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	return 0
}
