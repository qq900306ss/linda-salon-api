# Linda Salon — AWS 最省錢部署教學(一步一步)

這份指南帶你把整套系統部署到 AWS,目標是**每月接近 $0 美元**。

## 整體架構

```
客人 / 管理員瀏覽器
        │
        ├── CloudFront + S3  ←  beauty-salon-booking(客人預約網站,靜態檔案)
        ├── CloudFront + S3  ←  linda-salon-admin(管理後台,靜態檔案)
        │
        └── Lambda Function URL  ←  linda-salon-api(Go API)
                    │
                    ├── DynamoDB(5 張資料表,自動建立)
                    └── S3(圖片上傳,選用)
```

## 每月開銷預估

| 服務 | 免費額度 | 預估費用(小流量沙龍) |
|------|----------|------------------------|
| Lambda | 每月 100 萬次請求 + 40 萬 GB-秒(**永久免費**) | $0 |
| Lambda Function URL | 完全免費,不需 API Gateway | $0 |
| DynamoDB | 25 GB 儲存永久免費;On-Demand 讀寫每百萬次約 $0.25–$1.25 | ~$0–0.3 |
| S3(前端靜態檔 + 圖片) | 12 個月免費 5GB,之後每 GB 約 $0.025 | ~$0.1–0.3 |
| CloudFront | 每月 1 TB 流量 + 1000 萬次請求(**永久免費**) | $0 |
| CloudWatch Logs | 5 GB 免費,建議設定 7 天保留 | ~$0 |
| **總計** | | **約 US$0–1/月(約 NT$0–30)** |

> 對比原本架構:App Runner(~$25+/月)+ RDS PostgreSQL(~$15+/月)≈ **$40+/月** → 現在 **~$0–1/月**。

選用項目(要自訂網域才需要):
- Route 53 託管區域:$0.50/月
- 網域註冊:約 $12–15/年(.com)

---

## 事前準備

