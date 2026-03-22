package handlers

import (
	"code-mafia-backend/internal/database"
	"code-mafia-backend/internal/redis"
	"encoding/json"
	"log"
	"net/http"
)

type LeaderHandler struct {
	repo  *database.Repository
	redis *redis.Client
}

func NewLeaderHandler(repo *database.Repository, redisClient *redis.Client) *LeaderHandler {
	return &LeaderHandler{
		repo:  repo,
		redis: redisClient,
	}
}

func (h *LeaderHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Try to get from cache
	cachedData, err := h.redis.GetLeaderboard()
	if err == nil && cachedData != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":` + cachedData + `}`))
		return
	}

	// Fallback to database
	teams, err := h.repo.GetAllTeams()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching leaderboard")
		return
	}

	data, _ := json.Marshal(teams)
	h.redis.SetLeaderboard(string(data))

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"data": teams})
}

func (h *LeaderHandler) UpdateLeaderboard() {
	teams, err := h.repo.GetAllTeams()
	if err != nil {
		log.Printf("Error updating leaderboard: %v", err)
		return
	}

	data, _ := json.Marshal(teams)
	if err := h.redis.SetLeaderboard(string(data)); err != nil {
		log.Printf("Error caching leaderboard: %v", err)
	}
}
