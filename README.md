# Linda Salon API

琳達髮廊後端 API - 使用 Go + Gin + PostgreSQL + AWS S3

## 技術棧

- **語言**: Go 1.21+
- **框架**: Gin (RESTful API)
- **資料庫**: PostgreSQL (AWS RDS)
- **認證**: JWT (JSON Web Tokens)
- **儲存**: AWS S3
- **部署**: AWS App Runner

## 功能

- ✅ JWT 認證系統 (Access Token + Refresh Token)
- ✅ 角色權限控制 (Admin / Customer)
- ✅ 完整的 RESTful API
- ✅ 服務管理 (CRUD)
- ✅ 設計師管理 (CRUD + 排班)
- ✅ 預約系統 (建立、查詢、取消、確認)
- ✅ 統計報表 (Dashboard、營收報表)
- ✅ 圖片上傳 (AWS S3)
- ✅ CORS 支援
- ✅ 資料庫遷移
- ✅ 優雅關機

## 專案結構

```
linda-salon-api/
├── cmd/
│   └── api/
│       └── main.go              # 主程式入口
├── config/
│   └── config.go                # 配置管理
├── internal/
│   ├── auth/
│   │   └── jwt.go               # JWT 認證
│   ├── database/
│   │   └── database.go          # 資料庫連接
│   ├── handler/
│   │   ├── auth_handler.go      # 認證 API
│   │   ├── service_handler.go   # 服務 API
│   │   ├── stylist_handler.go   # 設計師 API
│   │   ├── booking_handler.go   # 預約 API
│   │   ├── statistics_handler.go # 統計 API
│   │   └── upload_handler.go    # 上傳 API
│   ├── middleware/
│   │   ├── auth.go              # 認證中介層
│   │   ├── cors.go              # CORS 中介層
│   │   └── logger.go            # 日誌中介層
│   ├── model/
│   │   ├── user.go              # 用戶模型
│   │   ├── service.go           # 服務模型
│   │   ├── stylist.go           # 設計師模型
│   │   └── booking.go           # 預約模型
│   └── repository/
│       ├── user_repository.go   # 用戶資料層
│       ├── service_repository.go # 服務資料層
│       ├── stylist_repository.go # 設計師資料層
│       └── booking_repository.go # 預約資料層
├── .env.example                 # 環境變數範例
├── .gitignore
├── go.mod
└── README.md
```

## API 端點

### 公開端點

#### 認證
- `POST /api/v1/auth/register` - 註冊
- `POST /api/v1/auth/login` - 登入
- `POST /api/v1/auth/refresh` - 更新 Token

#### 服務
- `GET /api/v1/services` - 取得服務列表
- `GET /api/v1/services/:id` - 取得單一服務

#### 設計師
- `GET /api/v1/stylists` - 取得設計師列表
- `GET /api/v1/stylists/:id` - 取得單一設計師
- `GET /api/v1/stylists/:id/schedules` - 取得設計師排班

### 需要認證的端點

#### 用戶
- `GET /api/v1/auth/profile` - 取得個人資料

#### 預約
- `GET /api/v1/bookings` - 取得預約列表
- `GET /api/v1/bookings/:id` - 取得單一預約
- `POST /api/v1/bookings` - 建立預約
- `POST /api/v1/bookings/:id/cancel` - 取消預約

#### 上傳
- `POST /api/v1/upload/image` - 上傳圖片

### 管理員端點 (需要 admin 角色)

#### 服務管理
- `POST /api/v1/admin/services` - 新增服務
- `PUT /api/v1/admin/services/:id` - 更新服務
- `DELETE /api/v1/admin/services/:id` - 刪除服務

#### 設計師管理
- `POST /api/v1/admin/stylists` - 新增設計師
- `PUT /api/v1/admin/stylists/:id` - 更新設計師
- `DELETE /api/v1/admin/stylists/:id` - 刪除設計師
- `POST /api/v1/admin/stylists/:id/schedules` - 新增排班

#### 預約管理
- `PATCH /api/v1/admin/bookings/:id/status` - 更新預約狀態

#### 統計報表
- `GET /api/v1/admin/statistics/dashboard` - Dashboard 統計
- `GET /api/v1/admin/statistics/revenue` - 營收報表

#### 上傳管理
- `DELETE /api/v1/admin/upload/image` - 刪除圖片

## 本地開發

### 1. 安裝 Go

確保已安裝 Go 1.21 或更高版本：
```bash
go version
```

### 2. 複製環境變數

```bash
cp .env.example .env
```

編輯 `.env` 檔案，填入實際的配置：
- 資料庫密碼
- JWT Secret
- AWS Access Key
- AWS Secret Key

### 3. 安裝依賴

```bash
go mod download
```

### 4. 執行資料庫遷移

啟動程式時會自動執行 migrations。

### 5. 啟動開發伺服器

```bash
go run cmd/api/main.go
```

伺服器會在 `http://localhost:8080` 啟動。

## 建置

```bash
# 建置二進位檔案
go build -o bin/api cmd/api/main.go

# 執行
./bin/api
```

## Docker 部署

```bash
# 建置 Docker 映像
docker build -t linda-salon-api .

# 執行容器
docker run -p 8080:8080 --env-file .env linda-salon-api
```

## AWS App Runner 部署

1. 推送至 GitHub
2. 登入 AWS Console
3. 前往 App Runner
4. 建立新服務
5. 連接 GitHub repository
6. 設定環境變數（從 .env）
7. 部署

## 測試

```bash
# 執行測試
go test ./...

# 測試覆蓋率
go test -cover ./...
```

## 環境變數

查看 `.env.example` 以了解所有必需的環境變數。

主要配置：
- `PORT` - 伺服器端口
- `DB_HOST` - PostgreSQL 主機
- `DB_PASSWORD` - 資料庫密碼
- `JWT_SECRET` - JWT 密鑰
- `AWS_ACCESS_KEY_ID` - AWS Access Key
- `AWS_SECRET_ACCESS_KEY` - AWS Secret Key
- `S3_BUCKET` - S3 儲存桶名稱

## 安全性

- 密碼使用 bcrypt 加密
- JWT Token 認證
- 角色權限控制
- SQL 注入防護 (GORM)
- CORS 設定
- 檔案上傳限制 (類型、大小)

## 授權

MIT License
