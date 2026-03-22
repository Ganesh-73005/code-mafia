package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
	ctx    context.Context
}

func NewClient(address, password string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")

	return &Client{
		client: rdb,
		ctx:    ctx,
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

// Token operations
func (c *Client) SetToken(username, token string, expiration time.Duration) error {
	key := fmt.Sprintf("token:%s", username)
	return c.client.Set(c.ctx, key, token, expiration).Err()
}

func (c *Client) GetToken(username string) (string, error) {
	key := fmt.Sprintf("token:%s", username)
	return c.client.Get(c.ctx, key).Result()
}

func (c *Client) DeleteToken(username string) error {
	key := fmt.Sprintf("token:%s", username)
	return c.client.Del(c.ctx, key).Err()
}

// PowerUp operations
func (c *Client) SetPowerUp(username, powerUp, data string, expiration time.Duration) error {
	key := fmt.Sprintf("powerup:%s:%s", username, powerUp)
	return c.client.Set(c.ctx, key, data, expiration).Err()
}

func (c *Client) GetPowerUp(username, powerUp string) (string, error) {
	key := fmt.Sprintf("powerup:%s:%s", username, powerUp)
	return c.client.Get(c.ctx, key).Result()
}

func (c *Client) DeletePowerUp(username, powerUp string) error {
	key := fmt.Sprintf("powerup:%s:%s", username, powerUp)
	return c.client.Del(c.ctx, key).Err()
}

func (c *Client) PowerUpExists(username, powerUp string) (bool, error) {
	key := fmt.Sprintf("powerup:%s:%s", username, powerUp)
	result, err := c.client.Exists(c.ctx, key).Result()
	return result > 0, err
}

func (c *Client) GetPowerUpTTL(username, powerUp string) (int, error) {
	key := fmt.Sprintf("powerup:%s:%s", username, powerUp)
	ttl, err := c.client.TTL(c.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int(ttl.Seconds()), nil
}

func (c *Client) GetAllPowerUpsForUser(username string) ([]string, error) {
	pattern := fmt.Sprintf("powerup:%s:*", username)
	return c.client.Keys(c.ctx, pattern).Result()
}

// Leaderboard operations
func (c *Client) SetLeaderboard(data string) error {
	return c.client.Set(c.ctx, "leaderboard", data, 0).Err()
}

func (c *Client) GetLeaderboard() (string, error) {
	return c.client.Get(c.ctx, "leaderboard").Result()
}

// Cache operations
func (c *Client) Set(key, value string, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, value, expiration).Err()
}

func (c *Client) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}

func (c *Client) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

func (c *Client) Exists(key string) (bool, error) {
	result, err := c.client.Exists(c.ctx, key).Result()
	return result > 0, err
}

// Pub/Sub operations
func (c *Client) Publish(channel, message string) error {
	return c.client.Publish(c.ctx, channel, message).Err()
}

func (c *Client) Subscribe(channel string) *redis.PubSub {
	return c.client.Subscribe(c.ctx, channel)
}
