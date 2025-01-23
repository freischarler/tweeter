package redis

import (
	"os"

	"github.com/go-redis/redis/v8"
)

// NewRedisClient creates a new Redis client
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),     // Redis address from environment variable
		Password: os.Getenv("REDIS_PASSWORD"), // Password from environment variable
		DB:       0,                           // Default database
	})
}
