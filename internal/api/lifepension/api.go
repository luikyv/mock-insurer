package lifepension

import (
	"net/http"

	"github.com/luikyv/go-oidc/pkg/provider"
	v1 "github.com/luikyv/mock-insurer/internal/api/lifepension/v1"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/lifepension"
)

type Server struct {
	host           string
	service        lifepension.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(host string, service lifepension.Service, consentService consent.Service, op *provider.Provider) Server {
	return Server{
		host:           host,
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) RegisterRoutes(mux *http.ServeMux) {
	muxV1, versionV1 := v1.NewServer(s.host, s.service, s.consentService, s.op).Handler()

	mux.Handle("/open-insurance/insurance-life-pension/v1/", middleware.VersionRouting(map[string]http.Handler{
		versionV1: muxV1,
	}))
}
