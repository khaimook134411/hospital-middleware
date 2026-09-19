package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type HISBaseURLMap map[string]string

func (h *HISBaseURLMap) Decode(value string) error {
	*h = make(HISBaseURLMap)
	if value == "" {
		return nil
	}
	for _, pair := range strings.Split(value, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid HIS_BASE_URLS entry: %q", pair)
		}
		(*h)[parts[0]] = parts[1]
	}
	return nil
}

type AppConfig struct {
	AppPort string `envconfig:"APP_PORT" default:"8080"`
	GinMode string `envconfig:"GIN_MODE" default:"release"`

	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     string `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"postgres"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"postgres"`
	DBName     string `envconfig:"DB_NAME" default:"hospital_middleware"`
	DBSSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`

	JWTSecret string        `envconfig:"JWT_SECRET" required:"true"`
	JWTTTL    time.Duration `envconfig:"JWT_TTL" default:"24h"`

	AdminUsername string `envconfig:"ADMIN_USERNAME"`
	AdminPassword string `envconfig:"ADMIN_PASSWORD"`

	HISMode     string        `envconfig:"HIS_MODE" default:"mock"`
	HISTimeout  time.Duration `envconfig:"HIS_TIMEOUT" default:"5s"`
	HISBaseURLs HISBaseURLMap `envconfig:"HIS_BASE_URLS"`
}

func GetConfig() (AppConfig, error) {
	var cfg AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
