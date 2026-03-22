package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Username  string    `json:"username" bson:"username"`
	Password  string    `json:"-" bson:"password"`
	TeamID    string    `json:"team_id" bson:"team_id"`
	Role      string    `json:"role" bson:"role"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type Team struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Name      string    `json:"name" bson:"name"`
	Points    int       `json:"points" bson:"points"`
	Coins     int       `json:"coins" bson:"coins"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type Challenge struct {
	ID          string     `json:"id" bson:"_id,omitempty"`
	Title       string     `json:"title" bson:"title"`
	Description string     `json:"description" bson:"description"`
	Difficulty  string     `json:"difficulty" bson:"difficulty"`
	Points      int        `json:"points" bson:"points"`
	TestCases   []TestCase `json:"test_cases" bson:"test_cases"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
}

type TestCase struct {
	Name           string `json:"name" bson:"name"`
	Type           string `json:"type" bson:"type"`
	Input          string `json:"input" bson:"input"`
	ExpectedOutput string `json:"expected_output" bson:"expected_output"`
}

type Submission struct {
	ID            string    `json:"id" bson:"_id,omitempty"`
	TeamID        string    `json:"team_id" bson:"team_id"`
	ChallengeID   string    `json:"challenge_id" bson:"challenge_id"`
	Code          string    `json:"code" bson:"code"`
	Status        string    `json:"status" bson:"status"`
	PointsAwarded int       `json:"points_awarded" bson:"points_awarded"`
	SubmittedAt   time.Time `json:"submitted_at" bson:"submitted_at"`
}

type PowerUp struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Effect      string `json:"effect" bson:"effect"`
	Cost        int    `json:"cost" bson:"cost"`
	Duration    int    `json:"duration" bson:"duration"`
}

type JWTClaims struct {
	Username string `json:"username"`
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
	Role     string `json:"role"`
}

type Judge0Submission struct {
	SourceCode     string `json:"source_code"`
	LanguageID     int    `json:"language_id"`
	Stdin          string `json:"stdin"`
	ExpectedOutput string `json:"expected_output"`
}

type Judge0Response struct {
	Token string `json:"token"`
}

type Judge0BatchResponse struct {
	Submissions []Judge0SubmissionResult `json:"submissions"`
}

type Judge0SubmissionResult struct {
	Token  string          `json:"token"`
	Stdout sql.NullString  `json:"stdout"`
	Stderr sql.NullString  `json:"stderr"`
	Status Judge0Status    `json:"status"`
	Time   sql.NullFloat64 `json:"time"`
}

type Judge0Status struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type TestResult struct {
	TestCase       string `json:"testCase"`
	Status         string `json:"status"`
	Input          string `json:"input"`
	Output         string `json:"output"`
	ExpectedOutput string `json:"expectedOutput"`
	IsCorrect      bool   `json:"isCorrect"`
	Runtime        string `json:"runtime"`
}

type SubmissionResponse struct {
	ChallengeID string       `json:"challengeId"`
	Title       string       `json:"title"`
	Results     []TestResult `json:"results"`
	Passed      int          `json:"passed"`
	Total       int          `json:"total"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type PowerUpAttack struct {
	PowerUp      string `json:"powerUp"`
	TargetUserID string `json:"targetUserID"`
	From         string `json:"from"`
	Token        string `json:"token"`
}

type SuicideAttack struct {
	TargetUserID  string `json:"targetUserID"`
	CurrentUserID string `json:"currentUserID"`
	From          string `json:"from"`
	Token         string `json:"token"`
}

type ActivePowerUp struct {
	PowerUp       string `json:"powerUp"`
	From          string `json:"from"`
	RemainingTime int    `json:"remainingTime"`
}
