package application

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
)

func TestFollowUser(t *testing.T) {
	mockDynamoDBClient := &MockDynamoDBClient{
		PutItemFunc: func(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	service := NewDynamoDBUserService(mockDynamoDBClient)

	t.Run("should follow user successfully", func(t *testing.T) {
		err := service.FollowUser("1", "2")
		assert.NoError(t, err)
	})

	t.Run("should return error if user tries to follow self", func(t *testing.T) {
		err := service.FollowUser("1", "1")
		assert.Error(t, err)
		assert.Equal(t, ErrCannotFollowSelf, err)
	})
}
