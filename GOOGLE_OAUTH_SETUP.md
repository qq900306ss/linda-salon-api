# Google OAuth 設定指南

## AWS App Runner 環境變數設定

登入 AWS Console → App Runner → 你的服務 → Configuration → Environment variables

新增以下環境變數：

```bash
GOOGLE_CLIENT_ID=<your-google-client-id>
GOOGLE_CLIENT_SECRET=<your-google-client-secret>
GOOGLE_REDIRECT_URL=https://<your-aws-apprunner-url>/api/v1/auth/google/callback
FRONTEND_URL=http://localhost:3000
```

**注意事項：**
- `GOOGLE_REDIRECT_URL` 必須與 Google Cloud Console 中的設定完全一致
- 部署到生產環境時，記得更新 `FRONTEND_URL`

## Google Cloud Console 設定

### 1. 更新已授權的重新導向 URI

前往：https://console.cloud.google.com/apis/credentials

點擊你的 OAuth 2.0 用戶端 ID，在「已授權的重新導向 URI」新增：

```
https://f82cb2me3v.ap-northeast-1.awsapprunner.com/api/v1/auth/google/callback
```

### 2. 本地開發（可選）

如果要在本地測試，也可以新增：

```
http://localhost:8080/api/v1/auth/google/callback
```

然後在本地執行時設定環境變數：

```bash
export GOOGLE_CLIENT_ID=<your-google-client-id>
export GOOGLE_CLIENT_SECRET=<your-google-client-secret>
export GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
export FRONTEND_URL=http://localhost:3000
```

### 3. 測試使用者

在開發模式下，只有「OAuth 同意畫面」中列出的測試使用者可以登入。

前往：https://console.cloud.google.com/apis/credentials/consent

在「測試使用者」區塊新增你的 Google 帳號。

## 登入流程

1. 用戶點擊「使用 Google 登入」
2. 前端呼叫 `GET /api/v1/auth/google/login` 取得 OAuth URL
3. 前端重定向到 Google 授權頁面
4. 用戶在 Google 授權
5. Google callback 到 `GET /api/v1/auth/google/callback`
6. 後端取得用戶資訊，建立/登入用戶
7. 後端生成 JWT，設定到 HTTP-only cookies
8. 後端重定向到 `${FRONTEND_URL}/?login=success`
9. 前端自動讀取 cookies 中的 JWT，取得用戶資訊

## 安全特性

- ✅ Google tokens 不經過前端瀏覽器
- ✅ JWT 儲存在 HTTP-only cookies，前端無法存取
- ✅ CSRF 保護（state parameter）
- ✅ 敏感資訊（Client Secret）只存在後端環境變數

## 疑難排解

### 錯誤：redirect_uri_mismatch

檢查：
1. `GOOGLE_REDIRECT_URL` 環境變數是否正確
2. Google Cloud Console 中的「已授權的重新導向 URI」是否包含此 URL
3. URL 是否完全一致（包括 http/https、port、path）

### 錯誤：invalid_state

CSRF 保護觸發，可能原因：
1. Cookie 被清除
2. 跨域問題
3. 超時（state cookie 有效期 10 分鐘）

### 部署後無法登入

確認：
1. AWS App Runner 環境變數已設定
2. Google Cloud Console 已新增生產環境的 redirect URI
3. CORS 設定允許前端域名
