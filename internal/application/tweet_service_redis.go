package application

import (
	"context"
	"strconv"
	"time"

	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/go-redis/redis/v8"
)

// RedisTweetService implements TweetService using Redis
type RedisTweetService struct {
	RedisClient *redis.Client
	Ctx         context.Context
}

// NewRedisTweetService creates a new RedisTweetService
func NewRedisTweetService(redisClient *redis.Client) *RedisTweetService {
	return &RedisTweetService{
		RedisClient: redisClient,
		Ctx:         context.Background(),
	}
}

// PostTweet posts a new tweet
func (s *RedisTweetService) PostTweet(userID, tweet string) (string, error) {
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

// GetTweet retrieves a tweet by its ID
func (s *RedisTweetService) GetTweet(tweetID string) (domain.Tweet, error) {
	tweetData, err := s.RedisClient.HGetAll(s.Ctx, "tweet:"+tweetID).Result()
	if err != nil {
		return domain.Tweet{}, err
	}

	if len(tweetData) == 0 {
		return domain.Tweet{}, domain.ErrTweetNotFound
	}

	timestamp, _ := strconv.ParseInt(tweetData["timestamp"], 10, 64)
	tweet := domain.Tweet{
		UserID:    tweetData["userID"],
		Content:   tweetData["content"],
		Timestamp: timestamp,
	}

	return tweet, nil
}

// GetPopularTweets retrieves the most popular tweets
func (s *RedisTweetService) GetPopularTweets(limit int) ([]domain.Tweet, error) {
	tweetIDs, err := s.RedisClient.ZRevRange(s.Ctx, "popular:tweets", 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	var tweets []domain.Tweet
	for _, tweetID := range tweetIDs {
		tweet, err := s.GetTweet(tweetID)
		if err != nil {
			continue
		}
		tweets = append(tweets, tweet)
	}

	return tweets, nil
}
