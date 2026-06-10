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

// StylistRepository persists model.Stylist items.
type StylistRepository struct {
	db *database.Client
}

// NewStylistRepository creates a StylistRepository.
func NewStylistRepository(db *database.Client) *StylistRepository {
	return &StylistRepository{db: db}
}

// List returns all stylists (optionally only active ones), sorted by name.
func (r *StylistRepository) List(ctx context.Context, activeOnly bool) ([]model.Stylist, error) {
	out, err := r.db.DB.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.db.Tables.Stylists),
	})
	if err != nil {
		return nil, fmt.Errorf("scan stylists: %w", err)
	}
	var stylists []model.Stylist
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &stylists); err != nil {
		return nil, fmt.Errorf("unmarshal stylists: %w", err)
	}
	result := make([]model.Stylist, 0, len(stylists))
	for i := range stylists {
		if activeOnly && !stylists[i].IsActive {
			continue
		}
		stylists[i].Normalize()
		result = append(result, stylists[i])
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// GetByID returns a stylist by id, or nil if not found.
func (r *StylistRepository) GetByID(ctx context.Context, id string) (*model.Stylist, error) {
	out, err := r.db.DB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Tables.Stylists),
		Key: map[string]dbtypes.AttributeValue{
			"id": &dbtypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get stylist %s: %w", id, err)
	}
	if out.Item == nil {
		return nil, nil
	}
	var stylist model.Stylist
	if err := attributevalue.UnmarshalMap(out.Item, &stylist); err != nil {
		return nil, fmt.Errorf("unmarshal stylist %s: %w", id, err)
	}
	stylist.Normalize()
	return &stylist, nil
}

// Put creates or replaces a stylist.
func (r *StylistRepository) Put(ctx context.Context, stylist *model.Stylist) error {
	av, err := attributevalue.MarshalMap(stylist)
	if err != nil {
		return fmt.Errorf("marshal stylist: %w", err)
	}
	_, err = r.db.DB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.Tables.Stylists),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("put stylist: %w", err)
	}
	return nil
}

// Delete removes a stylist by id.
func (r *StylistRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.DB.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.Tables.Stylists),
		Key: map[string]dbtypes.AttributeValue{
			"id": &dbtypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return fmt.Errorf("delete stylist %s: %w", id, err)
	}
	return nil
}
