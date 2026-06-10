package repository

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/qq900306ss/linda-salon-api/internal/database"
	"github.com/qq900306ss/linda-salon-api/internal/model"
)

// ServiceRepository persists model.Service items.
type ServiceRepository struct {
	db *database.Client
}

// NewServiceRepository creates a ServiceRepository.
func NewServiceRepository(db *database.Client) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// List returns all services (optionally only active ones), sorted by sortOrder.
func (r *ServiceRepository) List(ctx context.Context, activeOnly bool) ([]model.Service, error) {
	out, err := r.db.DB.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.db.Tables.Services),
	})
	if err != nil {
		return nil, fmt.Errorf("scan services: %w", err)
	}
	var services []model.Service
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &services); err != nil {
		return nil, fmt.Errorf("unmarshal services: %w", err)
	}
	result := make([]model.Service, 0, len(services))
	for _, s := range services {
		if activeOnly && !s.IsActive {
			continue
		}
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SortOrder != result[j].SortOrder {
			return result[i].SortOrder < result[j].SortOrder
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

// GetByID returns a service by id, or nil if not found.
func (r *ServiceRepository) GetByID(ctx context.Context, id string) (*model.Service, error) {
	out, err := r.db.DB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Tables.Services),
		Key: map[string]dbtypes.AttributeValue{
			"id": &dbtypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get service %s: %w", id, err)
	}
	if out.Item == nil {
		return nil, nil
	}
	var service model.Service
	if err := attributevalue.UnmarshalMap(out.Item, &service); err != nil {
		return nil, fmt.Errorf("unmarshal service %s: %w", id, err)
	}
	return &service, nil
}

// Put creates or replaces a service.
func (r *ServiceRepository) Put(ctx context.Context, service *model.Service) error {
	av, err := attributevalue.MarshalMap(service)
	if err != nil {
		return fmt.Errorf("marshal service: %w", err)
	}
	_, err = r.db.DB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.Tables.Services),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("put service: %w", err)
	}
	return nil
}

// Delete removes a service by id.
func (r *ServiceRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.DB.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.Tables.Services),
		Key: map[string]dbtypes.AttributeValue{
			"id": &dbtypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return fmt.Errorf("delete service %s: %w", id, err)
	}
	return nil
}
