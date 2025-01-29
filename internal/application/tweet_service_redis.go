package application

import (
	"context"
	"sort"
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
		"timestamp": time.Now().UnixNano(),
	}).Err()
	if err != nil {
		return "", err
	}

	err = s.RedisClient.RPush(s.Ctx, "user:timeline:"+userID, tweetID).Err()
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

// GetTimeline retrieves the timeline for a user
func (s *RedisTweetService) GetTimeline(userID string) ([]domain.Tweet, error) {
	// Fetch the list of followed users
	following, err := s.RedisClient.SMembers(s.Ctx, "user:following:"+userID).Result()
	if err != nil {
		return nil, err
	}

	var timeline []domain.Tweet

	// Collect tweets from each followed user
	for _, followeeID := range following {
		tweetIDs, _ := s.RedisClient.LRange(s.Ctx, "user:timeline:"+followeeID, 0, -1).Result()
		for _, tweetID := range tweetIDs {
			tweet, err := s.GetTweet(tweetID)
			if err != nil {
				continue
			}
			timeline = append(timeline, tweet)
		}
	}

	// Include the user's own tweets
	tweetIDs, _ := s.RedisClient.LRange(s.Ctx, "user:timeline:"+userID, 0, -1).Result()
	for _, tweetID := range tweetIDs {
		tweet, err := s.GetTweet(tweetID)
		if err != nil {
			continue
		}
		timeline = append(timeline, tweet)
	}

	// Sort tweets by timestamp
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Timestamp > timeline[j].Timestamp
	})

	return timeline, nil
}
