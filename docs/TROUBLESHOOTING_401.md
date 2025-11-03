# 401エラー トラブルシューティングガイド

## 問題: Difyアップロード時に401エラーが発生する

### エラーメッセージ例
```
❌ [Dify Upload] アップロード失敗 - ステータス: 401
{"code":"unauthorized","message":"Access token is invalid","status":401}
```

## 原因と解決方法

### 1. 環境変数名の確認

**問題**: `.env`ファイルで`DIFY_API_URL`を使用していたが、コードは`DIFY_ENDPOINT`を読み込んでいた

**解決方法**:
- `.env`ファイルを確認し、以下の名前で統一してください：
  ```env
  DIFY_ENDPOINT=https://api.dify.ai/v1
  DIFY_API_KEY=app-xxxxxxxxxxxx
  ```
- または、両方とも対応するようにコードを修正済み（後方互換性あり）

### 2. API Keyの形式確認

**正しい形式**:
- プレフィックス: `app-` で始まる
- 長さ: 通常28文字以上
- 例: `app-hogehoge`

**確認方法**:
```bash
# .envファイルを確認
cat .env | grep DIFY_API_KEY
```

### 3. 空白文字の確認

**問題**: API Keyの前後に余分な空白やタブがある

**解決方法**:
- `.env`ファイルで以下を確認：
  ```env
  # ❌ 間違い（前後に空白がある）
  DIFY_API_KEY= app-xxxxxxxxxxxx 
  
  # ✅ 正しい（空白なし）
  DIFY_API_KEY=app-xxxxxxxxxxxx
  ```

### 4. API Keyの有効性確認

**確認手順**:
1. [Dify管理画面](https://cloud.dify.ai/)にログイン
2. 対象のアプリケーションを開く
3. 左メニューから「API Access」を選択
4. 「API Key」が有効であることを確認
5. 必要に応じて新しいAPI Keyを生成

### 5. エンドポイントの確認

**Dify Cloud使用時**:
```env
DIFY_ENDPOINT=https://api.dify.ai/v1
```

**セルフホスト版使用時**:
```env
DIFY_ENDPOINT=https://your-dify-domain.com/v1
```

## デバッグログの見方

Bot起動時に以下のようなログが出力されます：

```log
🚀 Discord Bot 起動中...
✅ .envファイルを読み込みました
📋 環境変数チェック:
  DIFY_API_KEY: app-vUxi2G... (長さ: 28)
  DIFY_ENDPOINT/DIFY_API_URL: https://api.dify.ai/v1
```

画像アップロード時:
```log
🔍 [Dify Upload] DIFY_API_KEY詳細:
  - 元の長さ: 28文字
  - 先頭10文字: app-vUxi2G...
  - プレフィックス: app- (正常)
  - 空白文字: 前=false, 後=false
```

### 異常なログの例

```log
# API Keyが短すぎる
- 元の長さ: 10文字  ← ⚠️ 短すぎる

# プレフィックスが間違っている
- プレフィックス: api- (想定外: app-で始まるべき)  ← ⚠️ 間違い

# 空白が含まれている
- 空白文字: 前=true, 後=false  ← ⚠️ 先頭に空白がある
```

## チェックリスト

- [ ] `.env`ファイルが存在する
- [ ] `DIFY_API_KEY`が`app-`で始まる
- [ ] API Keyの長さが28文字以上
- [ ] API Keyの前後に空白がない
- [ ] `DIFY_ENDPOINT`が正しく設定されている
- [ ] DifyのダッシュボードでAPI Keyが有効
- [ ] ネットワーク接続が正常
- [ ] Difyのサービスが稼働中

## まだ解決しない場合

1. **API Keyを再生成**
   - Dify管理画面でAPI Keyを削除
   - 新しいAPI Keyを生成
   - `.env`ファイルを更新
   - Botを再起動

2. **Dify APIの動作確認**
   ```bash
   curl -X POST https://api.dify.ai/v1/files/upload \
     -H "Authorization: Bearer app-xxxxxxxxxxxx" \
     -F "file=@test.jpg" \
     -F "user=test"
   ```

3. **ログを確認**
   - Bot起動時のログをすべて保存
   - 401エラー発生時の詳細ログを確認
   - GitHub Issuesに報告する際にログを添付

## 関連ドキュメント

- [Dify API ドキュメント](https://docs.dify.ai/guides/application-publishing/developing-with-apis)
- [Discord Bot README](./README.md)
