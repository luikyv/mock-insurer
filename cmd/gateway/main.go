package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
)

func main() {
	// Load and parse the directory JWKS.
	directoryKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	directoryJWK := jose.JSONWebKey{
		KeyID:     "directory_signer",
		Algorithm: string(jose.PS256),
		Key:       directoryKey,
	}

	keystoreKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	keystoreJWK := jose.JSONWebKey{
		KeyID:     "keystore_signer",
		Algorithm: string(jose.PS256),
		Key:       keystoreKey,
	}

	// Define routes.

	directoryHandler := func() http.Handler {
		mux := http.NewServeMux()

		mux.HandleFunc("GET /.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request directory openid configuration")
			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{
				"authorization_endpoint": "https://directory.local/authorize",
				"id_token_signing_alg_values_supported": [
					"PS256",
					"RS256",
					"ES256"
				],
				"issuer": "https://directory.local",
				"jwks_uri": "https://directory.local/jwks",
				"mtls_endpoint_aliases": {
					"pushed_authorization_request_endpoint": "https://matls-directory.local/par",
					"token_endpoint": "https://matls-directory.local/token"
				},
				"pushed_authorization_request_endpoint": "https://directory.local/par",
				"token_endpoint": "https://directory.local/token"
				}
			`)
		})

		mux.HandleFunc("GET /jwks", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request directory jwks")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{directoryJWK.Public()}})
		})

		mux.HandleFunc("GET /authorize", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request directory authorize")
			http.Redirect(w, r, "https://app.mockinsurer.local/api/directory/callback?code=random_code", http.StatusSeeOther)
		})

		mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request directory token")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			grantType := r.FormValue("grant_type")
			if grantType == "client_credentials" {
				_, _ = io.WriteString(w, `{
					"access_token": "random_token",
					"token_type": "bearer"
				}`)
				return
			}

			joseSigner, _ := jose.NewSigner(jose.SigningKey{
				Algorithm: jose.SignatureAlgorithm(directoryJWK.Algorithm),
				Key:       directoryJWK,
			}, (&jose.SignerOptions{}).WithType("JWT"))

			idTokenClaims := map[string]any{
				"aud":   "mockinsurer",
				"iss":   "https://directory.local",
				"nonce": "gXGldLyaaty",
				"sub":   "admin",
				"trust_framework_profile": map[string]any{
					"basic_information": map[string]any{
						"status":     "Active",
						"user_email": "admin@mail.com",
					},
					"certification_manager": false,
					"org_access_details": map[string]any{
						"00000000-0000-0000-0000-000000000000": map[string]any{
							"org_admin":               true,
							"org_registration_number": "0000000000",
							"organisation_name":       "Mock Insurer",
						},
						"11111111-1111-1111-1111-111111111111": map[string]any{
							"org_admin":               true,
							"org_registration_number": "1111111111",
							"organisation_name":       "Participant",
						},
					},
					"super_user":  true,
					"system_user": true,
				},
				"txn": "q0RwM_vzkv39zoa0nTJDDaJm_VHpHLzSheB7waKB-tT",
				"iat": time.Now().Unix(),
				"exp": time.Now().Unix() + 60,
			}

			idToken, _ := jwt.Signed(joseSigner).Claims(idTokenClaims).Serialize()
			_, _ = io.WriteString(w, fmt.Sprintf(`{
				"access_token": "random_token",
				"id_token": "%s",
				"token_type": "bearer"
			}`, idToken))
		})

		mux.HandleFunc("POST /par", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request directory par")

			_ = r.ParseForm()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{
				"request_uri": "urn:ietf:params:oauth:request_uri:random_uri",
				"expires_in": 60
			}`)
		})

		mux.HandleFunc("GET /participants", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request directory participants")
			w.Header().Set("Content-Type", "application/json")
			// TODO: Use the participant data served with http.
			http.ServeFile(w, r, "/mocks/participants.json")
		})

		mux.HandleFunc("GET /organisations/{org_id}/softwarestatements/{ss_id}/assertion", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request directory software statement")
			joseSigner, _ := jose.NewSigner(jose.SigningKey{
				Algorithm: jose.SignatureAlgorithm(keystoreJWK.Algorithm),
				Key:       keystoreJWK,
			}, (&jose.SignerOptions{}).WithType("JWT"))

			ssClaims := map[string]any{
				"iss":        "Open Banking Brasil sandbox SSA issuer",
				"org_id":     "00000000-0000-0000-0000-000000000000",
				"org_name":   "Mock Insurer",
				"org_number": "00000000000000",
				"org_status": "Active",
				"software_api_webhook_uris": []string{
					"https://localhost.emobix.co.uk:8443/test-mtls/a/mockinsurer",
				},
				"software_client_id":   "11111111-1111-1111-1111-111111111111",
				"software_client_name": "Mock Insurer Client",
				"software_environment": "Sandbox",
				"software_id":          "11111111-1111-1111-1111-111111111111",
				"software_jwks_uri":    "https://keystore.local/00000000-0000-0000-0000-000000000000/11111111-1111-1111-1111-111111111111/application.jwks",
				"software_mode":        "Live",
				"software_origin_uris": []string{
					"https://mockinsurer.local",
				},
				"software_redirect_uris": []string{
					"https://localhost.emobix.co.uk:8443/test/a/mockinsurer/callback",
				},
				"software_roles": []string{
					"DADOS",
					"PAGTO",
				},
				"software_statement_roles": []any{
					map[string]any{
						"authorisation_domain": "Open Banking Brasil ",
						"role":                 "DADOS",
						"status":               "Active",
					},
					map[string]any{
						"authorisation_domain": "Open Banking Brasil ",
						"role":                 "PAGTO",
						"status":               "Active",
					},
				},
				"software_status":  "Active",
				"software_version": "1.00",
				"iat":              time.Now().Unix(),
			}

			ssa, _ := jwt.Signed(joseSigner).Claims(ssClaims).Serialize()

			w.Header().Set("Content-Type", "application/jwt")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, ssa)
		})

		return mux
	}

	keystoreHandler := func() http.Handler {
		mux := http.NewServeMux()

		mux.HandleFunc("GET /{org_id}/{software_id}/application.jwks", func(w http.ResponseWriter, r *http.Request) {
			softwareID := r.PathValue("software_id")
			log.Printf("request keystore client jwks for %s\n", softwareID)
			w.Header().Set("Content-Type", "application/json")
			if softwareID == "11111111-1111-1111-1111-111111111111" {
				http.ServeFile(w, r, "/keys/client_one_pub.jwks")
			} else {
				http.ServeFile(w, r, "/keys/client_two_pub.jwks")
			}
		})

		mux.HandleFunc("GET /{org_id}/application.jwks", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request keystore organization jwks")
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, "/keys/org_pub.jwks")
		})

		mux.HandleFunc("GET /openbanking.jwks", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request keystore open insurance jwks")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{keystoreJWK.Public()}})
		})

		return mux
	}

	mux := http.NewServeMux()

	mux.Handle("directory.local/", directoryHandler())
	mux.Handle("matls-directory.local/", directoryHandler())
	mux.Handle("auth.sandbox.directory.opinbrasil.com.br/", directoryHandler())
	mux.Handle("matls-api.sandbox.directory.opinbrasil.com.br/", directoryHandler())
	mux.Handle("keystore.local/", keystoreHandler())
	mux.Handle("keystore.sandbox.directory.opinbrasil.com.br/", keystoreHandler())

	// Serve static files for auth.mockinsurer.local/static.
	mux.Handle("auth.mockinsurer.local/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/auth/static"))))

	// Mock Insurer backend can be accessed from the host machine for local development.
	mbHandler := reverseProxyWithFallback("host.docker.internal:80", "insurer:80")
	mux.HandleFunc("auth.mockinsurer.local/", mbHandler)
	mux.HandleFunc("matls-auth.mockinsurer.local/", mbHandler)
	mux.HandleFunc("matls-api.mockinsurer.local/", mbHandler)
	mux.HandleFunc("app.mockinsurer.local/api/", mbHandler)

	// Mock Insurer frontend can be accessed from the host machine for local development.
	mbAppHandler := reverseProxyWithFallback("host.docker.internal:8080", "app:8080")
	mux.Handle("app.mockinsurer.local/", mbAppHandler)

	// Serve participant information over HTTP because the Conformance Suite
	// does not accept self-signed certificates.
	http.HandleFunc("GET directory.local/participants", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `[
			{
				"OrganisationId": "00000000-0000-0000-0000-000000000000",
				"AuthorisationServers": [
				{
					"AuthorisationServerId": "ee6fd655-5bb3-4446-9fac-e1788d9c4049",
					"OpenIDDiscoveryDocument": "https://auth.mockinsurer.local/.well-known/openid-configuration",
					"ApiResources": [
						{
							"ApiDiscoveryEndpoints": [
								{
									"ApiEndpoint": "https://matls-api.mockinsurer.local/open-insurance/consents/v3/consents"
								},
								{
									"ApiEndpoint": "https://matls-api.mockinsurer.local/open-insurance/consents/v3/consents/{consentId}"
								}
							],
							"ApiFamilyType": "consents",
							"ApiVersion": "3.2.0",
							"Status": "Active"
						},
						{
							"ApiDiscoveryEndpoints": [
							{
								"ApiEndpoint": "https://matls-api.mockinsurer.local/open-insurance/resources/v3/resources"
							}
							],
							"ApiFamilyType": "resources",
							"ApiVersion": "3.0.0",
							"Status": "Active"
						},
						{
                        	"ApiFamilyType": "quote-auto",
                        	"ApiVersion": "1.10.0",
                        	"Status": "Active",
                        	"ApiDiscoveryEndpoints": [
                        	    {
                        	        "ApiEndpoint": "https://matls-api.mockinsurer.local/open-insurance/quote-auto/v1/lead/request"
                        	    },
                        	    {
                        	        "ApiEndpoint": "https://matls-api.mockinsurer.local/open-insurance/quote-auto/v1/lead/request/{consentId}"
                        	    },
                        	    {
                        	        "ApiEndpoint": "https://matls-api.mockinsurer.local/open-insurance/quote-auto/v1/request/{consentId}/quote-status"
                        	    },
                        	    {
                        	        "ApiEndpoint": "https://matls-api.mockinsurer.local/open-insurance/quote-auto/v1/request"
                        	    },
                        	    {
                        	        "ApiEndpoint": "https://matls-api.mockinsurer.local/open-insurance/quote-auto/v1/request/{consentId}"
                        	    }
                        	]
                    	}
					]
				}
				]
			}
		]`)
	})
	go func() {
		if err := http.ListenAndServe(":80", nil); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	caCertPEM, err := os.ReadFile("/keys/ca.crt")
	if err != nil {
		log.Fatal("Failed to read CA file:", err)
	}
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caCertPEM); !ok {
		log.Fatal("Failed to append CA certs")
	}
	serverCert, err := tls.LoadX509KeyPair("/keys/server.crt", "/keys/server.key")
	if err != nil {
		log.Fatalf("failed to load server certificate: %v", err)
	}
	server := &http.Server{
		Addr:    ":443",
		Handler: loggingMiddleware(mux),
		TLSConfig: &tls.Config{
			// Only hosts starting with "matls-" require mTLS.
			GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
				log.Printf("picking tls config for %s\n", hello.ServerName)
				cfg := &tls.Config{
					Certificates: []tls.Certificate{serverCert},
					ClientAuth:   tls.NoClientCert,
					MinVersion:   tls.VersionTLS12,
				}
				if strings.HasPrefix(hello.ServerName, "matls-") {
					log.Println("mtls is required")
					cfg.ClientAuth = tls.RequireAndVerifyClientCert
					cfg.ClientCAs = caPool
				}
				return cfg, nil
			},
		},
	}

	log.Println("starting server")
	if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("server shutdown")
}

