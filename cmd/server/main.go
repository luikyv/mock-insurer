package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/luikyv/mock-insurer/cmd/cmdutil"
	"github.com/luikyv/mock-insurer/internal/acceptancebranchesabroad"
	acceptancebranchesabroadapi "github.com/luikyv/mock-insurer/internal/api/acceptancebranchesabroad"
	autoapi "github.com/luikyv/mock-insurer/internal/api/auto"
	capitalizationtitleapi "github.com/luikyv/mock-insurer/internal/api/capitalizationtitle"
	consentapi "github.com/luikyv/mock-insurer/internal/api/consent"
	customerapi "github.com/luikyv/mock-insurer/internal/api/customer"
	financialassistanceapi "github.com/luikyv/mock-insurer/internal/api/financialassistance"
	financialriskapi "github.com/luikyv/mock-insurer/internal/api/financialrisk"
	housingapi "github.com/luikyv/mock-insurer/internal/api/housing"
	lifepensionapi "github.com/luikyv/mock-insurer/internal/api/lifepension"
	oidcapi "github.com/luikyv/mock-insurer/internal/api/oidc"
	patrimonialapi "github.com/luikyv/mock-insurer/internal/api/patrimonial"
	quoteautoapi "github.com/luikyv/mock-insurer/internal/api/quoteauto"
	resourceapi "github.com/luikyv/mock-insurer/internal/api/resource"
	"github.com/luikyv/mock-insurer/internal/auto"
	"github.com/luikyv/mock-insurer/internal/client"
	"github.com/luikyv/mock-insurer/internal/customer"
	"github.com/luikyv/mock-insurer/internal/financialrisk"
	"github.com/luikyv/mock-insurer/internal/housing"
	"github.com/luikyv/mock-insurer/internal/idempotency"
	"github.com/luikyv/mock-insurer/internal/lifepension"
	"github.com/luikyv/mock-insurer/internal/patrimonial"
	quoteauto "github.com/luikyv/mock-insurer/internal/quote/auto"
	"github.com/luikyv/mock-insurer/internal/resource"
	"github.com/luikyv/mock-insurer/internal/webhook"

	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/capitalizationtitle"
	"github.com/luikyv/mock-insurer/internal/consent"
	"github.com/luikyv/mock-insurer/internal/financialassistance"
	"github.com/luikyv/mock-insurer/internal/oidc"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"github.com/luikyv/mock-insurer/internal/user"
	"gorm.io/gorm"
)

