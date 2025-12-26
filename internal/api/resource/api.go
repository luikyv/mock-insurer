package resource

import (
	"net/http"

	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	v2 "github.com/luikyv/mock-insurer/internal/api/resource/v2"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/resource"
)

type Server struct {
	host           string
	service        resource.Service
	consentService consent.Service
	op             *provider.Provider
}

func NewServer(host string, service resource.Service, consentService consent.Service, op *provider.Provider) Server {
	return Server{
		host:           host,
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (s Server) RegisterRoutes(mux *http.ServeMux) {
	muxV2, versionV2 := v2.NewServer(s.host, s.service, s.consentService, s.op).Handler()

	mux.Handle("/open-insurance/resources/v2/", middleware.VersionRouting(map[string]http.Handler{
		versionV2: muxV2,
	}))
}
