package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"golang.org/x/crypto/bcrypt"

	"github.com/qq900306ss/linda-salon-api/internal/model"
)

// Seed populates the tables with initial data when they are empty:
// an admin user, default settings, sample services and sample stylists.
func (c *Client) Seed(ctx context.Context, adminUsername, adminPassword string, isDefaultPassword bool) error {
	if err := c.seedAdmin(ctx, adminUsername, adminPassword, isDefaultPassword); err != nil {
		return err
	}
	if err := c.seedSettings(ctx); err != nil {
		return err
	}
	if err := c.seedServices(ctx); err != nil {
		return err
	}
	return c.seedStylists(ctx)
}

func (c *Client) putItem(ctx context.Context, table string, item interface{}) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal item for %s: %w", table, err)
	}
	_, err = c.DB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("put item into %s: %w", table, err)
	}
	return nil
}

// tableIsEmpty checks emptiness with a 1-item scan.
func (c *Client) tableIsEmpty(ctx context.Context, table string) (bool, error) {
	out, err := c.DB.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(table),
		Limit:     aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("scan %s: %w", table, err)
	}
	return len(out.Items) == 0, nil
}

func (c *Client) seedAdmin(ctx context.Context, username, password string, isDefaultPassword bool) error {
	empty, err := c.tableIsEmpty(ctx, c.Tables.Users)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	if isDefaultPassword {
		log.Printf("WARNING: ADMIN_PASSWORD not set, seeding admin %q with the default password — change it in production!", username)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "admin",
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	log.Printf("Seeding admin user %q", username)
	return c.putItem(ctx, c.Tables.Users, user)
}

func (c *Client) seedSettings(ctx context.Context) error {
	empty, err := c.tableIsEmpty(ctx, c.Tables.Settings)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	settings := model.Settings{
		ID:                  model.SettingsID,
		SalonName:           "Linda 美髮沙龍",
		Phone:               "02-2345-6789",
		Address:             "台北市大安區忠孝東路四段 100 號 2 樓",
		OpenTime:            "10:00",
		CloseTime:           "19:00",
		SlotIntervalMinutes: 30,
		ClosedWeekdays:      []int{1}, // 週一公休
	}
	log.Println("Seeding default settings")
	return c.putItem(ctx, c.Tables.Settings, settings)
}

func (c *Client) seedServices(ctx context.Context) error {
	empty, err := c.tableIsEmpty(ctx, c.Tables.Services)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	services := []model.Service{
		{ID: "svc-cut", Name: "剪髮", Description: "由專業設計師依臉型與髮質量身打造的剪髮造型服務", Category: "hair", DurationMinutes: 60, Price: 800, IsActive: true, SortOrder: 1},
		{ID: "svc-color", Name: "染髮", Description: "使用進口染劑，提供全染、挑染與設計染等多元選擇", Category: "hair", DurationMinutes: 120, Price: 2500, IsActive: true, SortOrder: 2},
		{ID: "svc-perm", Name: "燙髮", Description: "冷燙、熱塑燙與韓系紋理燙，打造自然持久的捲度", Category: "hair", DurationMinutes: 150, Price: 3000, IsActive: true, SortOrder: 3},
		{ID: "svc-treatment", Name: "護髮", Description: "深層結構式護髮，修護受損髮質並增加光澤", Category: "care", DurationMinutes: 60, Price: 1500, IsActive: true, SortOrder: 4},
		{ID: "svc-scalp", Name: "頭皮護理", Description: "頭皮檢測與深層淨化，改善出油、敏感等頭皮問題", Category: "care", DurationMinutes: 90, Price: 1800, IsActive: true, SortOrder: 5},
	}
	log.Println("Seeding sample services")
	for _, s := range services {
		if err := c.putItem(ctx, c.Tables.Services, s); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) seedStylists(ctx context.Context) error {
	empty, err := c.tableIsEmpty(ctx, c.Tables.Stylists)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	defaultSchedule := func(workDays []int) model.Schedule {
		return model.Schedule{
			WorkDays:  workDays,
			StartTime: "10:00",
			EndTime:   "19:00",
			DaysOff:   []string{},
		}
	}
	stylists := []model.Stylist{
		{
			ID: "sty-linda", Name: "Linda", Title: "創辦人 / 首席設計師",
			Bio:             "擁有 15 年資歷的首席設計師，擅長整體造型規劃與婚宴造型。",
			Specialties:     []string{"剪髮", "染髮", "造型設計"},
			YearsExperience: 15, Rating: 4.9, IsActive: true,
			Schedule: defaultSchedule([]int{2, 3, 4, 5, 6}),
		},
		{
			ID: "sty-amy", Name: "Amy", Title: "資深設計師",
			Bio:             "專精日韓系染燙與質感髮色，客製化打造專屬風格。",
			Specialties:     []string{"染髮", "燙髮"},
			YearsExperience: 8, Rating: 4.8, IsActive: true,
			Schedule: defaultSchedule([]int{0, 2, 3, 5, 6}),
		},
		{
			ID: "sty-kevin", Name: "Kevin", Title: "設計師",
			Bio:             "擅長男士剪裁與頭皮護理，提供細緻的顧客體驗。",
			Specialties:     []string{"剪髮", "頭皮護理"},
			YearsExperience: 5, Rating: 4.7, IsActive: true,
			Schedule: defaultSchedule([]int{0, 2, 4, 5, 6}),
		},
	}
	log.Println("Seeding sample stylists")
	for _, s := range stylists {
		if err := c.putItem(ctx, c.Tables.Stylists, s); err != nil {
			return err
		}
	}
	return nil
}