var (
	Env = cmdutil.EnvValue("ENV", cmdutil.LocalEnvironment)
	// OrgID is the Mock Insurer organization identifier.
	OrgID                   = cmdutil.EnvValue("ORG_ID", "00000000-0000-0000-0000-000000000000")
	BaseDomain              = cmdutil.EnvValue("BASE_DOMAIN", "mockinsurer.local")
	AuthHost                = "https://auth." + BaseDomain
	AuthMTLSHost            = "https://matls-auth." + BaseDomain
	APIMTLSHost             = "https://matls-api." + BaseDomain
	KeyStoreHost            = cmdutil.EnvValue("KEYSTORE_HOST", "https://keystore.local")
	SoftwareStatementIssuer = cmdutil.EnvValue("SS_ISSUER", "Open Banking Brasil sandbox SSA issuer")
	Port                    = cmdutil.EnvValue("PORT", "80")
	DBCredentials           = cmdutil.EnvValue("DB_CREDENTIALS", `{"username":"admin","password":"pass","host":"database.local","port":5432,"dbname":"mockinsurer","sslmode":"disable"}`)
	// TransportCertPath and TransportKeyPath are the file paths used for mutual TLS connections.
	TransportCertPath = cmdutil.EnvValue("TRANSPORT_CERT_PATH", "../../keys/server_transport.crt")
	TransportKeyPath  = cmdutil.EnvValue("TRANSPORT_KEY_PATH", "../../keys/server_transport.key")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.SetDefault(logger())
	slog.Info("setting up mock insurer", "env", Env)
	http.DefaultClient = httpClient()

	// Database.
	slog.Info("connecting to database")
	db, err := cmdutil.DB(ctx, DBCredentials)
	if err != nil {
		slog.Error("failed connecting to database", "error", err)
		os.Exit(1)
	}
	slog.Info("successfully connected to database")

	// Keys.
	transportTLSCert, err := tls.LoadX509KeyPair(TransportCertPath, TransportKeyPath)
	if err != nil {
		slog.Error("could not load transport TLS certificate", "error", err)
		os.Exit(1)
	}

	// Services.
	clientService := client.NewService(db)
	idempotencyService := idempotency.NewService(db)
	_ = webhook.NewService(clientService, mtlsHTTPClient(transportTLSCert))
	userService := user.NewService(db)
	resourceService := resource.NewService(db)
	consentService := consent.NewService(db, userService)
	customerService := customer.NewService(db)
	autoService := auto.NewService(db)
	capitalizationTitleService := capitalizationtitle.NewService(db)
	financialAssistanceService := financialassistance.NewService(db)
	acceptanceAndBranchesAbroadService := acceptancebranchesabroad.NewService(db)
	financialRiskService := financialrisk.NewService(db)
	housingService := housing.NewService(db)
	lifePensionService := lifepension.NewService(db)
	patrimonialService := patrimonial.NewService(db)
	quoteAutoService := quoteauto.NewService(db)

	op, err := openidProvider(
		db,
		clientService,
		userService,
		consentService,
		autoService,
		capitalizationTitleService,
		financialAssistanceService,
		acceptanceAndBranchesAbroadService,
		financialRiskService,
		housingService,
		lifePensionService,
		patrimonialService,
	)
	if err != nil {
		slog.Error("failed to create openid provider", "error", err)
		os.Exit(1)
	}

	// Servers.
	mux := http.NewServeMux()

	oidcapi.NewServer(AuthHost, op).RegisterRoutes(mux)
	consentapi.NewServer(APIMTLSHost, consentService, op, idempotencyService).RegisterRoutes(mux)
	resourceapi.NewServer(APIMTLSHost, resourceService, consentService, op).RegisterRoutes(mux)
	customerapi.NewServer(APIMTLSHost, customerService, consentService, op).RegisterRoutes(mux)
	autoapi.NewServer(APIMTLSHost, autoService, consentService, op).RegisterRoutes(mux)
	capitalizationtitleapi.NewServer(APIMTLSHost, capitalizationTitleService, consentService, op).RegisterRoutes(mux)
	financialassistanceapi.NewServer(APIMTLSHost, financialAssistanceService, consentService, op).RegisterRoutes(mux)
	acceptancebranchesabroadapi.NewServer(APIMTLSHost, acceptanceAndBranchesAbroadService, consentService, op).RegisterRoutes(mux)
	financialriskapi.NewServer(APIMTLSHost, financialRiskService, consentService, op).RegisterRoutes(mux)
	housingapi.NewServer(APIMTLSHost, housingService, consentService, op).RegisterRoutes(mux)
	lifepensionapi.NewServer(APIMTLSHost, lifePensionService, consentService, op).RegisterRoutes(mux)
	patrimonialapi.NewServer(APIMTLSHost, patrimonialService, consentService, op).RegisterRoutes(mux)
	quoteautoapi.NewServer(APIMTLSHost, quoteAutoService, idempotencyService, op).RegisterRoutes(mux)

	handler := middleware(mux)
	slog.Info("starting mock insurer")

	if err := http.ListenAndServe(":"+Port, handler); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start mock insurer", "error", err)
		os.Exit(1)
	}
}

func logger() *slog.Logger {
	return slog.New(&logCtxHandler{
		Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
			// Make sure time is logged in UTC.
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey {
					now := timeutil.DateTimeNow()
					return slog.Attr{Key: slog.TimeKey, Value: slog.StringValue(now.String())}
				}
				return attr
			},
		}),
	})
}

type logCtxHandler struct {
	slog.Handler
}

func (h *logCtxHandler) Handle(ctx context.Context, r slog.Record) error {
	if correlationID, ok := ctx.Value(api.CtxKeyCorrelationID).(string); ok {
		r.AddAttrs(slog.String("correlation_id", correlationID))
	}

	if interactionID, ok := ctx.Value(api.CtxKeyInteractionID).(string); ok {
		r.AddAttrs(slog.String("interaction_id", interactionID))
	}

	return h.Handler.Handle(ctx, r)
}

