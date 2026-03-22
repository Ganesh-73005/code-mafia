package websocket

import (
	"code-mafia-backend/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (c *Client) handlePowerUpAttack(payload interface{}) {
	data, _ := json.Marshal(payload)
	var attack models.PowerUpAttack
	if err := json.Unmarshal(data, &attack); err != nil {
		log.Printf("Error unmarshaling power-up attack: %v", err)
		return
	}

	// Verify token
	token, err := jwt.Parse(attack.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil // This should come from config
	})

	if err != nil || !token.Valid {
		c.sendError("Invalid token")
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	// Get coins
	team, err := c.hub.repo.GetTeamByID(teamID)
	if err != nil || team == nil {
		c.sendError("Error fetching team")
		return
	}

	// Calculate cost
	deductCoins := 5
	if attack.PowerUp == "wall-breaker" {
		deductCoins = 10
	} else if attack.PowerUp == "freeze" {
		deductCoins = 15
	}

	if team.Coins < deductCoins {
		c.sendMessage("coins-error", map[string]string{"message": "Not enough coins"})
		return
	}

	// Get target client
	targetClient := c.hub.GetClientByID(attack.TargetUserID)
	if targetClient == nil {
		c.sendError("Target user not found")
		return
	}

	targetUsername := targetClient.username

	// Check if shield is active
	hasShield, _ := c.hub.redis.PowerUpExists(targetUsername, "shield")

	if hasShield && attack.PowerUp == "wall-breaker" {
		if team.Coins < 10 {
			c.sendMessage("coins-error", map[string]string{"message": "Not enough coins"})
			return
		}
		c.hub.redis.DeletePowerUp(targetUsername, "shield")
		targetClient.sendMessage("shield-down", map[string]string{"message": fmt.Sprintf("%s took down your shield", attack.From)})
	} else if attack.PowerUp == "innocency" {
		if hasShield {
			c.hub.redis.DeletePowerUp(targetUsername, "shield")
			c.sendMessage("shield-down", map[string]string{"message": "You dropped your shield"})
			c.hub.repo.UpdateTeamCoins(teamID, team.Coins+8)
		} else {
			c.sendMessage("blocked-by-shield", map[string]string{"message": "You have no active shield!"})
		}
		return
	} else if hasShield && attack.PowerUp != "shield" {
		c.sendMessage("blocked-by-shield", map[string]string{"message": fmt.Sprintf("%s has an active shield!", targetUsername)})
		return
	} else if hasShield && attack.PowerUp == "shield" {
		c.sendMessage("blocked-by-shield", map[string]string{"message": "Shield already active"})
		return
	}

	// Set power-up
	expiryTime := 180 * time.Second
	if attack.PowerUp == "shield" {
		expiryTime = 300 * time.Second
	}

	powerUpData := map[string]string{
		"from":    attack.From,
		"powerUp": attack.PowerUp,
	}
	powerUpJSON, _ := json.Marshal(powerUpData)
	c.hub.redis.SetPowerUp(targetUsername, attack.PowerUp, string(powerUpJSON), expiryTime)

	// Send to target
	targetClient.sendMessage("receive power-up", map[string]string{
		"powerUp": attack.PowerUp,
		"from":    attack.From,
	})

	// Update coins
	c.hub.repo.UpdateTeamCoins(teamID, team.Coins-deductCoins)
}

func (c *Client) handleSuicideAttack(payload interface{}) {
	data, _ := json.Marshal(payload)
	var attack models.SuicideAttack
	if err := json.Unmarshal(data, &attack); err != nil {
		log.Printf("Error unmarshaling suicide attack: %v", err)
		return
	}

	// Verify token
	token, err := jwt.Parse(attack.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})

	if err != nil || !token.Valid {
		c.sendError("Invalid token")
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	team, err := c.hub.repo.GetTeamByID(teamID)
	if err != nil || team == nil {
		c.sendError("Error fetching team")
		return
	}

	if team.Coins < 5 {
		c.sendMessage("coins-error", map[string]string{"message": "Not enough coins"})
		return
	}

	targetClient := c.hub.GetClientByID(attack.TargetUserID)
	currentClient := c.hub.GetClientByID(attack.CurrentUserID)

	if targetClient == nil || currentClient == nil {
		c.sendError("User not found")
		return
	}

	targetHasShield, _ := c.hub.redis.PowerUpExists(targetClient.username, "shield")
	currentHasShield, _ := c.hub.redis.PowerUpExists(currentClient.username, "shield")

	if targetHasShield && currentHasShield {
		c.hub.redis.DeletePowerUp(targetClient.username, "shield")
		c.hub.redis.DeletePowerUp(currentClient.username, "shield")
		targetClient.sendMessage("shield-down", map[string]string{"message": fmt.Sprintf("%s has an active shield!", targetClient.username)})
		c.sendMessage("shield-down", map[string]string{"message": "You dropped your shield"})
		targetClient.sendMessage("receive power-up", map[string]string{"powerUp": "suicide-bomber", "from": attack.From})
		c.hub.repo.UpdateTeamCoins(teamID, team.Coins-10)
	} else {
		c.sendMessage("blocked-by-shield", map[string]string{"message": "Suicide bomber attack failed"})
	}
}

func (c *Client) handleGetActivePowerUps() {
	keys, err := c.hub.redis.GetAllPowerUpsForUser(c.username)
	if err != nil {
		log.Printf("Error getting active power-ups: %v", err)
		return
	}

	var activePowerUps []models.ActivePowerUp
	for _, key := range keys {
		parts := splitKey(key)
		if len(parts) < 3 {
			continue
		}
		powerUp := parts[2]

		data, err := c.hub.redis.GetPowerUp(c.username, powerUp)
		if err != nil {
			continue
		}

		var powerUpData map[string]string
		if err := json.Unmarshal([]byte(data), &powerUpData); err != nil {
			continue
		}

		ttl, _ := c.hub.redis.GetPowerUpTTL(c.username, powerUp)
		if ttl <= 0 {
			ttl = 180
		}

		activePowerUps = append(activePowerUps, models.ActivePowerUp{
			PowerUp:       powerUpData["powerUp"],
			From:          powerUpData["from"],
			RemainingTime: ttl,
		})
	}

	c.sendMessage("apply-active-powerups", activePowerUps)
}

func (c *Client) sendMessage(msgType string, payload interface{}) {
	message := map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	}
	data, _ := json.Marshal(message)
	select {
	case c.send <- data:
	default:
		log.Printf("Failed to send message to client %s", c.username)
	}
}

func (c *Client) sendError(message string) {
	c.sendMessage("error", map[string]string{"message": message})
}

func splitKey(key string) []string {
	return splitString(key, ":")
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}
