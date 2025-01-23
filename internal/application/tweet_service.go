package application

import (
	"context"
	"strconv"
	"time"

	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/go-redis/redis/v8"
)

// TweetService handles tweet-related operations
type TweetService struct {
	RedisClient *redis.Client
	Ctx         context.Context
}

// NewTweetService creates a new TweetService
func NewTweetService(redisClient *redis.Client) *TweetService {
	return &TweetService{
		RedisClient: redisClient,
		Ctx:         context.Background(),
	}
}

// PostTweet posts a new tweet
func (s *TweetService) PostTweet(userID, tweet string) (string, error) {
	if len(tweet) > domain.MaxTweetLength {
		return "", domain.ErrTweetTooLong
	}

	tweetID := strconv.FormatInt(s.RedisClient.Incr(s.Ctx, "tweetID:counter").Val(), 10)

	err := s.RedisClient.HSet(s.Ctx, "tweet:"+tweetID, map[string]interface{}{
		"userID":    userID,
		"content":   tweet,
		"timestamp": time.Now().Unix(),
	}).Err()
	if err != nil {
		return "", err
	}

	err = s.RedisClient.LPush(s.Ctx, "user:timeline:"+userID, tweetID).Err()
	if err != nil {
		return "", err
	}

	return tweetID, nil
}