func reverseProxyWithFallback(mainAddr, fallbackAddr string) http.HandlerFunc {

	mainURL, _ := url.Parse("http://" + mainAddr)
	mainReverseProxy := httputil.NewSingleHostReverseProxy(mainURL)
	fallbackURL, _ := url.Parse("http://" + fallbackAddr)
	fallbackReverseProxy := httputil.NewSingleHostReverseProxy(fallbackURL)

	var healthy atomic.Bool
	check := func() {
		c, err := net.DialTimeout("tcp", mainAddr, 200*time.Millisecond)
		if err != nil {
			// If we were previously healthy, log the transition to unhealthy.
			if healthy.Swap(false) {
				log.Printf("%s is unhealthy, falling back to %s\n", mainAddr, fallbackAddr)
			}
			return
		}

		_ = c.Close()
		// If we were previously unhealthy, log the transition.
		if !healthy.Swap(true) {
			log.Printf("%s is healthy\n", mainAddr)
		}
	}
	check()
	go func() {
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		for range t.C {
			check()
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			log.Println("No TLS connection established")
		} else if len(r.TLS.PeerCertificates) == 0 {
			log.Println("TLS established but no client certificate presented")
		} else {
			log.Println("Client certificate received:", r.TLS.PeerCertificates[0].Subject)
		}

		// Extract client certificate if available.
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			log.Println("client certificate found, forwarding it")
			clientCert := r.TLS.PeerCertificates[0]
			pemBytes := pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: clientCert.Raw,
			})

			r.Header.Set("X-Forwarded-Client-Cert", url.QueryEscape(string(pemBytes)))
		}

		if healthy.Load() {
			log.Printf("forwarding request to %s\n", mainAddr)
			mainReverseProxy.ServeHTTP(w, r)
		} else {
			log.Printf("falling back to %s\n", fallbackAddr)
			fallbackReverseProxy.ServeHTTP(w, r)
		}
	}

}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fapiID := r.Header.Get("X-Fapi-Interaction-Id")
		if fapiID == "" {
			fapiID = uuid.NewString()
		}

		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		slog.InfoContext(r.Context(), "incoming request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr, "interaction_id", fapiID)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		slog.InfoContext(r.Context(), "outgoing request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr, "interaction_id", fapiID, "status_code", wrapped.statusCode, "duration", duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
