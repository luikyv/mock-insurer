package housing

import (
	"net/http"

	"github.com/luikyv/go-oidc/pkg/provider"
	v1 "github.com/luikyv/mock-insurer/internal/api/housing/v1"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/housing"
)

type Server struct {
	host           string
	service        housing.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(host string, service housing.Service, consentService consent.Service, op *provider.Provider) Server {
	return Server{
		host:           host,
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) RegisterRoutes(mux *http.ServeMux) {
	muxV1, versionV1 := v1.NewServer(s.host, s.service, s.consentService, s.op).Handler()

	mux.Handle("/open-insurance/insurance-housing/v1/", middleware.VersionRouting(map[string]http.Handler{
		versionV1: muxV1,
	}))
}

