package model

// User represents an admin account stored in DynamoDB (PK: username).
type User struct {
	Username     string `json:"username" dynamodbav:"username"`
	PasswordHash string `json:"-" dynamodbav:"passwordHash"`
	Role         string `json:"role" dynamodbav:"role"`
	CreatedAt    string `json:"createdAt" dynamodbav:"createdAt"`
}
