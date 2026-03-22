package handlers

import (
	"code-mafia-backend/internal/database"
	"code-mafia-backend/internal/middleware"
	"code-mafia-backend/internal/redis"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo      *database.Repository
	redis     *redis.Client
	secretKey string
}

func NewAuthHandler(repo *database.Repository, redisClient *redis.Client, secretKey string) *AuthHandler {
	return &AuthHandler{
		repo:      repo,
		redis:     redisClient,
		secretKey: secretKey,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type VerifyResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user from database
	user, err := h.repo.GetUserByUsername(req.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if user == nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Get team info
	team, err := h.repo.GetTeamByID(user.TeamID)
	if err != nil || team == nil {
		respondWithError(w, http.StatusInternalServerError, "Team not found")
		return
	}

	// Create JWT token
	claims := jwt.MapClaims{
		"username":  user.Username,
		"team_id":   user.TeamID,
		"team_name": team.Name,
		"role":      user.Role,
		"exp":       time.Now().Add(6 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Store token in Redis
	if err := h.redis.SetToken(user.Username, tokenString, 6*time.Hour); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to store session")
		return
	}

	respondWithJSON(w, http.StatusOK, LoginResponse{Token: tokenString})
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	respondWithJSON(w, http.StatusOK, VerifyResponse{
		Valid:    true,
		Username: user.Username,
		TeamName: user.TeamName,
		Role:     user.Role,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Delete token from Redis
	if err := h.redis.DeleteToken(user.Username); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
