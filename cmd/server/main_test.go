package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	adapterHttp "github.com/freischarler/desafio-twitter/internal/adapters/http"
	"github.com/freischarler/desafio-twitter/internal/application"
	"github.com/freischarler/desafio-twitter/internal/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

type MockDynamoDBClient struct{}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	// Mock implementation
	return &dynamodb.GetItemOutput{}, nil
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	// Mock implementation
	return &dynamodb.PutItemOutput{}, nil
}

func (m *MockDynamoDBClient) Query(ctx context.Context, input *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	// Mock implementation
	return &dynamodb.QueryOutput{}, nil
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	// Mock implementation
	return &dynamodb.UpdateItemOutput{}, nil
}

type MockRedisClient struct{}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	// Mock implementation
	return redis.NewStringResult("", nil)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	// Mock implementation
	return redis.NewStatusResult("", nil)
}

func TestMain(m *testing.M) {
	// Set up mock environment variables
	os.Setenv("PORT", "8080")

	// Run the tests
	code := m.Run()

	// Clean up
	os.Unsetenv("PORT")

	os.Exit(code)
}

func TestServerSetup(t *testing.T) {
	dynamoDBClient := &MockDynamoDBClient{}
	redisClient := &MockRedisClient{}

	tweetService := application.NewDynamoRedisTweetService(dynamoDBClient, redisClient)
	userService := application.NewDynamoDBUserService(dynamoDBClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/tweet", adapterHttp.PostTweet(tweetService))
	mux.HandleFunc("/follow", adapterHttp.FollowUser(userService))
	mux.HandleFunc("/timeline/", adapterHttp.Timeline(tweetService))

	rateLimitMiddleware := middleware.RateLimitMiddleware(time.Minute, 100)
	handler := rateLimitMiddleware(mux)

	server := httptest.NewServer(handler)
	defer server.Close()

	// Test POST /tweet
	req, err := http.NewRequest("POST", server.URL+"/tweet", nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test POST /follow
	req, err = http.NewRequest("POST", server.URL+"/follow", nil)
	assert.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test GET /timeline/
	resp, err = http.Get(server.URL + "/timeline/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
