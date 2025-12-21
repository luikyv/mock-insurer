package quoteauto

import (
	"net/http"

	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	v1 "github.com/luikyv/mock-insurer/internal/api/quoteauto/v1"
	"github.com/luikyv/mock-insurer/internal/idempotency"
	"github.com/luikyv/mock-insurer/internal/quote/auto"
)

type Server struct {
	host               string
	service            auto.Service
	idempotencyService idempotency.Service
	op                 *provider.Provider
}

func NewServer(host string, service auto.Service, idempotencyService idempotency.Service, op *provider.Provider) Server {
	return Server{
		host:               host,
		service:            service,
		idempotencyService: idempotencyService,
		op:                 op,
	}
}

func (s Server) RegisterRoutes(mux *http.ServeMux) {
	muxV1, versionV1 := v1.NewServer(s.host, s.service, s.idempotencyService, s.op).Handler()

	mux.Handle("/open-insurance/quote-auto/v1/", middleware.VersionRouting(map[string]http.Handler{
		versionV1: muxV1,
	}))
}
