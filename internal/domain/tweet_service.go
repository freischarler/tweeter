package domain

type TweetService interface {
	PostTweet(userID, tweet string) (string, error)
	GetTweet(tweetID string) (Tweet, error)
	GetTimeline(userID string) ([]Tweet, error)
}
