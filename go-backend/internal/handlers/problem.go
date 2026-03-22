package handlers

import (
	"code-mafia-backend/internal/database"
	"code-mafia-backend/internal/middleware"
	"code-mafia-backend/internal/models"
	"code-mafia-backend/internal/redis"
	"encoding/json"
	"net/http"
)

type ProblemHandler struct {
	repo  *database.Repository
	redis *redis.Client
}

func NewProblemHandler(repo *database.Repository, redisClient *redis.Client) *ProblemHandler {
	return &ProblemHandler{
		repo:  repo,
		redis: redisClient,
	}
}

func (h *ProblemHandler) GetProblems(w http.ResponseWriter, r *http.Request) {
	// Try to get from cache first
	cachedData, err := h.redis.Get("challenges-user")
	if err == nil && cachedData != "" {
		var challenges []models.Challenge
		if err := json.Unmarshal([]byte(cachedData), &challenges); err == nil {
			respondWithJSON(w, http.StatusOK, map[string]interface{}{"qs": challenges})
			return
		}
	}

	// Fallback to database
	challenges, err := h.repo.GetAllChallenges()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching challenges")
		return
	}

	// Filter hidden test cases
	for i := range challenges {
		filteredTestCases := []models.TestCase{}
		for _, tc := range challenges[i].TestCases {
			if tc.Type != "hidden" {
				filteredTestCases = append(filteredTestCases, tc)
			}
		}
		challenges[i].TestCases = filteredTestCases
	}

	// Cache the result
	cacheData, _ := json.Marshal(challenges)
	h.redis.Set("challenges-user", string(cacheData), 0)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"qs": challenges})
}

func (h *ProblemHandler) GetChallengesSolvedStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	submissions, err := h.repo.GetSubmissionsByTeam(user.TeamID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching submissions")
		return
	}

	// Create status map
	statusMap := make(map[string]map[string]interface{})
	for _, sub := range submissions {
		if _, exists := statusMap[sub.ChallengeID]; !exists {
			statusMap[sub.ChallengeID] = map[string]interface{}{
				"status": sub.Status,
				"code":   sub.Code,
			}
		}
	}

	respondWithJSON(w, http.StatusOK, statusMap)
}
