package application

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

type mockDynamoDBClient struct {
	putItemOutput    *dynamodb.PutItemOutput
	getItemOutput    *dynamodb.GetItemOutput
	updateItemOutput *dynamodb.UpdateItemOutput
	err              error
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return m.putItemOutput, m.err
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.getItemOutput, m.err
}

func (m *mockDynamoDBClient) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return m.updateItemOutput, m.err
}

func TestPostTweet(t *testing.T) {
	mockClient := &mockDynamoDBClient{
		putItemOutput:    &dynamodb.PutItemOutput{},
		updateItemOutput: &dynamodb.UpdateItemOutput{},
		err:              nil,
	}
	service := NewDynamoDBTweetService(mockClient)

	tweetID, err := service.PostTweet("user1", "Hello, world!")
	assert.NoError(t, err)
	assert.NotEmpty(t, tweetID)
}

func TestGetTweet(t *testing.T) {
	mockClient := &mockDynamoDBClient{
		getItemOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"TweetID":   &types.AttributeValueMemberS{Value: "12345"},
				"UserID":    &types.AttributeValueMemberS{Value: "user1"},
				"Content":   &types.AttributeValueMemberS{Value: "Hello, world!"},
				"Timestamp": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().UnixNano(), 10)},
			},
		},
		err: nil,
	}
	service := NewDynamoDBTweetService(mockClient)

	tweet, err := service.GetTweet("12345")
	assert.NoError(t, err)
	assert.Equal(t, "12345", tweet.TweetID)
	assert.Equal(t, "user1", tweet.UserID)
	assert.Equal(t, "Hello, world!", tweet.Content)
}

func TestGetTimeline(t *testing.T) {
	mockClient := &mockDynamoDBClient{
		getItemOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"UserID": &types.AttributeValueMemberS{Value: "user1"},
				"Tweets": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "12345"},
				}},
			},
		},
		err: nil,
	}
	service := NewDynamoDBTweetService(mockClient)

	// Mock the GetTweet function to return a valid tweet
	mockClient.getItemOutput = &dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"TweetID":   &types.AttributeValueMemberS{Value: "12345"},
			"UserID":    &types.AttributeValueMemberS{Value: "user1"},
			"Content":   &types.AttributeValueMemberS{Value: "Hello, world!"},
			"Timestamp": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().UnixNano(), 10)},
		},
	}

	timeline, err := service.GetTimeline("user1")
	assert.NoError(t, err)
	assert.Len(t, timeline, 1)
	assert.Equal(t, "12345", timeline[0].TweetID)
	assert.Equal(t, "Hello, world!", timeline[0].Content)
}
