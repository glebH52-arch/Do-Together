package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid access token")

type JWTManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewJWTManager(secret string, ttl time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
		issuer: issuer,
	}
}

func (j *JWTManager) CreateAccessToken(userID int) (string, int64, error) {
	t := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		Issuer:    j.issuer,
		IssuedAt:  jwt.NewNumericDate(t),
		ExpiresAt: jwt.NewNumericDate(t.Add(j.ttl)),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := jwtToken.SignedString(j.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}
	return tokenString, int64(j.ttl.Seconds()), nil
}

func (j *JWTManager) VerifyAccessToken(tokenString string) (int, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return j.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(j.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || token == nil || !token.Valid {
		return 0, ErrInvalidToken
	}
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}
	return userID, nil
}
