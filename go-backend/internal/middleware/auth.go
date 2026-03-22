package middleware

import (
	"code-mafia-backend/internal/models"
	"code-mafia-backend/internal/redis"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

func VerifyToken(secretKey string, redisClient *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "No token provided")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization header")
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				respondWithError(w, http.StatusForbidden, "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondWithError(w, http.StatusForbidden, "Invalid token claims")
				return
			}

			username, ok := claims["username"].(string)
			if !ok {
				respondWithError(w, http.StatusForbidden, "Invalid token claims")
				return
			}

			// Verify token in Redis
			savedToken, err := redisClient.GetToken(username)
			if err != nil || savedToken != tokenString {
				respondWithError(w, http.StatusUnauthorized, "Session invalid or expired. Please login again.")
				return
			}

			// Extract user info from claims
			userClaims := models.JWTClaims{
				Username: username,
				TeamID:   claims["team_id"].(string),
				TeamName: claims["team_name"].(string),
				Role:     claims["role"].(string),
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminVerify(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "Token missing")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization header")
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				respondWithError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			role, ok := claims["role"].(string)
			if !ok || role != "admin" {
				respondWithError(w, http.StatusForbidden, "Admin access required")
				return
			}

			// Extract user info from claims
			userClaims := models.JWTClaims{
				Username: claims["username"].(string),
				TeamID:   claims["team_id"].(string),
				TeamName: claims["team_name"].(string),
				Role:     role,
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func GetUserFromContext(ctx context.Context) (*models.JWTClaims, bool) {
	user, ok := ctx.Value(UserContextKey).(models.JWTClaims)
	return &user, ok
}
