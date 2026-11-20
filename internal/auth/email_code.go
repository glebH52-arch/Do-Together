package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var ErrEmailVerificationSecretTooShort = errors.New(
	"email verification secret must contain at least 32 bytes",
)

type EmailCodeManager struct {
	secret []byte
}

func NewEmailCodeManager(secret string) (*EmailCodeManager, error) {
	secret = strings.TrimSpace(secret)

	if len(secret) < 32 {
		return nil, ErrEmailVerificationSecretTooShort
	}

	return &EmailCodeManager{secret: []byte(secret)}, nil
}
func (m *EmailCodeManager) Generate() (plainCode string, codeHash string, err error) {
	plainCode, err = generateEmailCode()
	if err != nil {
		return "", "", err
	}

	codeHash = m.Hash(plainCode)

	return plainCode, codeHash, nil
}
func (m *EmailCodeManager) Hash(code string) string {
	mac := hmac.New(sha256.New, m.secret)

	_, _ = mac.Write([]byte(code))

	sum := mac.Sum(nil)

	return hex.EncodeToString(sum)
}
func (m *EmailCodeManager) Verify(code, expectedHash string) bool {
	expected, err := hex.DecodeString(expectedHash)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(code))
	actual := mac.Sum(nil)

	return hmac.Equal(actual, expected)
}

func generateEmailCode() (string, error) {
	limit := big.NewInt(1_000_000)

	number, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", number.Int64()), nil
}
