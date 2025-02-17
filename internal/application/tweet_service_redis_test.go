package application

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func setupTestRedisClient() *redis.Client {
	// Setup a Redis client for testing
	// You can use a mock Redis client or a real Redis instance for testing
	// Here, we assume a real Redis instance is running on localhost:6379
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func TestRedisGetTimeline(t *testing.T) {
	redisClient := setupTestRedisClient()
	tweetService := NewRedisTweetService(redisClient)

	// Clean up Redis database before running the test
	redisClient.FlushDB(context.Background())

	// Create test data
	userID := "user1"
	followeeID := "user2"

	// User1 follows User2
	redisClient.SAdd(context.Background(), "user:following:"+userID, followeeID)

	// Post tweets for User2
	tweetService.PostTweet(followeeID, "Hello from User2!")
	time.Sleep(1 * time.Second) // Ensure different timestamps
	tweetService.PostTweet(followeeID, "Another tweet from User2!")

	// Post tweets for User1
	tweetService.PostTweet(userID, "Hello from User1!")
	time.Sleep(1 * time.Second) // Ensure different timestamps
	tweetService.PostTweet(userID, "Another tweet from User1!")

	// Get timeline for User1
	timeline, err := tweetService.GetTimeline(userID)
	assert.NoError(t, err)
	assert.Len(t, timeline, 4)

	// Verify the timeline contains the correct tweets
	expectedTweets := []string{
		"Another tweet from User1!",
		"Hello from User1!",
		"Another tweet from User2!",
		"Hello from User2!",
	}

	for i, tweet := range timeline {
		assert.Equal(t, expectedTweets[i], tweet.Content)
	}
}

func TestRedisGetTimeline_NoFollowing(t *testing.T) {
	redisClient := setupTestRedisClient()
	tweetService := NewRedisTweetService(redisClient)

	// Clean up Redis database before running the test
	redisClient.FlushDB(context.Background())

	// Create test data
	userID := "user1"

	// Post tweets for User1
	tweetService.PostTweet(userID, "Hello from User1!")
	tweetService.PostTweet(userID, "Another tweet from User1!")

	// Get timeline for User1
	timeline, err := tweetService.GetTimeline(userID)
	assert.NoError(t, err)
	assert.Len(t, timeline, 2)

	// Verify the timeline contains the correct tweets
	expectedTweets := []string{
		"Another tweet from User1!",
		"Hello from User1!",
	}

	for i, tweet := range timeline {
		assert.Equal(t, expectedTweets[i], tweet.Content)
	}
}

func TestRedisGetTimeline_NoTweets(t *testing.T) {
	redisClient := setupTestRedisClient()
	tweetService := NewRedisTweetService(redisClient)

	// Clean up Redis database before running the test
	redisClient.FlushDB(context.Background())

	// Create test data
	userID := "user1"

	// Get timeline for User1
	timeline, err := tweetService.GetTimeline(userID)
	assert.NoError(t, err)
	assert.Len(t, timeline, 0)
}
