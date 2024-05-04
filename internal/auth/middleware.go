package auth

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware создает промежуточное ПО для аутентификации JWT
func AuthMiddleware(secretKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				http.Error(w, "Authorization token not provided", http.StatusUnauthorized)
				return
			}

			claims, err := ValidateToken(tokenString, string(secretKey))
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Добавление claims в контекст запроса
			ctx := context.WithValue(r.Context(), "userClaims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken извлекает токен JWT из заголовка Authorization
func extractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
