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

// UserRepository persists admin accounts.
type UserRepository struct {
	db *database.Client
}

// NewUserRepository creates a UserRepository.
func NewUserRepository(db *database.Client) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername returns a user by username, or nil if not found.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	out, err := r.db.DB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Tables.Users),
		Key: map[string]dbtypes.AttributeValue{
			"username": &dbtypes.AttributeValueMemberS{Value: username},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", username, err)
	}
	if out.Item == nil {
		return nil, nil
	}
	var user model.User
	if err := attributevalue.UnmarshalMap(out.Item, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user %s: %w", username, err)
	}
	return &user, nil
}
