package application

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/freischarler/desafio-twitter/internal/domain"
)

var (
	ErrTweetTooLong  = errors.New("tweet too long")
	ErrTweetNotFound = errors.New("tweet not found")
	MaxTweetLength   = 280
)

// DynamoDBTweetService implements TweetService using DynamoDB
type DynamoDBTweetService struct {
	DynamoDBClient DynamoDBClient
	Ctx            context.Context
}

// NewDynamoDBTweetService creates a new DynamoDBTweetService
func NewDynamoDBTweetService(client DynamoDBClient) *DynamoDBTweetService {
	return &DynamoDBTweetService{
		DynamoDBClient: client,
		Ctx:            context.TODO(),
	}
}

// PostTweet posts a new tweet
func (s *DynamoDBTweetService) PostTweet(userID, tweet string) (string, error) {
	if len(tweet) > MaxTweetLength {
		return "", ErrTweetTooLong
	}

	tweetID := strconv.FormatInt(time.Now().UnixNano(), 10)

	_, err := s.DynamoDBClient.PutItem(s.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String("Tweets"),
		Item: map[string]types.AttributeValue{
			"TweetID":   &types.AttributeValueMemberS{Value: tweetID},
			"UserID":    &types.AttributeValueMemberS{Value: userID},
			"Content":   &types.AttributeValueMemberS{Value: tweet},
			"Timestamp": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().UnixNano(), 10)},
		},
	})
	if err != nil {
		return "", err
	}

	_, err = s.DynamoDBClient.UpdateItem(s.Ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("UserTimelines"),
		Key: map[string]types.AttributeValue{
			"UserID": &types.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression: aws.String("SET Tweets = list_append(if_not_exists(Tweets, :empty_list), :tweet_id)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tweet_id":   &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: tweetID}}},
			":empty_list": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		},
	})
	if err != nil {
		return "", err
	}

	return tweetID, nil
}

// GetTweet retrieves a tweet by its ID
func (s *DynamoDBTweetService) GetTweet(tweetID string) (domain.Tweet, error) {
	result, err := s.DynamoDBClient.GetItem(s.Ctx, &dynamodb.GetItemInput{
		TableName: aws.String("Tweets"),
		Key: map[string]types.AttributeValue{
			"TweetID": &types.AttributeValueMemberS{Value: tweetID},
		},
	})
	if err != nil {
		return domain.Tweet{}, err
	}

	if result.Item == nil {
		return domain.Tweet{}, ErrTweetNotFound
	}

	timestampAttr, ok := result.Item["Timestamp"].(*types.AttributeValueMemberN)
	if !ok || timestampAttr == nil {
		return domain.Tweet{}, ErrTweetNotFound
	}
	timestamp, _ := strconv.ParseInt(timestampAttr.Value, 10, 64)

	tweet := domain.Tweet{
		TweetID:   tweetID,
		UserID:    result.Item["UserID"].(*types.AttributeValueMemberS).Value,
		Content:   result.Item["Content"].(*types.AttributeValueMemberS).Value,
		Timestamp: timestamp,
	}

	return tweet, nil
}

// GetPopularTweets retrieves the most popular tweets
func (s *DynamoDBTweetService) GetPopularTweets(limit int) ([]domain.Tweet, error) {
	// Implement this function based on your criteria for "popular" tweets
	return nil, errors.New("not implemented")
}

// GetTimeline retrieves the timeline for a user
func (s *DynamoDBTweetService) GetTimeline(userID string) ([]domain.Tweet, error) {
	result, err := s.DynamoDBClient.GetItem(s.Ctx, &dynamodb.GetItemInput{
		TableName: aws.String("UserTimelines"),
		Key: map[string]types.AttributeValue{
			"UserID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil || result.Item["Tweets"] == nil {
		return nil, nil
	}

	tweetsAttr, ok := result.Item["Tweets"].(*types.AttributeValueMemberL)
	if !ok || tweetsAttr == nil || len(tweetsAttr.Value) == 0 {
		return nil, nil
	}

	tweetIDs := tweetsAttr.Value
	var timeline []domain.Tweet

	for _, tweetIDAttr := range tweetIDs {
		tweetID := tweetIDAttr.(*types.AttributeValueMemberS).Value
		tweet, err := s.GetTweet(tweetID)
		if err != nil {
			continue
		}
		timeline = append(timeline, tweet)
	}

	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Timestamp > timeline[j].Timestamp
	})

	return timeline, nil
}
