package application

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/go-redis/redis/v8"
)

var (
	MaxTweetLength = 280
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

type DynamoDBClient interface {
	PutItem(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Query(ctx context.Context, input *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// DynamoRedisTweetService implements TweetService using DynamoDB and Redis
type DynamoRedisTweetService struct {
	DynamoDBClient DynamoDBClient
	RedisClient    RedisClient
	Ctx            context.Context
}

// NewDynamoRedisTweetService creates a new DynamoRedisTweetService
func NewDynamoRedisTweetService(dynamoDBClient DynamoDBClient, redisClient RedisClient) *DynamoRedisTweetService {
	return &DynamoRedisTweetService{
		DynamoDBClient: dynamoDBClient,
		RedisClient:    redisClient,
		Ctx:            context.TODO(),
	}
}

// PostTweet posts a new tweet
func (s *DynamoRedisTweetService) PostTweet(userID, tweet string) (string, error) {
	if len(tweet) > MaxTweetLength {
		return "", domain.ErrTweetTooLong
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

	// Check if the user's timeline is in the cache
	cachedTimeline, err := s.RedisClient.Get(s.Ctx, "timeline:"+userID).Result()
	if err == nil {
		// If found in cache, update the cached timeline with the new tweet
		var timeline []domain.Tweet
		err = json.Unmarshal([]byte(cachedTimeline), &timeline)
		if err != nil {
			return "", err
		}

		// Add the new tweet to the timeline
		timeline = append(timeline, domain.Tweet{
			TweetID:   tweetID,
			UserID:    userID,
			Content:   tweet,
			Timestamp: time.Now().UnixNano(),
		})

		// Sort the timeline by timestamp
		sort.Slice(timeline, func(i, j int) bool {
			return timeline[i].Timestamp > timeline[j].Timestamp
		})

		// Cache the updated timeline in Redis
		timelineJSON, err := json.Marshal(timeline)
		if err != nil {
			return "", err
		}
		err = s.RedisClient.Set(s.Ctx, "timeline:"+userID, timelineJSON, 10*time.Minute).Err()
		if err != nil {
			return "", err
		}
	}

	return tweetID, nil
}

// GetTweet retrieves a tweet by its ID
func (s *DynamoRedisTweetService) GetTweet(tweetID string) (domain.Tweet, error) {
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
		return domain.Tweet{}, domain.ErrTweetNotFound
	}

	timestampAttr, ok := result.Item["Timestamp"].(*types.AttributeValueMemberN)
	if !ok || timestampAttr == nil {
		return domain.Tweet{}, domain.ErrTweetNotFound
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

// GetTimeline retrieves the timeline for a user
// GetTimeline retrieves the timeline for a user
func (s *DynamoRedisTweetService) GetTimeline(userID string) ([]domain.Tweet, error) {
	// Try to get the timeline from Redis cache
	cachedTimeline, err := s.RedisClient.Get(s.Ctx, "timeline:"+userID).Result()
	if err == redis.Nil {
		// Log cache miss
		log.Printf("Cache miss for user timeline: %s", userID)

		// If not found in cache, get it from DynamoDB
		timeline, err := s.getTimelineFromDynamoDB(userID)
		if err != nil {
			return nil, err
		}

		// Check if the timeline is empty
		if len(timeline) == 0 {
			return timeline, nil
		}

		// Cache the timeline in Redis
		timelineJSON, err := json.Marshal(timeline)
		if err != nil {
			return nil, err
		}
		err = s.RedisClient.Set(s.Ctx, "timeline:"+userID, timelineJSON, 10*time.Minute).Err()
		if err != nil {
			return nil, err
		}

		return timeline, nil
	} else if err != nil {
		return nil, err
	}

	// Log cache hit
	log.Printf("Cache hit for user timeline: %s", userID)

	// If found in cache, return the cached timeline
	var timeline []domain.Tweet
	err = json.Unmarshal([]byte(cachedTimeline), &timeline)
	if err != nil {
		return nil, err
	}

	return timeline, nil
}

// getTimelineFromDynamoDB retrieves the timeline for a user from DynamoDB
func (s *DynamoRedisTweetService) getTimelineFromDynamoDB(userID string) ([]domain.Tweet, error) {
	// Get the list of users the user is following
	following, err := s.getFollowing(userID)
	if err != nil {
		return nil, err
	}

	// Add the user themselves to the list
	following = append(following, userID)

	var timeline []domain.Tweet

	// Get the tweets for each user in the following list
	for _, followeeID := range following {
		result, err := s.DynamoDBClient.GetItem(s.Ctx, &dynamodb.GetItemInput{
			TableName: aws.String("UserTimelines"),
			Key: map[string]types.AttributeValue{
				"UserID": &types.AttributeValueMemberS{Value: followeeID},
			},
		})
		if err != nil {
			return nil, err
		}

		if result.Item == nil || result.Item["Tweets"] == nil {
			continue
		}

		tweetsAttr, ok := result.Item["Tweets"].(*types.AttributeValueMemberL)
		if !ok || tweetsAttr == nil || len(tweetsAttr.Value) == 0 {
			continue
		}

		tweetIDs := tweetsAttr.Value

		for _, tweetIDAttr := range tweetIDs {
			tweetID := tweetIDAttr.(*types.AttributeValueMemberS).Value
			tweet, err := s.GetTweet(tweetID)
			if err != nil {
				continue
			}
			timeline = append(timeline, tweet)
		}
	}

	// Sort the timeline by timestamp
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Timestamp > timeline[j].Timestamp
	})

	return timeline, nil
}

// getFollowing retrieves the list of users the user is following from DynamoDB
func (s *DynamoRedisTweetService) getFollowing(userID string) ([]string, error) {
	result, err := s.DynamoDBClient.Query(s.Ctx, &dynamodb.QueryInput{
		TableName:              aws.String("UserFollowers"),
		KeyConditionExpression: aws.String("UserID = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, err
	}

	var following []string
	for _, item := range result.Items {
		followeeID := item["FolloweeID"].(*types.AttributeValueMemberS).Value
		following = append(following, followeeID)
	}

	return following, nil
}
