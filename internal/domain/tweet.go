package domain

// Tweet represents a tweet with a timestamp
type Tweet struct {
	TweetID   string `json:"tweetID"`
	UserID    string `json:"userID"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

const MaxTweetLength = 280
