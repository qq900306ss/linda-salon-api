package model

// Service represents a salon service item (e.g. haircut, perm).
type Service struct {
	ID              string `json:"id" dynamodbav:"id"`
	Name            string `json:"name" dynamodbav:"name"`
	Description     string `json:"description" dynamodbav:"description"`
	Category        string `json:"category" dynamodbav:"category"`
	DurationMinutes int    `json:"durationMinutes" dynamodbav:"durationMinutes"`
	Price           int    `json:"price" dynamodbav:"price"`
	ImageURL        string `json:"imageUrl" dynamodbav:"imageUrl"`
	IsActive        bool   `json:"isActive" dynamodbav:"isActive"`
	SortOrder       int    `json:"sortOrder" dynamodbav:"sortOrder"`
}
