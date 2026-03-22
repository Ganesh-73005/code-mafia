package handlers

import (
	"code-mafia-backend/internal/database"
	"code-mafia-backend/internal/middleware"
	"code-mafia-backend/internal/redis"
	"net/http"
)

type GameHandler struct {
	repo  *database.Repository
	redis *redis.Client
}

func NewGameHandler(repo *database.Repository, redisClient *redis.Client) *GameHandler {
	return &GameHandler{
		repo:  repo,
		redis: redisClient,
	}
}

func (h *GameHandler) GetPowers(w http.ResponseWriter, r *http.Request) {
	powerUps, err := h.repo.GetAllPowerUps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching power-ups")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"data": powerUps})
}

func (h *GameHandler) GetCoins(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.repo.GetTeamByID(user.TeamID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching coins")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]int{"coins": team.Coins})
}
