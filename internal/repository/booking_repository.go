package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/qq900306ss/linda-salon-api/internal/database"
	"github.com/qq900306ss/linda-salon-api/internal/model"
)

// MaxRangeDays caps how many days a date-range query may span.
const MaxRangeDays = 92

// BookingRepository persists model.Booking items.
type BookingRepository struct {
	db *database.Client
}

// NewBookingRepository creates a BookingRepository.
func NewBookingRepository(db *database.Client) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create stores a new booking. The top-level phone attribute is kept in sync
// with the embedded customer phone for the phone-index GSI.
func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	booking.Phone = booking.Customer.Phone
	av, err := attributevalue.MarshalMap(booking)
	if err != nil {
		return fmt.Errorf("marshal booking: %w", err)
	}
	_, err = r.db.DB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.db.Tables.Bookings),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		return fmt.Errorf("put booking: %w", err)
	}
	return nil
}

// GetByID returns a booking by id, or nil if not found.
func (r *BookingRepository) GetByID(ctx context.Context, id string) (*model.Booking, error) {
	out, err := r.db.DB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.Tables.Bookings),
		Key: map[string]dbtypes.AttributeValue{
			"id": &dbtypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get booking %s: %w", id, err)
	}
	if out.Item == nil {
		return nil, nil
	}
	var booking model.Booking
	if err := attributevalue.UnmarshalMap(out.Item, &booking); err != nil {
		return nil, fmt.Errorf("unmarshal booking %s: %w", id, err)
	}
	return &booking, nil
}

// UpdateStatus sets the status and updatedAt of an existing booking.
func (r *BookingRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.DB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.Tables.Bookings),
		Key: map[string]dbtypes.AttributeValue{
			"id": &dbtypes.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:    aws.String("SET #status = :status, updatedAt = :updatedAt"),
		ConditionExpression: aws.String("attribute_exists(id)"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]dbtypes.AttributeValue{
			":status":    &dbtypes.AttributeValueMemberS{Value: status},
			":updatedAt": &dbtypes.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		return fmt.Errorf("update booking %s status: %w", id, err)
	}
	return nil
}

// ListByDate returns all bookings on a given date via the date-index GSI,
// sorted by time ascending.
func (r *BookingRepository) ListByDate(ctx context.Context, date string) ([]model.Booking, error) {
	var bookings []model.Booking
	paginator := dynamodb.NewQueryPaginator(r.db.DB, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.Tables.Bookings),
		IndexName:              aws.String(database.BookingDateIndex),
		KeyConditionExpression: aws.String("#date = :date"),
		ExpressionAttributeNames: map[string]string{
			"#date": "date",
		},
		ExpressionAttributeValues: map[string]dbtypes.AttributeValue{
			":date": &dbtypes.AttributeValueMemberS{Value: date},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("query bookings by date %s: %w", date, err)
		}
		var batch []model.Booking
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &batch); err != nil {
			return nil, fmt.Errorf("unmarshal bookings: %w", err)
		}
		bookings = append(bookings, batch...)
	}
	return bookings, nil
}

// ListByDateRange returns bookings within [from, to] (inclusive, "YYYY-MM-DD")
// by querying the date-index one day at a time. The range is capped at
// MaxRangeDays days.
func (r *BookingRepository) ListByDateRange(ctx context.Context, from, to string) ([]model.Booking, error) {
	start, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, fmt.Errorf("invalid from date %q", from)
	}
	end, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, fmt.Errorf("invalid to date %q", to)
	}
	if end.Before(start) {
		return nil, fmt.Errorf("to date is before from date")
	}
	if end.Sub(start) > MaxRangeDays*24*time.Hour {
		return nil, fmt.Errorf("date range exceeds %d days", MaxRangeDays)
	}

	var bookings []model.Booking
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		day, err := r.ListByDate(ctx, d.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, day...)
	}
	return bookings, nil
}

// ListByPhone returns a customer's bookings via the phone-index GSI,
// newest date first, capped at limit.
func (r *BookingRepository) ListByPhone(ctx context.Context, phone string, limit int) ([]model.Booking, error) {
	out, err := r.db.DB.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.Tables.Bookings),
		IndexName:              aws.String(database.BookingPhoneIndex),
		KeyConditionExpression: aws.String("phone = :phone"),
		ExpressionAttributeValues: map[string]dbtypes.AttributeValue{
			":phone": &dbtypes.AttributeValueMemberS{Value: phone},
		},
		ScanIndexForward: aws.Bool(false), // newest date first
		Limit:            aws.Int32(int32(limit)),
	})
	if err != nil {
		return nil, fmt.Errorf("query bookings by phone: %w", err)
	}
	var bookings []model.Booking
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &bookings); err != nil {
		return nil, fmt.Errorf("unmarshal bookings: %w", err)
	}
	return bookings, nil
}

// ScanAll returns every booking in the table. Data volume is expected to be
// small (single salon), so a full scan is acceptable.
func (r *BookingRepository) ScanAll(ctx context.Context) ([]model.Booking, error) {
	var bookings []model.Booking
	paginator := dynamodb.NewScanPaginator(r.db.DB, &dynamodb.ScanInput{
		TableName: aws.String(r.db.Tables.Bookings),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("scan bookings: %w", err)
		}
		var batch []model.Booking
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &batch); err != nil {
			return nil, fmt.Errorf("unmarshal bookings: %w", err)
		}
		bookings = append(bookings, batch...)
	}
	sort.Slice(bookings, func(i, j int) bool {
		if bookings[i].Date != bookings[j].Date {
			return bookings[i].Date < bookings[j].Date
		}
		return bookings[i].Time < bookings[j].Time
	})
	return bookings, nil
}