1. AWS 帳號(https://aws.amazon.com)
2. 區域統一選 **東京 `ap-northeast-1`**(離台灣最近)。Console 右上角確認區域。
3. 本機已安裝 Go 1.23+、Node.js 20+(你的電腦已經有了)

---

## 第一步:建立 Lambda 的 IAM 執行角色

1. AWS Console → 搜尋 **IAM** → 左側「角色 (Roles)」→「建立角色」
2. 信任實體類型:**AWS 服務**,使用案例選 **Lambda** → 下一步
3. 勾選以下受管政策:
   - `AWSLambdaBasicExecutionRole`(寫 CloudWatch 日誌)
   - `AmazonDynamoDBFullAccess`(讀寫 + 自動建表)
   - `AmazonS3FullAccess`(圖片上傳用;不用圖片上傳可不勾)
4. 角色名稱:`linda-salon-api-role` → 建立角色

> 進階(可選):正式上線後可把 FullAccess 換成只限 `linda-*` 資料表與單一 bucket 的自訂政策,更安全。

---

## 第二步:編譯 Go API 並打包

在 PowerShell 進入 API 專案資料夾:

```powershell
cd C:\Users\rr900\Desktop\linda-salon-api
git checkout rebuild
.\scripts\build-lambda.ps1
```

成功後資料夾內會出現 `function.zip`(已編譯成 Linux ARM64,並保留執行權限)。

---

## 第三步:建立 Lambda 函式

1. Console → 搜尋 **Lambda** →「建立函式」
2. 選「**從頭開始撰寫**」:
   - 函式名稱:`linda-salon-api`
   - 執行階段 (Runtime):**Amazon Linux 2023(provided.al2023)** ← Go 沒有版本限制,因為上傳的是編譯好的執行檔
   - 架構:**arm64**(比 x86 便宜 20%)
   - 執行角色:「使用現有角色」→ 選 `linda-salon-api-role`
3. 建立後,在「程式碼」頁籤 →「上傳來源」→「.zip 檔案」→ 上傳 `function.zip`
4. 「組態 (Configuration)」→「一般組態」→ 編輯:
   - 記憶體:**256 MB**
   - 逾時:**30 秒**(第一次呼叫要自動建 DynamoDB 資料表,比較久)
5. 「組態」→「環境變數」→ 新增:

   | 鍵 | 值 | 說明 |
   |----|----|------|
   | `JWT_SECRET` | 自己亂打一串 32+ 字元 | **必填**,JWT 簽章密鑰 |
   | `ADMIN_USERNAME` | `admin` | 管理員帳號 |
   | `ADMIN_PASSWORD` | 你的密碼 | **務必改掉預設值** `linda2024` |
   | `TABLE_PREFIX` | `linda` | DynamoDB 資料表前綴 |
   | `ALLOWED_ORIGINS` | 先填 `*`,之後改成兩個前端的 CloudFront 網址(逗號分隔) | CORS |
   | `S3_BUCKET` | 圖片 bucket 名稱(第六步建立後再回來填) | 選填 |

---

## 第四步:開啟 Function URL(免費的 API 端點)

1. Lambda →「組態」→「**函式 URL (Function URL)**」→「建立函式 URL」
2. 驗證類型 (Auth type):**NONE**(API 自己有 JWT 驗證)
3. **CORS 不要勾**(API 程式內已處理 CORS,兩邊都開會重複造成錯誤)
4. 建立後會得到網址,長這樣:
   `https://xxxxxxxx.lambda-url.ap-northeast-1.on.aws/`
   **把它記下來,這就是你的 API 網址。**

測試:瀏覽器開 `https://xxxxxxxx.lambda-url.ap-northeast-1.on.aws/api/services`
第一次呼叫會花 10–20 秒(自動建表 + 寫入範例資料),之後就快了。看到 `{"success":true,...}` 就成功了!

到 Console 搜尋 **DynamoDB** → 資料表,應該會看到 `linda-services`、`linda-stylists`、`linda-bookings`、`linda-users`、`linda-settings` 五張表。

---

## 第五步:建立圖片上傳的 S3 Bucket(選用)

不需要在後台上傳服務/設計師照片的話,可跳過。

1. Console → **S3** →「建立儲存貯體」
   - 名稱:`linda-salon-uploads-你的名字`(S3 名稱全球唯一)
   - 區域:東京
   - 「封鎖所有公開存取」:**取消勾選**(圖片要能被網頁讀取),確認警告
2. 建立後進入 bucket →「許可」→「儲存貯體政策」貼上(把 bucket 名稱換掉):

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "PublicReadUploads",
    "Effect": "Allow",
    "Principal": "*",
    "Action": "s3:GetObject",
    "Resource": "arn:aws:s3:::你的bucket名稱/uploads/*"
  }]
}
```

3. 同頁「跨來源資源共用 (CORS)」貼上(讓瀏覽器能直接上傳):

```json
[{
  "AllowedHeaders": ["*"],
  "AllowedMethods": ["PUT"],
  "AllowedOrigins": ["*"],
  "ExposeHeaders": []
}]
```

4. 回到 Lambda 環境變數,把 `S3_BUCKET` 填上這個 bucket 名稱。

---

## 第六步:建置兩個前端

兩個前端都要先設定 API 網址再建置。

**客人預約網站:**
```powershell
cd C:\Users\rr900\Desktop\beauty-salon-booking
git checkout rebuild
npm install
"NEXT_PUBLIC_API_URL=https://xxxxxxxx.lambda-url.ap-northeast-1.on.aws" | Out-File -Encoding utf8 .env.local
npm run build
```

**管理後台:**
```powershell
cd C:\Users\rr900\Desktop\linda-salon-admin
git checkout rebuild
npm install
"NEXT_PUBLIC_API_URL=https://xxxxxxxx.lambda-url.ap-northeast-1.on.aws" | Out-File -Encoding utf8 .env.local
npm run build
```

> 注意:API 網址結尾**不要**加 `/`。建置完成後各自的 `out/` 資料夾就是要上傳的靜態網站。

---

## 第七步:前端 S3 Bucket × 2

替兩個前端各建一個 bucket(**維持封鎖公開存取的預設**,等等用 CloudFront 安全存取):

1. S3 →「建立儲存貯體」:`linda-salon-booking-site-你的名字`(東京)
2. 再建一個:`linda-salon-admin-site-你的名字`
3. 分別進入 bucket →「上傳」→ 把對應專案 `out/` 資料夾**裡面的全部內容**拖進去上傳(是 out 裡面的檔案,不是 out 資料夾本身)

---

## 第八步:CloudFront × 2(免費 HTTPS + CDN)

先建一個網址改寫函式(兩個發行版共用):

1. Console → **CloudFront** → 左側「Functions」→「建立函式」
   - 名稱:`rewrite-index`,貼上以下程式碼後「發佈 (Publish)」:

```javascript
function handler(event) {
    var request = event.request;
    var uri = request.uri;
    if (uri.endsWith('/')) {
        request.uri += 'index.html';
    } else if (!uri.includes('.')) {
        request.uri += '/index.html';
    }
    return request;
}
```

接著替**每個**前端 bucket 建立發行版(以下做兩次):

2. CloudFront →「建立分佈 (Create distribution)」
   - 來源網域 (Origin domain):選你的前端 bucket
   - 來源存取 (Origin access):選 **Origin access control (OAC)** →「建立新的 OAC」→ 建立
   - 「檢視器通訊協定政策」:**Redirect HTTP to HTTPS**
   - 「快取政策」:CachingOptimized
   - 「函式關聯 (Function associations)」→ Viewer request → 選 **CloudFront Functions** → `rewrite-index`
   - 「預設根物件 (Default root object)」:`index.html`
   - 建立
3. 建立後 CloudFront 會顯示黃色提示「需要更新 S3 儲存貯體政策」→ 點「**複製政策**」→ 到該 S3 bucket「許可」→「儲存貯體政策」貼上儲存
4. 等待部署完成(約 5 分鐘),拿到網址:`https://dxxxxxxxx.cloudfront.net`

