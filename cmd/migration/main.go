package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/mock-insurer/cmd/cmdutil"
	"github.com/luikyv/mock-insurer/internal/client"
	"github.com/luikyv/mock-insurer/internal/oidc"
	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/gorm"
)

var (
	Env              = cmdutil.EnvValue("ENV", cmdutil.LocalEnvironment)
	OrgID            = cmdutil.EnvValue("ORG_ID", "00000000-0000-0000-0000-000000000000")
	DBCredentials    = cmdutil.EnvValue("DB_CREDENTIALS", `{"username":"admin","password":"pass","host":"database.local","port":5432,"dbname":"mockinsurer","sslmode":"disable"}`)
	DBMigrationsPath = cmdutil.EnvValue("DB_MIGRATIONS_PATH", "file://db/migrations")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.Info("setting up db migration and seeding", "env", Env)

	slog.Info("connecting to database")
	db, err := cmdutil.DB(ctx, DBCredentials)
	if err != nil {
		slog.Error("failed connecting to database", "error", err)
		os.Exit(1)
	}
	slog.Info("successfully connected to database")

	slog.Info("running database migrations")
	if err := runMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations completed successfully")

	slog.Info("seeding database")
	if err := seedDatabase(ctx, db); err != nil {
		slog.Error("failed to seed database", "error", err)
		os.Exit(1)
	}
	slog.Info("database seeding completed successfully")
}

func runMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(DBMigrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	slog.Info("running migrations")
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no migrations to run")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("migrations completed successfully")
	return nil
}

func seedDatabase(ctx context.Context, db *gorm.DB) error {
	if err := seedUsuario1(ctx, db); err != nil {
		return fmt.Errorf("failed to seed usuario1: %w", err)
	}

	if Env == cmdutil.LocalEnvironment {
		if err := seedOAuthClients(ctx, db); err != nil {
			return fmt.Errorf("failed to create OAuth client: %w", err)
		}
	}

	return nil
}

func seedOAuthClients(ctx context.Context, db *gorm.DB) error {
	scopes := "openid consents consent resources customers insurance-auto quote-auto quote-auto-lead dynamic-fields"
	testClientOne := &client.Client{
		ID: "client_one",
		Data: goidc.Client{
			ID: "client_one",
			ClientMeta: goidc.ClientMeta{
				Name:                 "Client One",
				RedirectURIs:         []string{"https://localhost.emobix.co.uk:8443/test/a/mockinsurer/callback"},
				GrantTypes:           []goidc.GrantType{"authorization_code", "client_credentials", "implicit", "refresh_token"},
				ResponseTypes:        []goidc.ResponseType{"code id_token"},
				PublicJWKSURI:        "https://keystore.local/00000000-0000-0000-0000-000000000000/11111111-1111-1111-1111-111111111111/application.jwks",
				ScopeIDs:             scopes,
				IDTokenKeyEncAlg:     "RSA-OAEP",
				IDTokenContentEncAlg: "A256GCM",
				TokenAuthnMethod:     goidc.ClientAuthnPrivateKeyJWT,
				TokenAuthnSigAlg:     goidc.PS256,
				CustomAttributes: map[string]any{
					oidc.OrgIDKey: OrgID,
				},
			},
		},
		Name:      "Client One",
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testClientOne).Error; err != nil {
		return fmt.Errorf("failed to save test client one: %w", err)
	}

	testClientTwo := &client.Client{
		ID: "client_two",
		Data: goidc.Client{
			ID: "client_two",
			ClientMeta: goidc.ClientMeta{
				Name:                 "Client Two",
				RedirectURIs:         []string{"https://localhost.emobix.co.uk:8443/test/a/mockinsurer/callback"},
				GrantTypes:           []goidc.GrantType{"authorization_code", "client_credentials", "implicit", "refresh_token"},
				ResponseTypes:        []goidc.ResponseType{"code id_token"},
				PublicJWKSURI:        "https://keystore.local/00000000-0000-0000-0000-000000000000/22222222-2222-2222-2222-222222222222/application.jwks",
				ScopeIDs:             scopes,
				IDTokenKeyEncAlg:     "RSA-OAEP",
				IDTokenContentEncAlg: "A256GCM",
				TokenAuthnMethod:     goidc.ClientAuthnPrivateKeyJWT,
				TokenAuthnSigAlg:     goidc.PS256,
				CustomAttributes: map[string]any{
					oidc.OrgIDKey: OrgID,
				},
			},
		},
		Name:      "Client Two",
		OrgID:     OrgID,
		UpdatedAt: timeutil.DateTimeNow(),
	}
	if err := db.WithContext(ctx).Omit("CreatedAt").Save(testClientTwo).Error; err != nil {
		return fmt.Errorf("failed to save test client two: %w", err)
	}

	return nil
}

func pointerOf[T any](v T) *T {
	return &v
}

func mustParseBrazilDate(s string) timeutil.BrazilDate {
	date, _ := timeutil.ParseBrazilDate(s)
	return date
}
