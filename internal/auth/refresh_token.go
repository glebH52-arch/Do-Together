package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidRefreshToken = errors.New("invalid refresh token")

func GenerateRefreshToken() (token, sessionID, tokenHash string, err error) {

	sessionID, err = generateRandomString(32)
	if err != nil {
		return "", "", "", err
	}

	secret, err := generateRandomString(32)
	if err != nil {
		return "", "", "", err
	}
	token = sessionID + "." + secret

	tokenHash = HashRefreshSecret(secret)
	return token, sessionID, tokenHash, nil
}
func ParseRefreshToken(token string) (sessionID, secret string, err error) {
	sessionID, secret, found := strings.Cut(token, ".")
	if !found {
		return "", "", ErrInvalidRefreshToken
	}
	if sessionID == "" || secret == "" {
		return "", "", ErrInvalidRefreshToken
	}
	if strings.Contains(secret, ".") {
		return "", "", ErrInvalidRefreshToken
	}
	return sessionID, secret, nil

}
func HashRefreshSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}
func generateRandomString(size int) (string, error) {
	if size <= 0 {
		return "", fmt.Errorf("size must be positive")
	}
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func GenerateRefreshTokenForSession(
	sessionID string,
) (token, tokenHash string, err error) {
	secret, err := generateRandomString(32)
	if err != nil {
		return "", "", err
	}
	token = sessionID + "." + secret

	tokenHash = HashRefreshSecret(secret)
	return token, tokenHash, nil
}
