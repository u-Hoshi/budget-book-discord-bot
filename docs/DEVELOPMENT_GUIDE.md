# Discord Bot 開発・運用ガイド

## 📋 プロジェクト概要

Discord上で画像（レシート等）をアップロードし、Dify AIを使用して自動的に内容を解析・処理するBotです。

## 🏗️ アーキテクチャ

```
Discord → Bot (Go) → Dify API → 処理結果 → Discord
                ↓
           ヘルスチェック
```

### 主要コンポーネント
- **Discord Bot**: メッセージ受信・レスポンス
- **HTTP Server**: ヘルスチェックエンドポイント
- **Image Processing**: 画像圧縮・最適化
- **Dify Integration**: AI処理ワークフロー
- **Health Check**: 継続稼働保証

## 🛠️ 開発環境

### 必要な環境変数
```bash
# Discord設定
APPLICATION_ID=your_discord_app_id
DISCORD_TOKEN=your_discord_bot_token

# Dify設定
DIFY_API_KEY=app-xxxxxxxxxxxxx
DIFY_ENDPOINT=https://api.dify.ai/v1
DIFY_INPUT_NAME=receipt_images

# サーバー設定
PORT=8080
HEALTH_CHECK_URL=https://your-app.koyeb.app

# 画像処理設定（オプション）
IMAGE_MAX_WIDTH=1500
IMAGE_QUALITY=85
ENABLE_COMPRESSION=true
```

### ローカル開発
```bash
# 依存関係インストール
go mod download

# 実行
go run main.go

# ビルド
go build -o main .
```

## 🚀 デプロイ

### Koyeb デプロイ手順

1. **GitHubリポジトリ準備**
   ```bash
   git add .
   git commit -m "Deploy to Koyeb"
   git push origin main
   ```

2. **Koyebでサービス作成**
   - GitHub連携
   - Build command: 自動検出
   - Run command: `./main`
   - Port: `8080`

3. **環境変数設定**
   - すべての必須環境変数を設定
   - `HEALTH_CHECK_URL`を忘れずに設定

4. **デプロイ実行**
   - 自動ビルド・デプロイ開始
   - ログで起動確認

### Docker デプロイ（手動）
```bash
# イメージビルド
docker build -t discord-bot .

# ローカル実行
docker run -p 8080:8080 --env-file .env discord-bot
```

## 🔍 トラブルシューティング

### よくある問題と解決法

#### 1. Koyeb Deep Sleep問題
**症状**: `No traffic detected in the past 300 seconds`
**解決**: [KOYEB_DEEP_SLEEP_ISSUE.md](./KOYEB_DEEP_SLEEP_ISSUE.md)参照

#### 2. 401認証エラー
**症状**: `Access token is invalid`
**解決**: [TROUBLESHOOTING_401.md](./TROUBLESHOOTING_401.md)参照

#### 3. 権限エラー
**症状**: `permission denied`
**解決**: 一時ディレクトリ使用に修正済み

#### 4. 重複処理
**症状**: 同じメッセージが2回処理される
**解決**: イベントハンドラー重複登録を修正済み

### ログ確認方法

#### Koyebログ
```bash
# リアルタイムログ
Koyeb Dashboard → Service → Logs

# 重要なログパターン
✅ Bot起動完了        # 正常起動
🔍 ヘルスチェック実行中  # Deep Sleep防止
❌ エラー表示        # 問題発生
```

#### ローカルログ
```bash
# 実行時ログ
go run main.go

# 出力例
🚀 Discord Bot 起動中...
✅ 必要な環境変数が設定されています。
🌐 HTTPサーバーを開始: ポート 8080
🕐 ヘルスチェックの定期実行を開始しました (5分間隔)
✅ Bot起動完了 - Ctrl+Cで終了
```

## 📊 監視・運用

### ヘルスチェック監視
- **エンドポイント**: `GET /`
- **期待レスポンス**: 
  ```json
  {
    "status": "ok",
    "timestamp": "2025-11-04T10:00:00Z",
    "version": "1.0.0",
    "uptime": "2h30m15s"
  }
  ```

### パフォーマンス指標
- **応答時間**: Discord APIレスポンス < 3秒
- **処理成功率**: > 95%
- **稼働率**: > 99%（Deep Sleep対策後）

### アラート設定
- ヘルスチェック失敗
- Discord API エラー率上昇
- Dify API エラー率上昇

## 🔒 セキュリティ

### 機密情報管理
- トークン・APIキーはログ出力しない
- 環境変数で管理
- `.env`ファイルは`.gitignore`に追加

### 権限設定
- Discord Bot: 最小限の権限のみ
- Container: 非rootユーザーで実行
- File System: 一時ディレクトリのみ使用

## 📈 今後の改善案

### 機能拡張
- [ ] 複数画像対応
- [ ] 結果の永続化
- [ ] ユーザー管理機能
- [ ] 統計・レポート機能

### パフォーマンス改善
- [ ] 画像処理の並列化
- [ ] キャッシュ機能
- [ ] レスポンス時間最適化

### 運用改善
- [ ] メトリクス収集
- [ ] 自動テスト追加
- [ ] CI/CD pipeline構築

---

**最終更新**: 2025年11月4日  
**担当者**: Development Team