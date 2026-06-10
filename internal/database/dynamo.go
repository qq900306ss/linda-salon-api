// Package database provides the DynamoDB client, table management and seeding.
package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Client wraps the DynamoDB client and resolved table names.
type Client struct {
	DB     *dynamodb.Client
	Tables TableNames
}

// TableNames holds the fully-prefixed table names.
type TableNames struct {
	Services string
	Stylists string
	Bookings string
	Users    string
	Settings string
}

// Booking table GSI names.
const (
	BookingDateIndex  = "date-index"
	BookingPhoneIndex = "phone-index"
)

// New creates a DynamoDB client using the default AWS credential chain.
func New(ctx context.Context, tablePrefix string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	if tablePrefix == "" {
		tablePrefix = "linda"
	}
	return &Client{
		DB: dynamodb.NewFromConfig(cfg),
		Tables: TableNames{
			Services: tablePrefix + "-services",
			Stylists: tablePrefix + "-stylists",
			Bookings: tablePrefix + "-bookings",
			Users:    tablePrefix + "-users",
			Settings: tablePrefix + "-settings",
		},
	}, nil
}

// EnsureTables idempotently creates all tables (PAY_PER_REQUEST) and waits
// for newly created tables to become ACTIVE.
func (c *Client) EnsureTables(ctx context.Context) error {
	simpleTables := []struct {
		name string
		pk   string
	}{
		{c.Tables.Services, "id"},
		{c.Tables.Stylists, "id"},
		{c.Tables.Users, "username"},
		{c.Tables.Settings, "id"},
	}

	for _, t := range simpleTables {
		input := &dynamodb.CreateTableInput{
			TableName:   aws.String(t.name),
			BillingMode: types.BillingModePayPerRequest,
			AttributeDefinitions: []types.AttributeDefinition{
				{AttributeName: aws.String(t.pk), AttributeType: types.ScalarAttributeTypeS},
			},
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String(t.pk), KeyType: types.KeyTypeHash},
			},
		}
		if err := c.ensureTable(ctx, t.name, input); err != nil {
			return err
		}
	}

	bookingsInput := &dynamodb.CreateTableInput{
		TableName:   aws.String(c.Tables.Bookings),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("date"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("time"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("phone"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String(BookingDateIndex),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("date"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("time"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String(BookingPhoneIndex),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("phone"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("date"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
	return c.ensureTable(ctx, c.Tables.Bookings, bookingsInput)
}

// ensureTable creates the table if it does not already exist.
func (c *Client) ensureTable(ctx context.Context, name string, input *dynamodb.CreateTableInput) error {
	_, err := c.DB.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(name)})
	if err == nil {
		return nil
	}
	var notFound *types.ResourceNotFoundException
	if !errors.As(err, &notFound) {
		return fmt.Errorf("describe table %s: %w", name, err)
	}

	log.Printf("Creating DynamoDB table %s ...", name)
	if _, err := c.DB.CreateTable(ctx, input); err != nil {
		var inUse *types.ResourceInUseException
		if errors.As(err, &inUse) {
			// Another cold start created it concurrently; fall through to wait.
			log.Printf("Table %s is being created concurrently", name)
		} else {
			return fmt.Errorf("create table %s: %w", name, err)
		}
	}

	waiter := dynamodb.NewTableExistsWaiter(c.DB)
	if err := waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(name)}, 2*time.Minute); err != nil {
		return fmt.Errorf("wait for table %s: %w", name, err)
	}
	log.Printf("Table %s is ACTIVE", name)
	return nil
}
