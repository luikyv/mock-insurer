package middleware

import (
	"crypto/x509"
	"log/slog"
	"net/http"

	"github.com/luikyv/mock-insurer/internal/api"
	"github.com/luikyv/mock-insurer/internal/oidc"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

func MTLS(next http.Handler, caCertPool *x509.CertPool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "verifying client certificate")

		cert, err := oidc.ClientCert(r)
		if err != nil {
			slog.DebugContext(r.Context(), "could not get client certificate", "error", err)
			api.WriteError(w, r, api.NewError("UNAUTHORISED", http.StatusUnauthorized, "invalid certificate: could not get client certificate"))
			return
		}

		opts := x509.VerifyOptions{
			Roots:       caCertPool,
			CurrentTime: timeutil.Now(),
			KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}

		_, err = cert.Verify(opts)
		if err != nil {
			slog.DebugContext(r.Context(), "could not verify client certificate", "error", err)
			api.WriteError(w, r, api.NewError("UNAUTHORISED", http.StatusUnauthorized, "invalid certificate: could not verify client certificate"))
			return
		}

		slog.InfoContext(r.Context(), "client certificate verified successfully", "subject", cert.Subject.CommonName)
		next.ServeHTTP(w, r)
	})
}
