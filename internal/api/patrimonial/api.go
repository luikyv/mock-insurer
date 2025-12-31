package patrimonial

import (
	"net/http"

	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	v1 "github.com/luikyv/mock-insurer/internal/api/patrimonial/v1"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/patrimonial"
)

type Server struct {
	host           string
	service        patrimonial.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(host string, service patrimonial.Service, consentService consent.Service, op *provider.Provider) Server {
	return Server{
		host:           host,
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) RegisterRoutes(mux *http.ServeMux) {
	muxV1, versionV1 := v1.NewServer(s.host, s.service, s.consentService, s.op).Handler()

	mux.Handle("/open-insurance/insurance-patrimonial/v1/", middleware.VersionRouting(map[string]http.Handler{
		versionV1: muxV1,
	}))
}