完成後:
- 預約網站網址 → 給客人用
- 管理後台網址 → `https://dxxxxxxxx.cloudfront.net/login/` 登入(帳密 = Lambda 環境變數設的)

最後回 Lambda 把 `ALLOWED_ORIGINS` 改成:
`https://d預約網站.cloudfront.net,https://d管理後台.cloudfront.net`

---

## 第九步:收尾(省錢 + 安全)

1. **CloudWatch 日誌保留**:Console → CloudWatch → 日誌群組 → `/aws/lambda/linda-salon-api` → 動作 → 編輯保留設定 → **7 天**(避免日誌無限累積收費)
2. **預算警報**:Console → Billing → Budgets → 建立預算 → 金額 $5/月 → 超過寄 Email 通知你(保險用)
3. 確認 `ADMIN_PASSWORD` 不是預設值

---

## 之後怎麼更新?

**更新 API:**
```powershell
cd C:\Users\rr900\Desktop\linda-salon-api
.\scripts\build-lambda.ps1
```
→ Lambda Console → 上傳新的 `function.zip`

**更新前端:**
```powershell
npm run build
```
→ 把 `out/` 內容重新上傳到 S3(覆蓋)→ CloudFront →「無效判定 (Invalidations)」→ 建立 → 路徑填 `/*`(每月 1000 條路徑免費)

---

## 疑難排解

| 問題 | 原因 / 解法 |
|------|-------------|
| 第一次呼叫 API 超過 10 秒 | 正常,冷啟動 + 自動建表。之後約 100–300ms;閒置一陣子後的第一發也會慢 1–2 秒(冷啟動),小流量免費方案的正常現象 |
| 前端顯示「無法連線」 | 檢查 `.env.local` 的 API 網址(結尾不能有 `/`)、是否重新 build 過 |
| 瀏覽器 Console 出現 CORS 錯誤 | `ALLOWED_ORIGINS` 要填 CloudFront 網址(https 開頭、無結尾斜線);Function URL 的 CORS 設定要保持關閉 |
| 登入回 401 | 帳密 = Lambda 環境變數 `ADMIN_USERNAME` / `ADMIN_PASSWORD`;改環境變數後立即生效 |
| 網頁重新整理變 404 | CloudFront Function `rewrite-index` 沒掛到 Viewer request |
| 圖片上傳失敗 | `S3_BUCKET` 環境變數、bucket CORS(PUT)、公開讀取政策三者都要設好 |