func httpClient() *http.Client {
	tlsConfig := &tls.Config{
		Renegotiation: tls.RenegotiateOnceAsClient,
	}
	if Env == cmdutil.LocalEnvironment {
		tlsConfig.InsecureSkipVerify = true
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}

func mtlsHTTPClient(cert tls.Certificate) *http.Client {
	tlsConfig := &tls.Config{
		Certificates:  []tls.Certificate{cert},
		MinVersion:    tls.VersionTLS12,
		Renegotiation: tls.RenegotiateOnceAsClient,
	}
	if Env == cmdutil.LocalEnvironment {
		tlsConfig.InsecureSkipVerify = true
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}

func openidProvider(
	db *gorm.DB,
	clientService client.Service,
	userService user.Service,
	consentService consent.Service,
	autoService auto.Service,
	capitalizationTitleService capitalizationtitle.Service,
	financialAssistanceService financialassistance.Service,
	acceptanceAndBranchesAbroadService acceptancebranchesabroad.Service,
	financialRiskService financialrisk.Service,
	housingService housing.Service,
	lifePensionService lifepension.Service,
	patrimonialService patrimonial.Service,
) (*provider.Provider, error) {
	var scopes = []goidc.Scope{
		goidc.ScopeOpenID,
		consent.ScopeID,
		consent.Scope,
		resource.Scope,
		customer.Scope,
		auto.Scope,
		capitalizationtitle.Scope,
		financialassistance.Scope,
		acceptancebranchesabroad.Scope,
		financialrisk.Scope,
		housing.Scope,
		lifepension.Scope,
		patrimonial.Scope,
		quoteauto.Scope,
		quoteauto.ScopeLead,
		goidc.NewScope("dynamic-fields"),
	}

	var jwks goidc.JSONWebKeySet
	if err := json.Unmarshal([]byte(`{
      "keys": [
        {
	  	  "kid": "signer",
          "p": "5BKxIVlA8DKoAbXnyNr-M_nAAi63lUCrCki7ADrsifHgTspQydfdQVA8DcqS0JxaHGlWr-mCjrMSSd8x1WOWW8TNqf0NF9O3XZGuCG35xbLG8V72pIMPM5HWr91RTQ0w6FqYkRJsot2ZYK53rtsDSwqQPK7LbRZTaSs-MCB-6SE",
          "kty": "RSA",
          "q": "u1_MSt9DNMqgL1N24S5VHXYmNH8p1ZP70KqH4WmJuYQfbgqQ7sU0L7nkR_H_IHHZqL3bruYVNPcKaE7tmHH5sRkix_R_MudjynV2la03UCKoSvnUgb0dguW9xDHKaXyVhzi24OPjolhhu0RYOqzF2GSJ2yZ0Z1zjPNLksEhxC2c",
          "d": "lWi6shKVV8-nggjqc8PmpWOmMIDvfOiYUWVinjwyDEueljRBFUqrc8Z_lNQraEfM9dQ-GfNycEM9wN581H6M80hoVepTROMSYPZfDE_mX6aE48OReo6hJQvB3tUAuSkdjQj9_Tc_TLQEott-L89IJsAqP7AQSS0WvzfJL4O-YtIiyYNqbgbRfVTfaGAMIUKlO_dEf8jsbigBAbGVT7LIcAf9UokUBKc7Kudl_xCMzbDdM03xJeC5Ml0peOmnnTGc2NdBHSyXITnkrOJlnQWz5ZzyiJ3Os9Zm34gWcdXDz8emS9AHftqv6c9FbmBq9jMNU4_tiIMdAo6-p-xizlJawQ",
          "e": "AQAB",
          "use": "sig",
          "qi": "TsQgLFpb6TozodDm_zoUcoPY6sWlijhvHgFFKAjnB2ssCJ6lO7X3bnep_cUt-dtkV7eZnh5A0cUNVtNnP3ni9EkpxWwsSTrG_TZHmbHxaHXnF-G4s6lrYGWBXbgucALegEFpmuJUThxpLEbpPOUzsWTxI66nOd6T5Quwx7qGv30",
          "dp": "0FrfJMckEwtD_qQOxqiBeDweFCBXqGs2liORaolqFC86quAa4_pnb8Z7xmGctCVSEQiOoBAkLHcdKw1Suk3LS7TD6hp6Pp00s69lnN_TQa-sHU-S5QGx_nup9GmsX0bAulQhcs6xHixxdSiNv9jm7kQNNtK8lsDBnJ9bpZ3aMuE",
          "alg": "PS256",
          "dq": "Z7QtrYLD_4PmBEt9kEPEd_ncS1HWJY8x39uCOQ_gWfz2KEFQ1dXvfDq2TdtyCNL6VJo_7B0Lv7S63eBRP_5U49-1kFWR0OqgIH3ClDS6WG_WFSkQpH22x6u_y8aC8L8zQxPwo6d9ZWzlKnA5JMBa_9klM1WlN1ABtLhEOgzeBCE",
          "n": "pu8AVLEIfYppnbU0r2M1PNhCvYpGnVXbSXj-OxRX72e_us-pYg2KnkTOPTIm3vQ53GsVYb8ajktGxvjWNzBeI2-OPXhwhvBKG14m_EON8t3_6fiB6PKsoFU474LLHilOr4TwOUh_oYjv_-5Ej5x1Je6XMHnsKkDCCmO1tzKoGZnoFgXxov12dZ84U374q5zwLzngPk7BC2Q0G7wIFbwf1Xm5ECSHXFHT_17iaRhu2s5eQ6B1dgx9RJBXjN-cgqZQIeNptbqXH67I3LaM_JcbKfrpx7KbDWivvKrfeWTyBJuJ9t8WD7k_4lfbbb4HUMKM761MgiIMv7GAZ8sItqU3Rw"
        },
        {
		  "kid": "encrypter",
          "p": "_aSA0u5saMEl1hc9-Sglp9LDOeZcgs_Gw7Olxefs77bIjMQpFwrFsIWR4HH6K9nscTIAKNM9AVq30Y1TTB0idebzPbjECB90KgYa3hm2g4A6pHkaOuHs0RGTWbWavDUkQka-CSB8hE7sTNSrmDpG8FbLihuSzDFWCdLGsDqXeuk",
          "kty": "RSA",
          "q": "gvBSWfBZtjHBqhwxXdO5k9J0nNqPta-sBuKc7PNhODbr0UWNkHcailKWs3f0ViXaSRAEW-EB9Ty4plgMBjy-ycc4va1Rfg-6Pn_tnVYbB5-4nmHO8vAFZR4EP4MHipyizJfNPuSlawLNc71Eo5lAUWzPRpTBZ1XvQ9AZgx-wA2M",
          "d": "et4yFr71HRMW2epVzYPNcNfqGJqTU7NsbVMCSH-ZDJ_ysPn5CgTmAK-NZh2hJvra4RCBgpOQiEYqEqX5jc3xPZyTUtCTJwRpgVNLnhylk031hy22qA2QqWRsWGLBxRvgP8gb9intIs6MkrIiPkO2t5o3J9OYpF7aO40mXH5CM2EJm-FxqGuKMVb_zWVqImmh3mqC2GlPBsiZLcHeFIbtGopsel07nngBBSmCOf7XAmtqYvZAGiJQkd1poI7p_c5n7x3aj1jPGShVLzfLBWqNipoZk0GfbY7qTlkY6dT2x098V_MSpSip9tkQ__whdHOlR5GE_HT0vlmhfwixZKaTEQ",
          "e": "AQAB",
          "use": "enc",
          "qi": "eiD4hKfSwXUVN8q14yL2JK4rUt0heIZ93CHVtkonsA8VasPOI1E6D51WaFRHaJxgvn7CiY16h2qg9xjP1uMBNcuscSKRnyqAeGJuyPh576-FWxJlZSqh9PoSxj4eHQMCWmBBi7TL820hrgA2mhc0KLekCRT36-89-Va7G74N5A8",
          "dp": "n1NJNLZd1MOXD8-Tt0HXvX6v8VvZurXnhiD_vbw84isv-PRzVy0GFycgBhuyaP8__a7J2NswE_y3QOOEcmhOsD79hkTcprmTT558HA2MzzeqHoyPxHMMPhvLMmvYIedDunoTf0ovzTCCUJS6oSniS7BJtJwzbx6CjDMhaau0YZk",
          "alg": "RSA-OAEP",
          "dq": "G1DXXTvu-ztWE47eHZzV0ijNewt9f4GueaE865G6bmfGulmwNrsiJkkkdzxHFNHAwA0_W4uNRQPt4YXsvEBf7OhKxgcqQQo26GL3xyL3cJe5hBETg0rfVUD10eob4Kbcr6Hbh4tblv92rPaHIzoNWO9CLo9J6azbxWHccKZjqdE",
          "n": "gbulO7BqCAKwVy3ZqrR033OM1Mp-SqOViwD1manyHjhDSB5dPLL8AG9zdl8hoQwQO8TVR4Ske2oYLkr9zxtWROTYKvF6Ssp0W5Df-sE6lEnMRqPr0GNrIubA0i2I0-uuK26N-x2_KJZbrMviH8qAdQGKopJ1-9DTvgXbOZmzQDuP3s0V8BB7pSroOaBpE7wtKAr5akPElbw_XR7m5ocmbd2TIHu8kdLU4W60Aha7x427KaYhetbtVkkS3h6j7FP9Wm2iMSkneo2ZA0WP4N4jqv3wqA2c7d_IeQNWmUxFrIoApmhy4MoMMDXjmWM_7JwH1UK6RsaknAfT7C0YJjVDGw"
        }
      ]
    }`), &jwks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal openid provider jwks: %w", err)
	}

	op, err := provider.New(goidc.ProfileFAPI1, AuthHost, func(_ context.Context) (goidc.JSONWebKeySet, error) {
		return jwks, nil
	})
	if err != nil {
		return nil, err
	}

	opts := []provider.Option{
		provider.WithClientStorage(oidc.NewClientManager(clientService)),
		provider.WithAuthnSessionStorage(oidc.NewAuthnSessionManager(db)),
		provider.WithGrantSessionStorage(oidc.NewGrantSessionManager(db)),
		provider.WithScopes(scopes...),
		provider.WithTokenOptions(oidc.TokenOptionsFunc()),
		provider.WithAuthorizationCodeGrant(),
		provider.WithImplicitGrant(),
		provider.WithRefreshTokenGrant(func(_ context.Context, _ *goidc.Client, _ goidc.GrantInfo) bool { return true }, 3600),
		provider.WithClientCredentialsGrant(),
		provider.WithTokenAuthnMethods(goidc.ClientAuthnPrivateKeyJWT),
		provider.WithPrivateKeyJWTSignatureAlgs(goidc.PS256),
		provider.WithMTLS(AuthMTLSHost, oidc.ClientCert),
		provider.WithTLSCertTokenBindingRequired(),
		provider.WithPAR(oidc.HandlePARSessionFunc(), 60),
		provider.WithUnregisteredRedirectURIsForPAR(),
		provider.WithJAR(goidc.PS256),
		provider.WithJAREncryption(goidc.RSA_OAEP),
		provider.WithJARContentEncryptionAlgs(goidc.A256GCM),
		provider.WithJARM(goidc.PS256),
		provider.WithIssuerResponseParameter(),
		provider.WithPKCE(goidc.CodeChallengeMethodSHA256),
		provider.WithACRs(oidc.ACROpenInsuranceLOA2, oidc.ACROpenInsuranceLOA3),
		provider.WithUserInfoSignatureAlgs(goidc.PS256),
		provider.WithUserInfoEncryption(goidc.RSA_OAEP),
		provider.WithUserInfoContentEncryptionAlgs(goidc.A256GCM),
		provider.WithIDTokenSignatureAlgs(goidc.PS256),
		provider.WithIDTokenEncryption(goidc.RSA_OAEP),
		provider.WithIDTokenContentEncryptionAlgs(goidc.A256GCM),
		provider.WithHandleGrantFunc(oidc.HandleGrantFunc(op, consentService)),
		provider.WithPolicies(oidc.Policies(
			AuthHost,
			userService,
			consentService,
			autoService,
			capitalizationTitleService,
			financialAssistanceService,
			acceptanceAndBranchesAbroadService,
			financialRiskService,
			housingService,
			lifePensionService,
			patrimonialService,
		)...),
		provider.WithNotifyErrorFunc(oidc.LogError),
		provider.WithDCR(oidc.DCRFunc(oidc.DCRConfig{
			Scopes:       scopes,
			KeyStoreHost: KeyStoreHost,
			SSIssuer:     SoftwareStatementIssuer,
		}), nil),
	}
	if err := op.WithOptions(opts...); err != nil {
		return nil, err
	}

	return op, nil
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, api.CtxKeyCorrelationID, uuid.NewString())
		if fapiID := r.Header.Get("X-Fapi-Interaction-Id"); fapiID != "" {
			ctx = context.WithValue(ctx, api.CtxKeyInteractionID, fapiID)
		}
		slog.InfoContext(ctx, "request received", "method", r.Method, "path", r.URL.Path)

		start := timeutil.DateTimeNow()
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "error", rec, "stack", string(debug.Stack()))
				api.WriteError(w, r, fmt.Errorf("internal error: %v", rec))
			}
			slog.InfoContext(ctx, "request completed", slog.Duration("duration", time.Since(start.Time)))
		}()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
