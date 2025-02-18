package dynamoDb

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

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

// GetLocalConfiguration returns the local configuration for DynamoDB
func GetLocalConfiguration(endpoint string) func(*dynamodb.Options) {
	return func(options *dynamodb.Options) {
		options.Region = "us-west-2"
		options.Credentials = credentials.NewStaticCredentialsProvider("local", "local", "local")
		options.BaseEndpoint = aws.String(endpoint)
	}
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
