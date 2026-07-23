package middleware

import (
	"context"
	"do-together/internal/auth"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		fields := strings.Fields(authorization)
		if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
			status := http.StatusUnauthorized
			http.Error(w, http.StatusText(status), status)
			return
		}
		userID, err := m.jwtManager.VerifyAccessToken(fields[1])
		if err != nil {
			status := http.StatusUnauthorized
			http.Error(w, http.StatusText(status), status)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	if ok && userID > 0 {
		return userID, true
	}
	return 0, false
}
