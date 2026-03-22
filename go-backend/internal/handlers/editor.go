package handlers

import (
	"bytes"
	"code-mafia-backend/internal/config"
	"code-mafia-backend/internal/database"
	"code-mafia-backend/internal/middleware"
	"code-mafia-backend/internal/models"
	"code-mafia-backend/internal/redis"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type EditorHandler struct {
	repo   *database.Repository
	redis  *redis.Client
	config *config.Config
}

func NewEditorHandler(repo *database.Repository, redisClient *redis.Client, cfg *config.Config) *EditorHandler {
	return &EditorHandler{
		repo:   repo,
		redis:  redisClient,
		config: cfg,
	}
}

type CodeSubmissionRequest struct {
	QuestionID string `json:"question_id"` // Changed to string to match problem IDs like "two-sum"
	LanguageID int    `json:"language_id"`
	SourceCode string `json:"source_code"`
}

func (h *EditorHandler) RunTestCases(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CodeSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get challenge from cache or database
	challenge, err := h.getChallenge(req.QuestionID, false)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Challenge not found")
		return
	}

	// Extract test cases
	testCases, err := h.extractTestCases(challenge, false)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error extracting test cases")
		return
	}

	// Submit to Judge0
	results, err := h.submitToJudge0(testCases, req.LanguageID, req.SourceCode)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Judge0 error: %v", err))
		return
	}

	passed := 0
	for _, result := range results {
		if result.IsCorrect {
			passed++
		}
	}

	response := models.SubmissionResponse{
		ChallengeID: challenge.ID,
		Title:       challenge.Title,
		Results:     results,
		Passed:      passed,
		Total:       len(testCases),
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *EditorHandler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CodeSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get challenge with all test cases (including hidden)
	challenge, err := h.getChallenge(req.QuestionID, true)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Challenge not found")
		return
	}

	// Extract all test cases
	testCases, err := h.extractTestCases(challenge, true)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error extracting test cases")
		return
	}

	// Submit to Judge0
	results, err := h.submitToJudge0(testCases, req.LanguageID, req.SourceCode)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Judge0 error: %v", err))
		return
	}

	passed := 0
	for _, result := range results {
		if result.IsCorrect {
			passed++
		}
	}

	// Store submission
	if err := h.storeSubmission(user.TeamID, challenge.ID, req.SourceCode, passed, len(testCases), challenge.Points); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error storing submission")
		return
	}

	response := models.SubmissionResponse{
		ChallengeID: challenge.ID,
		Title:       challenge.Title,
		Results:     results,
		Passed:      passed,
		Total:       len(testCases),
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *EditorHandler) GetPoints(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.repo.GetTeamByID(user.TeamID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching points")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]int{"points": team.Points})
}

func (h *EditorHandler) getChallenge(id string, includeHidden bool) (*models.Challenge, error) {
	// Try to get from cache first
	cacheKey := "challenges-user"
	if includeHidden {
		cacheKey = "challenges-judge0"
	}

	cachedData, err := h.redis.Get(cacheKey)
	if err == nil && cachedData != "" {
		var challenges []models.Challenge
		if err := json.Unmarshal([]byte(cachedData), &challenges); err == nil {
			for _, challenge := range challenges {
				if challenge.ID == id {
					return &challenge, nil
				}
			}
		}
	}

	// Fallback to database
	return h.repo.GetChallengeByID(id)
}

func (h *EditorHandler) extractTestCases(challenge *models.Challenge, includeHidden bool) ([]models.TestCase, error) {
	var testCases []models.TestCase
	for _, tc := range challenge.TestCases {
		if !includeHidden && tc.Type == "hidden" {
			continue
		}
		testCases = append(testCases, tc)
	}

	return testCases, nil
}

func (h *EditorHandler) submitToJudge0(testCases []models.TestCase, languageID int, sourceCode string) ([]models.TestResult, error) {
	// Prepare submissions
	var submissions []models.Judge0Submission
	for _, tc := range testCases {
		submissions = append(submissions, models.Judge0Submission{
			SourceCode:     base64.StdEncoding.EncodeToString([]byte(sourceCode)),
			LanguageID:     languageID,
			Stdin:          base64.StdEncoding.EncodeToString([]byte(tc.Input)),
			ExpectedOutput: base64.StdEncoding.EncodeToString([]byte(tc.ExpectedOutput)),
		})
	}

	// Submit batch
	batchRequest := map[string]interface{}{
		"submissions": submissions,
	}
	jsonData, _ := json.Marshal(batchRequest)

	req, err := http.NewRequest("POST", h.config.JudgeZeroAPI+"/submissions/batch?base64_encoded=true", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if h.config.XAuthToken != "" {
		req.Header.Set("X-Auth-Token", h.config.XAuthToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var batchResponse []models.Judge0Response
	if err := json.Unmarshal(body, &batchResponse); err != nil {
		return nil, err
	}

	// Get tokens
	var tokens []string
	for _, r := range batchResponse {
		tokens = append(tokens, r.Token)
	}
	tokenString := strings.Join(tokens, ",")

	// Poll for results
	return h.pollJudge0(tokenString, testCases)
}

func (h *EditorHandler) pollJudge0(tokens string, testCases []models.TestCase) ([]models.TestResult, error) {
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("judge0 request timeout")
		case <-ticker.C:
			req, _ := http.NewRequest("GET", h.config.JudgeZeroAPI+"/submissions/batch?tokens="+tokens+"&fields=*&base64_encoded=true", nil)
			if h.config.XAuthToken != "" {
				req.Header.Set("X-Auth-Token", h.config.XAuthToken)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				continue
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var batchResponse models.Judge0BatchResponse
			if err := json.Unmarshal(body, &batchResponse); err != nil {
				continue
			}

			// Check if all completed
			allCompleted := true
			for _, sub := range batchResponse.Submissions {
				if sub.Status.ID <= 2 {
					allCompleted = false
					break
				}
			}

			if allCompleted {
				return h.decodeResults(batchResponse, testCases), nil
			}
		}
	}
}

func (h *EditorHandler) decodeResults(response models.Judge0BatchResponse, testCases []models.TestCase) []models.TestResult {
	var results []models.TestResult
	for i, sub := range response.Submissions {
		tc := testCases[i]

		var output string
		if sub.Stdout.Valid {
			decoded, _ := base64.StdEncoding.DecodeString(sub.Stdout.String)
			output = string(decoded)
		}

		expectedOutput, _ := base64.StdEncoding.DecodeString(tc.ExpectedOutput)
		isHidden := tc.Type == "hidden"
		isCorrect := sub.Status.ID == 3 && strings.TrimSpace(output) == strings.TrimSpace(string(expectedOutput))

		runtime := "N/A"
		if sub.Time.Valid {
			runtime = fmt.Sprintf("%.3fs", sub.Time.Float64)
		}

		result := models.TestResult{
			TestCase:       tc.Name,
			Status:         sub.Status.Description,
			Input:          tc.Input,
			Output:         output,
			ExpectedOutput: string(expectedOutput),
			IsCorrect:      isCorrect,
			Runtime:        runtime,
		}

		if isHidden {
			result.Input = "hidden"
			result.Output = "hidden"
			result.ExpectedOutput = "hidden"
		}

		results = append(results, result)
	}
	return results
}

func (h *EditorHandler) storeSubmission(teamID string, challengeID, sourceCode string, passed, total, challengePoints int) error {
	// Determine status
	status := "Incomplete"
	if passed == total {
		status = "Accepted"
	} else if passed > 0 {
		status = "Partial"
	}

	// Calculate coins
	coins := 0
	if challengePoints == 10 {
		coins = 5
	} else if challengePoints == 20 {
		coins = 7
	} else if challengePoints == 30 {
		coins = 10
	}

	// Calculate points awarded
	pointsAwarded := (passed * challengePoints) / total

	// Check if submission exists
	existing, err := h.repo.GetSubmissionByTeamAndChallenge(teamID, challengeID)
	if err != nil {
		return err
	}

	if existing == nil {
		// Create new submission
		if err := h.repo.CreateSubmission(teamID, challengeID, sourceCode, status, pointsAwarded); err != nil {
			return err
		}

		// Update team points and coins
		team, err := h.repo.GetTeamByID(teamID)
		if err != nil {
			return err
		}

		newPoints := team.Points + pointsAwarded
		newCoins := team.Coins
		if passed == total {
			newCoins += coins
		}

		return h.repo.UpdateTeamPointsAndCoins(teamID, newPoints, newCoins)
	} else if pointsAwarded >= existing.PointsAwarded {
		// Update existing submission
		scoreDiff := pointsAwarded - existing.PointsAwarded

		if err := h.repo.UpdateSubmission(existing.ID, sourceCode, status, pointsAwarded); err != nil {
			return err
		}

		// Update team points
		team, err := h.repo.GetTeamByID(teamID)
		if err != nil {
			return err
		}

		newPoints := team.Points + scoreDiff
		newCoins := team.Coins
		if passed == total && existing.Status != "Accepted" {
			newCoins += coins
		}

		return h.repo.UpdateTeamPointsAndCoins(teamID, newPoints, newCoins)
	}

	return nil
}
