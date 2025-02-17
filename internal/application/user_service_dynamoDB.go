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

// DynamoDBClient is an interface that defines the methods we need from DynamoDB
type DynamoDBClient interface {
	GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

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
