package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/qq900306ss/linda-salon-api/internal/database"
	"github.com/qq900306ss/linda-salon-api/internal/model"
)

// SettingsRepository persists the single global settings item.
type SettingsRepository struct {
	db *database.Client
}

// NewSettingsRepository creates a SettingsRepository.
func NewSettingsRepository(db *database.Client) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get returns the global settings, or nil if not seeded yet.
func (r *SettingsRepository) Get(ctx context.Context) (*model.Settings, error) {
	out, err := r.db.DB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Tables.Settings),
		Key: map[string]dbtypes.AttributeValue{
			"id": &dbtypes.AttributeValueMemberS{Value: model.SettingsID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}
	var settings model.Settings
	if err := attributevalue.UnmarshalMap(out.Item, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal settings: %w", err)
	}
	settings.Normalize()
	return &settings, nil
}

// Put replaces the global settings item.
func (r *SettingsRepository) Put(ctx context.Context, settings *model.Settings) error {
	settings.ID = model.SettingsID
	av, err := attributevalue.MarshalMap(settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	_, err = r.db.DB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.Tables.Settings),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("put settings: %w", err)
	}
	return nil
}
