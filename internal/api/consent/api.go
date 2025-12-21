package consent

import (
	"net/http"

	"github.com/luikyv/go-oidc/pkg/provider"
	v2 "github.com/luikyv/mock-insurer/internal/api/consent/v2"
	"github.com/luikyv/mock-insurer/internal/api/middleware"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/idempotency"
)

type BankConfig interface {
	Host() string
}

type Server struct {
	host               string
	service            consent.Service
	op                 *provider.Provider
	idempotencyService idempotency.Service
}

func NewServer(host string, service consent.Service, op *provider.Provider, idempotencyService idempotency.Service) Server {
	return Server{
		host:               host,
		service:            service,
		op:                 op,
		idempotencyService: idempotencyService,
	}
}

func (s Server) RegisterRoutes(mux *http.ServeMux) {
	muxV2, versionV2 := v2.NewServer(s.host, s.service, s.op, s.idempotencyService).Handler()

	mux.Handle("/open-insurance/consents/v2/", middleware.VersionRouting(map[string]http.Handler{
		versionV2: muxV2,
	}))
}
