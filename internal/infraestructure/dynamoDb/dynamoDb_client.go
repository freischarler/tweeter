package dynamoDb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	UserFollowersTable = "UserFollowers"
	TweetsTable        = "Tweets"
	UserTimelinesTable = "UserTimelines"
)

type clientOptions func(*dynamodb.Options)

type DynamoConfigurator struct {
	client *dynamodb.Client
}

func GetLocalConfiguration(endpoint string) clientOptions {
	return func(options *dynamodb.Options) {
		options.Region = "us-west-2"
		options.Credentials = credentials.NewStaticCredentialsProvider("local", "local", "local")
		options.BaseEndpoint = aws.String(endpoint)
	}
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient() (*dynamodb.Client, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	endpoint := os.Getenv("DYNAMO_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://dynamodb-local:8000"
	}

	log.Printf("Using endpoint: %s", endpoint)

	client := dynamodb.NewFromConfig(cfg, GetLocalConfiguration(endpoint))

	// Test connection with retries
	maxRetries := 5
	delay := 2 * time.Second
	err = TestDynamoDBConnection(client, maxRetries, delay)
	if err != nil {
		log.Fatalf("Failed to connect to DynamoDB after %d attempts: %v", maxRetries, err)
		return nil, err
	}

	log.Println("Conectado a DynamoDB Local correctamente.")

	return client, nil
}

// TestDynamoDBConnection tests the connection to DynamoDB by listing tables with retries
func TestDynamoDBConnection(client *dynamodb.Client, maxRetries int, delay time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = client.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
		if err == nil {
			return nil
		}
		log.Printf("Failed to connect to DynamoDB (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}
	return err
}

func NewDynamoConfigurator(c *dynamodb.Client) DynamoConfigurator {
	return DynamoConfigurator{
		client: c,
	}
}

// SetupDatabase configura la base de datos, creando tablas si no existen
func (setup DynamoConfigurator) SetupDatabase() {
	setup.createTableIfNotExists(UserFollowersTable, setup.createUserFollowersTable)
	setup.createTableIfNotExists(TweetsTable, setup.createTweetsTable)
	setup.createTableIfNotExists(UserTimelinesTable, setup.createUserTimelinesTable)
}

// createTableIfNotExists verifica si una tabla existe y, si no, la crea
func (setup DynamoConfigurator) createTableIfNotExists(tableName string, createFunc func() error) {
	exists, err := setup.TableExists(tableName)
	if err != nil {
		panic(err)
	}
	if !exists {
		if err := createFunc(); err != nil {
			log.Fatalf("Error creating table %s: %v", tableName, err)
		}
	}
}

// TableExists verifica si una tabla ya existe en DynamoDB
func (setup DynamoConfigurator) TableExists(tableName string) (bool, error) {
	_, err := setup.client.DescribeTable(
		context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(tableName)},
	)
	if err != nil {
		var notFoundEx *types.ResourceNotFoundException
		if errors.As(err, &notFoundEx) {
			log.Printf("Table %v does not exist.\n", tableName)
			return false, nil
		}
		log.Printf("Couldn't determine existence of table %v. Error: %v\n", tableName, err)
		return false, err
	}
	return true, nil
}

// createTable crea una tabla en DynamoDB
func (setup DynamoConfigurator) createTable(tableName string, input *dynamodb.CreateTableInput) error {
	table, err := setup.client.CreateTable(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to create table %v: %v\n", tableName, err)
		return err
	}

	waiter := dynamodb.NewTableExistsWaiter(setup.client)
	err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName)}, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to wait on create table %v: %v\n", tableName, err)
		return err
	}

	fmt.Printf("Table %s created successfully: %v\n", tableName, table.TableDescription)
	return nil
}

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
