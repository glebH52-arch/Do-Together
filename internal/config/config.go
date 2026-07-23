package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	AccessTokenTTL time.Duration
}

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	accessTokenTTLText := os.Getenv("ACCESS_TOKEN_TTL")
	accessTokenTTL, err := time.ParseDuration(accessTokenTTLText)
	if err != nil {
		return nil, fmt.Errorf("parse ACCESS_TOKEN_TTL: %w", err)
	}
	if accessTokenTTL <= 0 {
		return nil, fmt.Errorf("ACCESS_TOKEN_TTL must be positive")
	}
	return &Config{
		DatabaseURL:    databaseURL,
		JWTSecret:      jwtSecret,
		AccessTokenTTL: accessTokenTTL,
	}, nil
}
