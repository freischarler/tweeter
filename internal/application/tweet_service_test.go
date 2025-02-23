package application

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

type MockDynamoDBClient struct {
	PutItemFunc    func(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItemFunc    func(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItemFunc func(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	QueryFunc      func(ctx context.Context, input *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return m.PutItemFunc(ctx, input, opts...)
}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.GetItemFunc(ctx, input, opts...)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return m.UpdateItemFunc(ctx, input, opts...)
}

func (m *MockDynamoDBClient) Query(ctx context.Context, input *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return m.QueryFunc(ctx, input, opts...)
}

type MockRedisClient struct {
	GetFunc func(ctx context.Context, key string) *redis.StringCmd
	SetFunc func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return m.GetFunc(ctx, key)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return m.SetFunc(ctx, key, value, expiration)
}

func TestPostTweet(t *testing.T) {
	mockDynamoDBClient := &MockDynamoDBClient{
		PutItemFunc: func(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
		UpdateItemFunc: func(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			return &dynamodb.UpdateItemOutput{}, nil
		},
	}
	mockRedisClient := &MockRedisClient{
		GetFunc: func(ctx context.Context, key string) *redis.StringCmd {
			return redis.NewStringResult("", redis.Nil)
		},
		SetFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
			return redis.NewStatusResult("", nil)
		},
	}

	service := NewDynamoRedisTweetService(mockDynamoDBClient, mockRedisClient)

	t.Run("should post tweet successfully", func(t *testing.T) {
		tweetID, err := service.PostTweet("1", "Hello World")
		assert.NoError(t, err)
		assert.NotEmpty(t, tweetID)
	})

	t.Run("should return error if tweet is too long", func(t *testing.T) {
		longTweet := make([]byte, MaxTweetLength+1)
		tweetID, err := service.PostTweet("1", string(longTweet))
		assert.Error(t, err)
		assert.Equal(t, domain.ErrTweetTooLong, err)
		assert.Empty(t, tweetID)
	})
}

func TestGetTweet(t *testing.T) {
	mockDynamoDBClient := &MockDynamoDBClient{
		GetItemFunc: func(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			if input.Key["TweetID"].(*types.AttributeValueMemberS).Value == "1" {
				return &dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{
						"TweetID":   &types.AttributeValueMemberS{Value: "1"},
						"UserID":    &types.AttributeValueMemberS{Value: "1"},
						"Content":   &types.AttributeValueMemberS{Value: "Hello World"},
						"Timestamp": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().UnixNano(), 10)},
					},
				}, nil
			}
			return &dynamodb.GetItemOutput{}, nil
		},
	}
	mockRedisClient := &MockRedisClient{}

	service := NewDynamoRedisTweetService(mockDynamoDBClient, mockRedisClient)

	t.Run("should get tweet successfully", func(t *testing.T) {
		tweet, err := service.GetTweet("1")
		assert.NoError(t, err)
		assert.Equal(t, "1", tweet.TweetID)
		assert.Equal(t, "1", tweet.UserID)
		assert.Equal(t, "Hello World", tweet.Content)
	})

	t.Run("should return error if tweet not found", func(t *testing.T) {
		tweet, err := service.GetTweet("2")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrTweetNotFound, err)
		assert.Empty(t, tweet.TweetID)
	})
}

