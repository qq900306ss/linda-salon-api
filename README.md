# Linda Salon API

美髮沙龍預約系統後端 API，採用 **AWS Lambda + DynamoDB** Serverless 架構，
日常營運成本趨近於零。

## 架構

```
前端 (瀏覽器)
   │  HTTPS
   ▼
Lambda Function URL  (API Gateway v2 payload 格式)
   │
   ▼
AWS Lambda (Go, arm64, provided.al2023)
   │  cmd/lambda → ginadapter.NewV2 → Gin Router
   ▼
DynamoDB (PAY_PER_REQUEST 隨用隨付)
   ├── {prefix}-services   服務項目
   ├── {prefix}-stylists   設計師
   ├── {prefix}-bookings   預約 (GSI: date-index, phone-index)
   ├── {prefix}-users      管理員帳號 (bcrypt 雜湊)
   └── {prefix}-settings   全域設定 (單一項目 id="global")

S3 (選用)：圖片上傳，透過 Presigned PUT URL
```

- 冷啟動時會自動建立資料表（冪等），並在資料表為空時植入預設資料：
  管理員帳號、預設營業設定、5 筆範例服務與 3 位範例設計師。
- 所有回應皆使用統一封套格式：
  - 成功：`{"success": true, "data": ...}`
  - 失敗：`{"success": false, "error": {"code": "...", "message": "..."}}`

### 專案結構

```
cmd/
  lambda/      Lambda 進入點 (ginadapter V2)
  server/      本機開發伺服器 (:4000)
internal/
  app/         組裝設定、資料庫、路由
  auth/        JWT (HS256) 簽發與驗證
  database/    DynamoDB 客戶端、建表、種子資料
  handler/     Gin HTTP handlers 與路由
  middleware/  CORS、請求記錄、JWT 驗證
  model/       資料模型
  repository/  DynamoDB 存取層
  service/     純商業邏輯（時段產生、重疊判斷、統計）
scripts/
  build-lambda.ps1  Windows 打包腳本
  zip.go            產生保留 unix 執行權限的 function.zip
```

## 環境變數

| 變數 | 必填 | 預設值 | 說明 |
|------|------|--------|------|
| `JWT_SECRET` | 正式環境必填 | 開發用預設值（會輸出警告） | JWT HS256 簽章密鑰 |
| `ADMIN_USERNAME` | 否 | `admin` | 初始管理員帳號（僅首次植入時生效） |
| `ADMIN_PASSWORD` | 正式環境必填 | `linda2024`（會輸出警告） | 初始管理員密碼（僅首次植入時生效） |
| `TABLE_PREFIX` | 否 | `linda` | DynamoDB 資料表名稱前綴 |
| `S3_BUCKET` | 否 | （空） | 圖片上傳的 S3 bucket，未設定時上傳 API 回傳 `UPLOAD_NOT_CONFIGURED` |
| `ALLOWED_ORIGINS` | 否 | `*` | CORS 允許來源，逗號分隔 |
| `PORT` | 否 | `4000` | 本機開發伺服器埠號（僅 `cmd/server`） |

## 本機開發

需要具備 DynamoDB 存取權限的 AWS 憑證
（`aws configure` 或 `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_REGION` 環境變數）。

```powershell
# 啟動本機伺服器 (http://localhost:4000)
go run ./cmd/server

# 測試與靜態檢查
go test ./...
go vet ./...
```

首次啟動會自動建立資料表並植入種子資料，可用以下方式驗證：

```powershell
curl http://localhost:4000/health
curl http://localhost:4000/api/services
```

## API 路由

### 公開路由 (`/api`)

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/services` | 服務列表（僅啟用，依 sortOrder 排序） |
| GET | `/api/services/:id` | 單一服務 |
| GET | `/api/stylists` | 設計師列表（僅啟用） |
| GET | `/api/stylists/:id` | 單一設計師 |
| GET | `/api/timeslots?stylistId=&date=&serviceId=` | 可預約時段 |
| POST | `/api/bookings` | 建立預約（時段衝突回 409） |
| GET | `/api/bookings/lookup?phone=` | 以電話查詢預約（最新 20 筆） |
| POST | `/api/auth/admin/login` | 管理員登入，回傳 JWT（24 小時效期） |

### 管理路由 (`/api/admin`，需 `Authorization: Bearer {token}`)

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/admin/bookings?date=&from=&to=&status=&stylistId=` | 預約查詢（預設今日） |
| PATCH | `/api/admin/bookings/:id/status` | 更新預約狀態 |
| DELETE | `/api/admin/bookings/:id` | 取消預約（狀態改為 cancelled） |
| GET/POST | `/api/admin/services` | 服務列表（含停用）／新增 |
| PUT/DELETE | `/api/admin/services/:id` | 更新／刪除服務 |
| GET/POST | `/api/admin/stylists` | 設計師列表（含停用）／新增 |
| PUT/DELETE | `/api/admin/stylists/:id` | 更新／刪除設計師 |
| GET/PUT | `/api/admin/stylists/:id/schedule` | 取得／更新班表 |
| GET | `/api/admin/customers` | 顧客彙總（依電話分組） |
| GET | `/api/admin/statistics/dashboard` | 儀表板統計 |
| GET | `/api/admin/statistics/revenue?from=&to=` | 每日營收報表 |
| POST | `/api/admin/uploads` | 取得 S3 上傳用 Presigned URL |
| GET/PUT | `/api/admin/settings` | 取得／更新全域設定 |

## 建置與部署

### Windows

```powershell
.\scripts\build-lambda.ps1   # 產生 function.zip (linux/arm64)
```

### macOS / Linux

```bash
make zip   # 產生 function.zip
```

產生的 `function.zip` 可直接上傳至 Lambda
（Runtime: `provided.al2023`、Architecture: `arm64`、Handler: `bootstrap`）。

完整部署步驟（IAM 權限、Function URL、環境變數設定等）請參考
**AWS-DEPLOYMENT-GUIDE.md**。
