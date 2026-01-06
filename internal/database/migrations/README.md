# Database Migrations

這個資料夾包含所有的資料庫 migration 檔案。

## Migration 命名規則

每個 migration 檔案應該使用以下命名格式：
```
v{version}_{description}.go
```

例如：
- `v1_make_user_fields_nullable.go`
- `v2_add_payment_table.go`
- `v3_add_user_settings.go`

## 如何新增 Migration

### 1. 建立新的 migration 檔案

在 `internal/database/migrations/` 資料夾中建立新檔案，例如 `v2_add_new_feature.go`：

```go
package migrations

import (
	"log"
	"gorm.io/gorm"
)

// V2AddNewFeature 新增新功能的資料庫修改
func V2AddNewFeature(tx *gorm.DB) error {
	log.Println("  [V2] Adding new feature...")

	// 在這裡執行 SQL 或使用 GORM 操作
	if err := tx.Exec("ALTER TABLE users ADD COLUMN new_field VARCHAR(255)").Error; err != nil {
		return err
	}

	log.Println("    - Added new_field column")
	return nil
}
```

### 2. 註冊 migration

在 `internal/database/migration.go` 的 `migrationList` 中新增你的 migration：

```go
var migrationList = []struct {
	version string
	name    string
	fn      MigrationFunc
}{
	{
		version: "v1",
		name:    "make_user_fields_nullable",
		fn:      migrations.V1MakeUserFieldsNullable,
	},
	{
		version: "v2",  // 新增這裡
		name:    "add_new_feature",
		fn:      migrations.V2AddNewFeature,
	},
}
```

### 3. 部署

當應用程式啟動時，系統會自動檢查並執行所有未執行的 migration。

## Migration 追蹤

系統會在資料庫中建立 `migrations` 表來追蹤已執行的 migration：

| Column    | Type      | Description              |
|-----------|-----------|--------------------------|
| id        | uint      | Primary key              |
| version   | string    | Migration 版本 (v1, v2)   |
| name      | string    | Migration 名稱            |
| applied_at| timestamp | 執行時間                  |

## 最佳實踐

1. **不要修改已部署的 migration**：一旦 migration 已經在生產環境執行，就不應該修改它
2. **使用交易**：所有 migration 都在交易中執行，失敗會自動回滾
3. **檢查現有狀態**：在修改欄位前先檢查是否已存在，避免錯誤
4. **加入日誌**：在 migration 中加入清楚的日誌訊息
5. **測試先行**：在開發環境測試 migration 後再部署到生產環境

## 範例 Migration

### 新增欄位
```go
func V2AddColumn(tx *gorm.DB) error {
	return tx.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN DEFAULT FALSE").Error
}
```

### 修改欄位
```go
func V3ModifyColumn(tx *gorm.DB) error {
	return tx.Exec("ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500)").Error
}
```

### 新增索引
```go
func V4AddIndex(tx *gorm.DB) error {
	return tx.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error
}
```

### 新增表
```go
func V5AddTable(tx *gorm.DB) error {
	return tx.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			amount DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`).Error
}
```
