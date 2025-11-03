# 401エラー対策 - 実装内容まとめ

## 🎯 実装した機能

### 1. 起動時の環境変数詳細ログ

**実装箇所**: `main()`関数

```go
📋 環境変数チェック:
  APPLICATION_ID: 1423938471... (長さ: 19)
  DISCORD_TOKEN: MTQyMzkzOD... (長さ: 59)
  DIFY_API_KEY: app-vUxi2G... (長さ: 28)
  DIFY_ENDPOINT/DIFY_API_URL: https://api.dify.ai/v1
```

**機能**:
- 各環境変数の存在確認
- 値の長さチェック
- 機密情報のマスキング表示

### 2. API Keyの詳細検証

**実装箇所**: `uploadImageToDify()`関数

```go
🔍 [Dify Upload] DIFY_API_KEY詳細:
  - 元の長さ: 28文字
  - 先頭10文字: app-vUxi2G...
  - プレフィックス: app- (正常)
  - 空白文字: 前=false, 後=false
```

**検証項目**:
- ✅ API Keyの長さ
- ✅ プレフィックス（`app-`で始まるか）
- ✅ 前後の空白文字の有無
- ✅ 自動トリミング処理

### 3. Authorizationヘッダーの詳細ログ

**実装箇所**: `uploadImageToDify()`と`runDifyWorkflowWithImage()`

```go
🔍 [Dify Upload] ヘッダー詳細:
  - Content-Type: multipart/form-data; boundary=...
  - Authorization: Bearer app-vUxi2G...
  - Authorization長さ: 28文字 (Bearer含む: 35文字)
```

**確認内容**:
- 実際に送信されるAuthorizationヘッダー
- Bearer tokenの形式
- ヘッダー全体の長さ

### 4. 401エラー時の診断メッセージ

**実装箇所**: `uploadImageToDify()`のエラーハンドリング

```go
🔍 [Dify Upload] 401エラー診断:
  ❌ API Keyが無効です。以下を確認してください:
  1. .envファイルのDIFY_API_KEYが正しいか
  2. API Keyの先頭に 'app-' が付いているか
  3. API Keyの前後に余分な空白がないか
  4. Difyの管理画面でAPI Keyが有効か
  5. 使用しているエンドポイント(https://...)が正しいか
```

### 5. 環境変数名の後方互換性対応

**問題**: `.env`で`DIFY_API_URL`を使用していたが、コードは`DIFY_ENDPOINT`を読んでいた

**解決**: 両方に対応

```go
// DIFY_ENDPOINTとDIFY_API_URLの両方をサポート
difyEndpoint := os.Getenv("DIFY_ENDPOINT")
if difyEndpoint == "" {
    difyEndpoint = os.Getenv("DIFY_API_URL")
}
```

### 6. ヘルパー関数の追加

**追加した関数**:

1. `maskString(s string) string`
   - 機密情報を安全に表示

2. `getPrefixInfo(s string) string`
   - API Keyのプレフィックスを検証

3. `hasLeadingSpace(s string) bool`
   - 先頭の空白をチェック

4. `hasTrailingSpace(s string) bool`
   - 末尾の空白をチェック

## 📋 .envファイルの修正

### 変更前
```env
DIFY_API_URL=https://api.dify.ai/v1
DIFY_API_KEY=app-vUxi2Givn01MQgK3rZd51ytk
DIFY_FILE_UPLOAD_URL=https://api.dify.ai/v1/files/upload
```

### 変更後（推奨）
```env
DIFY_ENDPOINT=https://api.dify.ai/v1
DIFY_API_KEY=app-vUxi2Givn01MQgK3rZd51ytk
```

**注**: `DIFY_API_URL`も引き続きサポートされます（後方互換性）

## 🔍 デバッグ方法

### 1. Bot起動時のログを確認

```bash
cd /Users/hoshi/pg/discord-bot/1
go run main.go
```

以下の情報が表示されます:
- ✅ .envファイルの読み込み状況
- 📋 全環境変数の値（マスク済み）
- 🚀 Bot起動メッセージ

### 2. 画像アップロード時の詳細ログ

`!upload`コマンドで画像を送信すると、以下が表示されます:

```
📤 [STEP 6/8] Difyへのアップロード開始: image.jpg
  🔑 [Dify Upload] 環境変数チェック中...
  🔍 [Dify Upload] DIFY_API_KEY詳細: ...
  📂 [Dify Upload] ファイルオープン中: ...
  📦 [Dify Upload] multipart/form-data作成中...
  🌐 [Dify Upload] リクエスト作成中: POST https://...
  🔍 [Dify Upload] ヘッダー詳細: ...
  📤 [Dify Upload] HTTPリクエスト送信中...
  📥 [Dify Upload] レスポンス受信 - Status: ...
```

### 3. エラー時の対応

401エラーが発生した場合:
1. ログに表示される診断メッセージを確認
2. `docs/TROUBLESHOOTING_401.md`を参照
3. チェックリストに従って設定を確認

## 📚 作成したドキュメント

1. **TROUBLESHOOTING_401.md**
   - 401エラーの詳細な解決方法
   - チェックリスト
   - デバッグ手順
   - よくある問題と解決策

## 🚀 使用方法

1. **環境変数を正しく設定**
   ```bash
   # .envファイルを編集
   nano .env
   ```

2. **Botを起動**
   ```bash
   go run main.go
   ```

3. **ログを確認**
   - 起動時のログで環境変数が正しく読み込まれているか確認
   - API Keyの形式が正しいか確認

4. **画像をテスト**
   - Discordで画像を添付
   - `!upload`コマンドを送信
   - 詳細ログを確認

## ✅ 期待される動作

### 正常な場合
```
✅ [STEP 6/8 完了] Difyアップロード成功 - FileID: abc123
```

### API Key問題がある場合
```
❌ [Dify Upload] アップロード失敗 - ステータス: 401
🔍 [Dify Upload] 401エラー診断:
  ❌ API Keyが無効です。以下を確認してください:
  ...（診断メッセージ）
```

## 🔧 今後のメンテナンス

- ログレベルの調整が必要な場合は、各`log.Printf`を調整
- 新しいエラーパターンが見つかった場合は診断メッセージを追加
- TROUBLESHOOTING_401.mdを更新して知見を共有
