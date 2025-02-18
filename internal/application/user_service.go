package application

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	ErrCannotFollowSelf = errors.New("cannot follow self")
)

// DynamoDBUserService handles user-related operations using DynamoDB
type DynamoDBUserService struct {
	DynamoDBClient DynamoDBClient
	Ctx            context.Context
}

// NewDynamoDBUserService creates a new UserService with a DynamoDB client
func NewDynamoDBUserService(client DynamoDBClient) *DynamoDBUserService {
	return &DynamoDBUserService{
		DynamoDBClient: client,
		Ctx:            context.TODO(),
	}
}

// FollowUser allows a user to follow another user
func (s *DynamoDBUserService) FollowUser(followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	_, err := s.DynamoDBClient.PutItem(s.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String("UserFollowers"),
		Item: map[string]types.AttributeValue{
			"UserID":     &types.AttributeValueMemberS{Value: followerID},
			"FolloweeID": &types.AttributeValueMemberS{Value: followeeID},
		},
	})
	if err != nil {
		return err
	}

	_, err = s.DynamoDBClient.PutItem(s.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String("UserFollowers"),
		Item: map[string]types.AttributeValue{
			"UserID":     &types.AttributeValueMemberS{Value: followeeID},
			"FolloweeID": &types.AttributeValueMemberS{Value: followerID},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
