package redis

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

// NewRedisClient creates a new Redis client
func NewRedisClient() *redis.Client {
	// Create a new Redis client with options from environment variables
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),     // Redis address from environment variable
		Password: os.Getenv("REDIS_PASSWORD"), // Password from environment variable
		DB:       0,                           // Default database
	})

	// Establish memory limit
	err := rdb.ConfigSet(context.Background(), "maxmemory", "256mb").Err()
	if err != nil {
		log.Fatalf("Error setting maxmemory: %v", err)
	}

	// Establish memory eviction policy
	err = rdb.ConfigSet(context.Background(), "maxmemory-policy", "allkeys-lru").Err()
	if err != nil {
		log.Fatalf("Error setting maxmemory-policy: %v", err)
	}

	return rdb
}
