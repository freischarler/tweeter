package dynamoDb

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// createUserFollowersTable crea la tabla UserFollowers
func (setup DynamoConfigurator) createUserFollowersTable() error {
	tableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(UserFollowersTable),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("UserID"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("FolloweeID"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("UserID"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("FolloweeID"), KeyType: types.KeyTypeRange},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}
	return setup.createTable(UserFollowersTable, tableInput)
}

// createTweetsTable crea la tabla Tweets
func (setup DynamoConfigurator) createTweetsTable() error {
	tableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(TweetsTable),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("TweetID"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("UserID"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("TweetID"), KeyType: types.KeyTypeHash},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("UserIDIndex"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("UserID"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}
	return setup.createTable(TweetsTable, tableInput)
}

// createUserTimelinesTable crea la tabla UserTimelines
func (setup DynamoConfigurator) createUserTimelinesTable() error {
	tableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(UserTimelinesTable),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("UserID"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("UserID"), KeyType: types.KeyTypeHash},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}
	return setup.createTable(UserTimelinesTable, tableInput)
}
