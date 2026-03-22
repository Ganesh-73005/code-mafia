package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	id       string
	username string
	teamID   int
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, secretKey string) {
	// Get username from query parameter or auth header
	username := r.URL.Query().Get("username")
	token := r.URL.Query().Get("token")

	if username == "" || token == "" {
		http.Error(w, "Missing username or token", http.StatusBadRequest)
		return
	}

	// Verify token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !parsedToken.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// team_id is stored as a string in JWT
	teamIDStr, ok := claims["team_id"].(string)
	if !ok {
		http.Error(w, "Invalid team_id in token", http.StatusUnauthorized)
		return
	}

	// Convert string to int
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		http.Error(w, "Invalid team_id format", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		id:       generateClientID(),
		username: username,
		teamID:   teamID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "power-up attack":
		c.handlePowerUpAttack(msg["payload"])
	case "suicide-attack":
		c.handleSuicideAttack(msg["payload"])
	case "get-active-powerups":
		c.handleGetActivePowerUps()
	}
}

func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
