package dynamoDb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	UserFollowersTable = "UserFollowers"
	TweetsTable        = "Tweets"
	UserTimelinesTable = "UserTimelines"
)

type DynamoConfigurator struct {
	client *dynamodb.Client
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
