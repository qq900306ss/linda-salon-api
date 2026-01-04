# Go 版本固定設定

## 問題
本地的 Go linter 會自動將 go.mod 升級到較新版本（1.23/1.24），並加入 `toolchain` 指令，
但 **AWS App Runner 的 `runtime: go1` 只支援 Go 1.18.10**，導致部署失敗。

## 重要提醒
⚠️ **AWS App Runner Go 限制**：
- `runtime: go1` = Go 1.18.10（固定版本）
- 不支援 Go 1.19+
- 不支援 `toolchain` 指令
- 標準庫套件如 `slices`, `maps`, `cmp` 在 Go 1.21+ 才加入，1.18 沒有
- 建議使用 2022 年底發布的套件版本（Go 1.18 時代）

## 解決方案

### 方法 1: 設定環境變數（推薦）

在你的電腦設定環境變數：

**Windows (PowerShell):**
```powershell
[System.Environment]::SetEnvironmentVariable('GOTOOLCHAIN', 'local', 'User')
```

**Windows (命令提示字元):**
```cmd
setx GOTOOLCHAIN local
```

**Mac/Linux (在 ~/.bashrc 或 ~/.zshrc):**
```bash
export GOTOOLCHAIN=local
```

設定後重新啟動 VSCode。

### 方法 2: 每次 commit 前執行 fix script

```bash
./fix-go-mod.sh
git add go.mod
git commit -m "your message"
git push
```

### 方法 3: 使用 Git Hook (自動化)

建立 `.git/hooks/pre-commit` 檔案：

```bash
#!/bin/bash
# 自動修正 go.mod
sed -i '/^toolchain/d' go.mod
sed -i 's/^go [0-9]\+\.[0-9]\+.*/go 1.18/' go.mod
git add go.mod
```

給予執行權限：
```bash
chmod +x .git/hooks/pre-commit
```

## 驗證

執行以下命令驗證設定：
```bash
echo $GOTOOLCHAIN  # 應該顯示 "local"
```

## 相關檔案

- `.vscode/settings.json` - VSCode Go 擴充功能設定（停用自動格式化）
- `.go-version` - 指定 Go 版本為 1.18
- `go.env` - Go 環境變數設定
- `fix-go-mod.sh` - 手動修正 script
