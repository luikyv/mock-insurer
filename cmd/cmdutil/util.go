package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/luikyv/mock-insurer/internal/timeutil"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Environment string

const (
	LocalEnvironment Environment = "LOCAL"
)

func DB(ctx context.Context, credentials string) (*gorm.DB, error) {
	type dbSecret struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		DBName   string `json:"dbname"`
		Engine   string `json:"engine"`
		SSLMode  string `json:"sslmode"`
	}

	var secret dbSecret
	if err := json.Unmarshal([]byte(credentials), &secret); err != nil {
		return nil, fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	if secret.SSLMode == "" {
		secret.SSLMode = "require"
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
		secret.Host, secret.Port, secret.Username, secret.Password, secret.DBName, secret.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return timeutil.DateTimeNow().Time
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// EnvValue retrieves an environment variable or returns a fallback value if not found.
func EnvValue[T ~string](key, fallback T) T {
	if value, exists := os.LookupEnv(string(key)); exists {
		return T(value)
	}
	return fallback
}

// PointerOf returns a pointer to the given value.
func PointerOf[T any](value T) *T {
	return &value
}