func TestGetTimeline(t *testing.T) {
	mockDynamoDBClient := &MockDynamoDBClient{
		GetItemFunc: func(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			if input.TableName != nil && *input.TableName == "UserTimelines" {
				return &dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{
						"UserID": &types.AttributeValueMemberS{Value: "1"},
						"Tweets": &types.AttributeValueMemberL{Value: []types.AttributeValue{
							&types.AttributeValueMemberS{Value: "1"},
							&types.AttributeValueMemberS{Value: "2"},
						}},
					},
				}, nil
			}
			if input.Key["TweetID"].(*types.AttributeValueMemberS).Value == "1" {
				return &dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{
						"TweetID":   &types.AttributeValueMemberS{Value: "1"},
						"UserID":    &types.AttributeValueMemberS{Value: "1"},
						"Content":   &types.AttributeValueMemberS{Value: "Hello World"},
						"Timestamp": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().UnixNano(), 10)},
					},
				}, nil
			}
			if input.Key["TweetID"].(*types.AttributeValueMemberS).Value == "2" {
				return &dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{
						"TweetID":   &types.AttributeValueMemberS{Value: "2"},
						"UserID":    &types.AttributeValueMemberS{Value: "1"},
						"Content":   &types.AttributeValueMemberS{Value: "Hello Again"},
						"Timestamp": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().UnixNano(), 10)},
					},
				}, nil
			}
			return &dynamodb.GetItemOutput{}, nil
		},
		QueryFunc: func(ctx context.Context, input *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			if *input.TableName == "UserFollowers" {
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"UserID":     &types.AttributeValueMemberS{Value: "1"},
							"FolloweeID": &types.AttributeValueMemberS{Value: "2"},
						},
					},
				}, nil
			}
			return &dynamodb.QueryOutput{}, nil
		},
	}
	mockRedisClient := &MockRedisClient{
		GetFunc: func(ctx context.Context, key string) *redis.StringCmd {
			return redis.NewStringResult("", nil)
		},
		SetFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
			return redis.NewStatusResult("", nil)
		},
	}

	service := NewDynamoRedisTweetService(mockDynamoDBClient, mockRedisClient)

	t.Run("should get timeline successfully", func(t *testing.T) {
		mockRedisClient.GetFunc = func(ctx context.Context, key string) *redis.StringCmd {
			tweets := []domain.Tweet{
				{TweetID: "1", UserID: "1", Content: "Hello World", Timestamp: time.Now().UnixNano()},
				{TweetID: "2", UserID: "1", Content: "Hello Again", Timestamp: time.Now().UnixNano()},
			}
			tweetsJSON, _ := json.Marshal(tweets)
			return redis.NewStringResult(string(tweetsJSON), nil)
		}

		timeline, err := service.GetTimeline("1")
		assert.NoError(t, err)
		assert.Len(t, timeline, 2)
		assert.Equal(t, "1", timeline[0].TweetID)
		assert.Equal(t, "2", timeline[1].TweetID)
	})

	t.Run("should return empty timeline if no tweets found", func(t *testing.T) {
		mockRedisClient.GetFunc = func(ctx context.Context, key string) *redis.StringCmd {
			return redis.NewStringResult("", redis.Nil)
		}

		mockDynamoDBClient.GetItemFunc = func(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"UserID": &types.AttributeValueMemberS{Value: "1"},
					"Tweets": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
				},
			}, nil
		}

		timeline, err := service.GetTimeline("1")
		assert.NoError(t, err)
		assert.Empty(t, timeline)
	})

	t.Run("should handle cache hit", func(t *testing.T) {
		mockRedisClient.GetFunc = func(ctx context.Context, key string) *redis.StringCmd {
			tweets := []domain.Tweet{
				{TweetID: "1", UserID: "1", Content: "Hello World", Timestamp: time.Now().UnixNano()},
				{TweetID: "2", UserID: "1", Content: "Hello Again", Timestamp: time.Now().UnixNano()},
			}
			tweetsJSON, _ := json.Marshal(tweets)
			return redis.NewStringResult(string(tweetsJSON), nil)
		}

		timeline, err := service.GetTimeline("1")
		assert.NoError(t, err)
		assert.Len(t, timeline, 2)
		assert.Equal(t, "1", timeline[0].TweetID)
		assert.Equal(t, "2", timeline[1].TweetID)
	})

	t.Run("should get timeline with followed user's tweet", func(t *testing.T) {
		mockRedisClient.GetFunc = func(ctx context.Context, key string) *redis.StringCmd {
			return redis.NewStringResult("", redis.Nil)
		}

		mockDynamoDBClient.GetItemFunc = func(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			if input.TableName != nil && *input.TableName == "UserTimelines" {
				if input.Key["UserID"].(*types.AttributeValueMemberS).Value == "1" {
					return &dynamodb.GetItemOutput{
						Item: map[string]types.AttributeValue{
							"UserID": &types.AttributeValueMemberS{Value: "1"},
							"Tweets": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
						},
					}, nil
				}
				if input.Key["UserID"].(*types.AttributeValueMemberS).Value == "2" {
					return &dynamodb.GetItemOutput{
						Item: map[string]types.AttributeValue{
							"UserID": &types.AttributeValueMemberS{Value: "2"},
							"Tweets": &types.AttributeValueMemberL{Value: []types.AttributeValue{
								&types.AttributeValueMemberS{Value: "1"},
							}},
						},
					}, nil
				}
			}
			if input.Key["TweetID"].(*types.AttributeValueMemberS).Value == "1" {
				return &dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{
						"TweetID":   &types.AttributeValueMemberS{Value: "1"},
						"UserID":    &types.AttributeValueMemberS{Value: "2"},
						"Content":   &types.AttributeValueMemberS{Value: "Hello World"},
						"Timestamp": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().UnixNano(), 10)},
					},
				}, nil
			}
			return &dynamodb.GetItemOutput{}, nil
		}

		mockDynamoDBClient.QueryFunc = func(ctx context.Context, input *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			if *input.TableName == "UserFollowers" {
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"UserID":     &types.AttributeValueMemberS{Value: "1"},
							"FolloweeID": &types.AttributeValueMemberS{Value: "2"},
						},
					},
				}, nil
			}
			return &dynamodb.QueryOutput{}, nil
		}

		timeline, err := service.GetTimeline("1")
		assert.NoError(t, err)
		assert.Len(t, timeline, 1)
		assert.Equal(t, "1", timeline[0].TweetID)
	})
}
