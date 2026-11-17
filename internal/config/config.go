package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL              string
	JWTSecret                string
	AccessTokenTTL           time.Duration
	RedisAddr                string
	RedisPassword            string
	RedisDB                  int
	RefreshTokenIdleTTL      time.Duration
	RefreshTokenAbsoluteTTL  time.Duration
	EmailVerificationSecret  string
	EmailVerificationCodeTTL time.Duration
}

func Load() (*Config, error) {
	emailVerificationSecret := os.Getenv(
		"EMAIL_VERIFICATION_SECRET",
	)
	if len(emailVerificationSecret) < 32 {
		return nil, fmt.Errorf(
			"EMAIL_VERIFICATION_SECRET must be at least 32 characters",
		)
	}

	emailVerificationCodeTTLText := os.Getenv(
		"EMAIL_VERIFICATION_CODE_TTL",
	)
	emailVerificationCodeTTL, err := time.ParseDuration(
		emailVerificationCodeTTLText,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse EMAIL_VERIFICATION_CODE_TTL: %w",
			err,
		)
	}
	if emailVerificationCodeTTL <= 0 {
		return nil, fmt.Errorf(
			"EMAIL_VERIFICATION_CODE_TTL must be positive",
		)
	}
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
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR  is required")
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	if err != nil {
		return nil, fmt.Errorf("REDIS_DB must be nubmers")
	}
	if redisDB < 0 {
		return nil, fmt.Errorf("REDIS_DB must be non-negative")
	}
	refreshTokenIdleTTLText := os.Getenv("REFRESH_TOKEN_IDLE_TTL")
	refreshTokenIdleTTL, err := time.ParseDuration(refreshTokenIdleTTLText)
	if err != nil {
		return nil, fmt.Errorf("parse REFRESH_TOKEN_IDLE_TTL: %w", err)
	}
	if refreshTokenIdleTTL <= 0 {
		return nil, fmt.Errorf("REFRESH_TOKEN_IDLE_TTL must be positive")
	}
	refreshTokenAbsoluteTTLText := os.Getenv("REFRESH_TOKEN_ABSOLUTE_TTL")
	refreshTokenAbsoluteTTL, err := time.ParseDuration(refreshTokenAbsoluteTTLText)
	if err != nil {
		return nil, fmt.Errorf("parse REFRESH_TOKEN_ABSOLUTE_TTL: %w", err)
	}
	if refreshTokenAbsoluteTTL <= 0 {
		return nil, fmt.Errorf("REFRESH_TOKEN_ABSOLUTE_TTL must be positive")
	}
	if refreshTokenIdleTTL > refreshTokenAbsoluteTTL {
		return nil, fmt.Errorf(
			"REFRESH_TOKEN_IDLE_TTL must not exceed REFRESH_TOKEN_ABSOLUTE_TTL",
		)
	}
	return &Config{
		DatabaseURL:              databaseURL,
		JWTSecret:                jwtSecret,
		AccessTokenTTL:           accessTokenTTL,
		RedisAddr:                redisAddr,
		RedisPassword:            redisPassword,
		RedisDB:                  redisDB,
		RefreshTokenIdleTTL:      refreshTokenIdleTTL,
		RefreshTokenAbsoluteTTL:  refreshTokenAbsoluteTTL,
		EmailVerificationSecret:  emailVerificationSecret,
		EmailVerificationCodeTTL: emailVerificationCodeTTL,
	}, nil
}
