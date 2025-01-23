package application

import "github.com/freischarler/desafio-twitter/internal/domain"

// TweetService defines the interface for tweet-related operations
type TweetService interface {
	PostTweet(userID, tweet string) (string, error)
	GetTweet(tweetID string) (domain.Tweet, error)
	GetPopularTweets(limit int) ([]domain.Tweet, error)
}
